package main

import (
	"net"
	"flag"
	"sync"
	"time"
	"fmt"
	"log"
	"code.cloudfoundry.org/bytefmt"
)

const (
	MaxReadBuffer = 32 * 1024
)

type Config struct {
	listenAddr        *string
	remoteAddr        *string
	readLocalLimit    uint64
	readRemoteLimit   uint64
	readLocalReverse  *bool
	readRemoteReverse *bool
}

func (config *Config) IsLocalReadlLimit() bool {
	if config.readLocalLimit == 0 {
		return false
	} else {
		return true
	}
}

func (config *Config) IsRemoteReadlLimit() bool {
	if config.readRemoteLimit == 0 {
		return false
	} else {
		return true
	}
}

// 限制流速功能所用,每秒都会上升,表示可读取多少字节
type WaterMask struct {
	bytes uint64
	cond  *sync.Cond
	mtx   sync.Mutex
}

// 增加水位,增加可读字节数目
func (waterMask *WaterMask) AddMask(n uint64) {
	waterMask.mtx.Lock()
	defer waterMask.mtx.Unlock()
	waterMask.bytes += n
	waterMask.cond.Signal()
}

// 可以读取到多少字节
func (waterMask *WaterMask) WaitCanReadBytes() (rtn uint64) {
	waterMask.mtx.Lock()
	defer waterMask.mtx.Unlock()

	for !(waterMask.bytes > 0) {
		waterMask.cond.Wait()
	}

	if waterMask.bytes > MaxReadBuffer {
		rtn = uint64(MaxReadBuffer)
		waterMask.bytes -= MaxReadBuffer
	} else {
		rtn = waterMask.bytes
		waterMask.bytes = 0
	}

	return rtn
}

func NewWaterMask() *WaterMask {
	waterMask := &WaterMask{}
	waterMask.cond = sync.NewCond(&waterMask.mtx)
	return waterMask
}

type Context struct {
	readLocalWaterMask  *WaterMask
	stopReadLocalCount  chan struct{}
	readRemoteWaterMask *WaterMask
	stopReadRemoteCount chan struct{}

	localConn  net.Conn
	remoteConn net.Conn
	// localConn -> proxy -> remoteConn
	// localConn <- proxy <- remoteConn
	closeDone chan struct{}
}

func NewContext() *Context {
	context := &Context{}
	context.readLocalWaterMask = NewWaterMask()
	context.stopReadLocalCount = make(chan struct{}, 1)
	context.readRemoteWaterMask = NewWaterMask()
	context.stopReadRemoteCount = make(chan struct{}, 1)
	context.closeDone = make(chan struct{}, 2)
	return context
}

func AddMaskPerSecond(waterMask *WaterMask, limitPerSecond uint64, stopCount chan struct{}) {
	ticket := time.NewTicker(time.Second)

	for {
		select {
		case <-ticket.C:
			waterMask.AddMask(limitPerSecond)
		case <-stopCount:
			break
		}
	}
	log.Printf("ticker stop")
}
func Relay(config *Config, context *Context) {
	defer context.localConn.Close()

	var err error
	context.remoteConn, err = net.Dial("tcp", *config.remoteAddr)
	if err != nil {
		log.Printf("connect err: %v -> %v\n", context.localConn.RemoteAddr(), *config.remoteAddr)
		return
	}

	defer context.remoteConn.Close()

	log.Printf("relay: %v <-> %v\n", context.localConn.RemoteAddr(), *config.remoteAddr)

	if config.IsLocalReadlLimit() {
		go AddMaskPerSecond(context.readLocalWaterMask, config.readLocalLimit, context.stopReadLocalCount)
	}

	if config.IsRemoteReadlLimit() {
		go AddMaskPerSecond(context.readLocalWaterMask, config.readRemoteLimit, context.stopReadRemoteCount)
	}

	// localConn -> proxy -> remoteConn
	go func(config *Config, context *Context) {
		fixedBuffer := make([]byte, MaxReadBuffer)

		for {
			var buf []byte

			if config.IsLocalReadlLimit() {
				canReadBytes := context.readLocalWaterMask.WaitCanReadBytes()
				buf = make([]byte, canReadBytes)
			} else {
				buf = fixedBuffer
			}

			rn, err := context.localConn.Read(buf)

			if err != nil {
				break
			}

			wn, err := context.remoteConn.Write(buf[:rn])
			_ = wn
			if err != nil {
				break
			}
		}
		if config.IsLocalReadlLimit() {
			context.stopReadLocalCount <- struct{}{}
		}
		context.closeDone <- struct{}{}

		log.Printf("done: %v -> %v\n", context.localConn.RemoteAddr(), *config.remoteAddr)
	}(config, context)

	// localConn <- proxy <- remoteConn
	go func(config *Config, context *Context) {
		fixedBuffer := make([]byte, MaxReadBuffer)

		for {
			var buf []byte

			if config.IsRemoteReadlLimit() {
				canReadBytes := context.readRemoteWaterMask.WaitCanReadBytes()
				buf = make([]byte, canReadBytes)
			} else {
				buf = fixedBuffer
			}

			rn, err := context.remoteConn.Read(buf)

			if err != nil {
				break
			}

			wn, err := context.localConn.Write(buf[:rn])
			_ = wn
			if err != nil {
				break
			}
		}
		if config.IsRemoteReadlLimit() {
			context.stopReadRemoteCount <- struct{}{}
		}
		context.closeDone <- struct{}{}

		log.Printf("done: %v <- %v\n", context.localConn.RemoteAddr(), *config.remoteAddr)
	}(config, context)

	for i := 0; i < 2; i++ {
		<-context.closeDone
	}

	log.Printf("done: %v <-> %v\n", context.localConn.RemoteAddr(), *config.remoteAddr)
}

func ReadConfig(config *Config) bool {
	config.listenAddr = flag.String("ListenAddr", "", "local listen addr")
	config.remoteAddr = flag.String("RemoteAddr", "", "remote connect addr")
	config.readLocalReverse = flag.Bool("readLocalReverse", false, "random 1 bit Reverse")
	config.readRemoteReverse = flag.Bool("readRemoteReverse", false, "random 1 bit Reverse")
	readLocalLimitStr := flag.String("readLocalLimit", "", "N[T,G,M,K,B] per second")
	readRemoteLimitStr := flag.String("readRemoteLimit", "", "N[T,G,M,K,B] per second")
	flag.Parse()

	if *config.listenAddr == "" || *config.remoteAddr == "" {
		flag.Usage()
		return false
	}

	var err error
	if *readLocalLimitStr != "" {
		config.readLocalLimit, err = bytefmt.ToBytes(*readLocalLimitStr)
		if err != nil {
			fmt.Printf("err: %v", err)
			flag.Usage()
			return false
		}
	}

	if *readRemoteLimitStr != "" {
		config.readRemoteLimit, err = bytefmt.ToBytes(*readRemoteLimitStr)
		if err != nil {
			fmt.Printf("err: %v", err)
			flag.Usage()
			return false
		}
	}

	return true
}

func main() {
	config := &Config{}

	if !ReadConfig(config) {
		return
	}

	listener, err := net.Listen("tcp", *config.listenAddr)
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	for {
		context := NewContext()

		var err error
		context.localConn, err = listener.Accept()
		if err != nil {
			continue
		}

		go Relay(config, context)
	}
}
