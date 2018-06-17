package main

import (
	"net"
	"flag"
	"sync"
	"time"
	"fmt"
	"log"
	"code.cloudfoundry.org/bytefmt"
	"math/rand"
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

func (config *Config) IsReverseLocal() bool {
	if *config.readLocalReverse {
		return true
	} else {
		return false
	}
}

func (config *Config) IsReverseRemote() bool {
	if *config.readRemoteReverse {
		return true
	} else {
		return false
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
	waterMask.bytes += n
	waterMask.mtx.Unlock()
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

	// 反转1bit时所用
	rand *rand.Rand
}

func NewContext() *Context {
	ctx := &Context{}
	ctx.readLocalWaterMask = NewWaterMask()
	ctx.stopReadLocalCount = make(chan struct{}, 1)
	ctx.readRemoteWaterMask = NewWaterMask()
	ctx.stopReadRemoteCount = make(chan struct{}, 1)
	ctx.closeDone = make(chan struct{}, 2)
	ctx.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	return ctx
}

func AddMaskPerSecond(waterMask *WaterMask, limitPerSecond uint64, stopCount chan struct{}) {
	ticket := time.NewTicker(time.Second)

	quit := false

	for {
		select {
		case <-ticket.C:
			waterMask.AddMask(limitPerSecond)
		case <-stopCount:
			quit = true
			break
		}

		if quit {
			break
		}
	}
	log.Printf("ticker stop")
}

func Reserve1Bit(n byte, pos uint32) byte {
	if pos > 7 {
		return 0
	}
	m := byte(1 << pos)
	m &= n
	if m == 0 {
		n |= byte(1 << pos)
		return n
	} else {
		n &= ^(1 << pos)
		return n
	}
}

func Relay(config *Config, ctx *Context) {
	defer ctx.localConn.Close()

	var err error
	ctx.remoteConn, err = net.Dial("tcp", *config.remoteAddr)
	if err != nil {
		log.Printf("connect err: %v -> %v\n", ctx.localConn.RemoteAddr(), *config.remoteAddr)
		return
	}

	defer ctx.remoteConn.Close()

	log.Printf("relay: %v <-> %v\n", ctx.localConn.RemoteAddr(), *config.remoteAddr)

	if config.IsLocalReadlLimit() {
		go AddMaskPerSecond(ctx.readLocalWaterMask, config.readLocalLimit, ctx.stopReadLocalCount)
	}

	if config.IsRemoteReadlLimit() {
		go AddMaskPerSecond(ctx.readLocalWaterMask, config.readRemoteLimit, ctx.stopReadRemoteCount)
	}

	// localConn -> proxy -> remoteConn
	go func(config *Config, ctx *Context) {
		fixedBuffer := make([]byte, MaxReadBuffer)

		for {
			var buf []byte

			if config.IsLocalReadlLimit() {
				canReadBytes := ctx.readLocalWaterMask.WaitCanReadBytes()
				buf = fixedBuffer[:canReadBytes]
			} else {
				buf = fixedBuffer
			}

			rn, err := ctx.localConn.Read(buf)

			if err != nil {
				break
			}

			if config.IsReverseLocal() {
				bytePos := ctx.rand.Intn(rn)
				buf[bytePos] = Reserve1Bit(buf[bytePos], uint32(ctx.rand.Intn(8)))
			}

			wn, err := ctx.remoteConn.Write(buf[:rn])
			_ = wn
			if err != nil {
				break
			}
		}
		if config.IsLocalReadlLimit() {
			ctx.stopReadLocalCount <- struct{}{}
		}
		ctx.closeDone <- struct{}{}

		log.Printf("done: %v -> %v\n", ctx.localConn.RemoteAddr(), *config.remoteAddr)
	}(config, ctx)

	// localConn <- proxy <- remoteConn
	go func(config *Config, ctx *Context) {
		fixedBuffer := make([]byte, MaxReadBuffer)

		for {
			var buf []byte

			if config.IsRemoteReadlLimit() {
				canReadBytes := ctx.readRemoteWaterMask.WaitCanReadBytes()
				buf = fixedBuffer[:canReadBytes]
			} else {
				buf = fixedBuffer
			}

			rn, err := ctx.remoteConn.Read(buf)

			if err != nil {
				break
			}

			if config.IsReverseRemote() {
				bytePos := ctx.rand.Intn(rn)
				buf[bytePos] = Reserve1Bit(buf[bytePos], uint32(ctx.rand.Intn(8)))
			}

			wn, err := ctx.localConn.Write(buf[:rn])
			_ = wn
			if err != nil {
				break
			}
		}
		if config.IsRemoteReadlLimit() {
			ctx.stopReadRemoteCount <- struct{}{}
		}
		ctx.closeDone <- struct{}{}

		log.Printf("done: %v <- %v\n", ctx.localConn.RemoteAddr(), *config.remoteAddr)
	}(config, ctx)

	for i := 0; i < 2; i++ {
		<-ctx.closeDone
	}

	log.Printf("done: %v <-> %v\n", ctx.localConn.RemoteAddr(), *config.remoteAddr)
}

func ReadConfig(config *Config) bool {
	config.listenAddr = flag.String("ListenAddr", "", "local listen addr")
	config.remoteAddr = flag.String("RemoteAddr", "", "remote connect addr")
	config.readLocalReverse = flag.Bool("readLocalReverse", false, "random 1 bit Reverse")
	config.readRemoteReverse = flag.Bool("readRemoteReverse", false, "random 1 bit Reverse")
	readLocalLimitStr := flag.String("readLocalLimit", "", "N[T,G,M,K,B] per second (1024hex)")
	readRemoteLimitStr := flag.String("readRemoteLimit", "", "N[T,G,M,K,B] per second (1024hex)")
	flag.Parse()

	if *config.listenAddr == "" || *config.remoteAddr == "" {
		flag.Usage()
		return false
	}

	var err error
	if *readLocalLimitStr != "" {
		config.readLocalLimit, err = bytefmt.ToBytes(*readLocalLimitStr)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			flag.Usage()
			return false
		}
	}

	if *readRemoteLimitStr != "" {
		config.readRemoteLimit, err = bytefmt.ToBytes(*readRemoteLimitStr)
		if err != nil {
			fmt.Printf("err: %v\n", err)
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
		ctx := NewContext()

		var err error
		ctx.localConn, err = listener.Accept()
		if err != nil {
			continue
		}

		go Relay(config, ctx)
	}
}
