package main

import (
	"os"
	"fmt"
	"net"
	"sync"
	"github.com/yqsy/recipes/hub/common"
	"bufio"
	"log"
)

var usage = `Usage:
%v listenAddr
`

const (
	MaxReadSize = 16384
)

// 可做pub,亦可做sub
type Context struct {
	conn net.Conn

	// sub or pub
	way string

	// pub 数量
	pubNum int
}

type Subs struct {
	// hash set
	array map[*Context]struct{}

	mtx sync.Mutex
}

func (subs *Subs) addSub(sub *Context) {
	subs.mtx.Lock()
	defer subs.mtx.Unlock()
	subs.array[sub] = struct{}{}
}

func (subs *Subs) removeSub(sub *Context) {
	subs.mtx.Lock()
	defer subs.mtx.Unlock()
	delete(subs.array, sub)
}

func (subs *Subs) pubMsg(msg *common.Msg) {
	subs.mtx.Lock()
	defer subs.mtx.Unlock()

	for sub, _ := range subs.array {

		wn, err := sub.conn.Write(msg.Serialize())
		_ = wn
		if err != nil {
			// 肯定已经关闭了
			delete(subs.array, sub)
			log.Printf("write error force unsub %v\n", sub.conn.RemoteAddr())
		}
	}
}

type Global struct {
	// key: topic value: subsMap slice
	subsMap map[string]*Subs

	// 保护map本身
	mtx sync.Mutex
}

func (gb *Global) addSub(topic string, sub *Context) {
	gb.mtx.Lock()
	defer gb.mtx.Unlock()

	if subs, ok := gb.subsMap[topic]; ok {
		subs.addSub(sub)
	} else {
		gb.subsMap[topic] = NewSubs(sub)
	}
}

func (gb *Global) removeSub(topic string, sub *Context) {
	gb.mtx.Lock()
	defer gb.mtx.Unlock()
	if subs, ok := gb.subsMap[topic]; ok {
		subs.removeSub(sub)
	}
}

func (gb *Global) pubMsg(msg *common.Msg) {
	gb.mtx.Lock()
	if subs, ok := gb.subsMap[msg.Topic]; !ok {
		gb.mtx.Unlock()
		return
	} else {
		gb.mtx.Unlock()
		subs.pubMsg(msg)
	}
}

func NewSubs(sub *Context) *Subs {
	subs := &Subs{}
	subs.array = make(map[*Context]struct{})
	subs.array[sub] = struct{}{}
	return subs
}

func serve(ctx *Context, gb *Global) {
	defer ctx.conn.Close()

	bufReader := bufio.NewReaderSize(ctx.conn, MaxReadSize)

	for {
		msg, err := common.ReadMsg(bufReader)
		if err != nil {
			log.Printf("read msg error: %v\n", err)
			break
		}

		if msg.Cmd == "sub" {
			gb.addSub(msg.Topic, ctx)
			ctx.way = "sub"
			log.Printf("sub %v %v\n", msg.Topic, ctx.conn.RemoteAddr())
		} else if msg.Cmd == "unsub" {
			gb.removeSub(msg.Topic, ctx)
			log.Printf("unsub %v %v\n", msg.Topic, ctx.conn.RemoteAddr())
		} else if msg.Cmd == "pub" {
			gb.pubMsg(msg)
			ctx.way = "pub"
			ctx.pubNum += 1
		}
	}

	if ctx.way == "pub" {
		log.Printf("pub exit pubNum: %v %v\n", ctx.pubNum, ctx.conn.RemoteAddr())
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	gb := &Global{subsMap: make(map[string]*Subs)}

	for {
		ctx := &Context{}
		ctx.conn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		go serve(ctx, gb)
	}
}
