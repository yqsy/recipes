package main

import (
	"os"
	"fmt"
	"net"
	"log"
	"sync"
	"github.com/yqsy/recipes/chat/common"
)

var usage = `Usage:
%v listenAddr
`

type Context struct {
	conn net.Conn
}

type Global struct {
	// hash set
	clients map[*Context]struct{}

	mtx sync.Mutex
}

func (gb *Global) addClientToRome(client *Context) {
	gb.mtx.Lock()
	defer gb.mtx.Unlock()
	gb.clients[client] = struct{}{}
}

func (gb *Global) removeClientInRome(client *Context) {
	gb.mtx.Lock()
	defer gb.mtx.Unlock()
	delete(gb.clients, client)
}

func (gb *Global) broadCastMsg(ctx *Context, msg *common.Message) {
	gb.mtx.Lock()
	defer gb.mtx.Unlock()

	for client, _ := range gb.clients {

		if client == ctx {
			continue
		}

		err := common.WriteToConn(client.conn, msg)

		if err != nil {
			// 肯定已经关闭了
			delete(gb.clients, client)
			log.Printf("write error force delete %v\n", client.conn.RemoteAddr())
		}
	}
}

func serve(ctx *Context, gb *Global) {
	defer ctx.conn.Close()

	log.Printf("new client %v\n", ctx.conn.RemoteAddr())

	gb.addClientToRome(ctx)
	defer gb.removeClientInRome(ctx)

	for {
		msg, err := common.ReadFromConn(ctx.conn)
		if err != nil {
			break
		}

		gb.broadCastMsg(ctx, msg)
	}

	log.Printf("client exit %v\n", ctx.conn.RemoteAddr())
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

	gb := &Global{clients: make(map[*Context]struct{})}

	for {
		ctx := &Context{}
		ctx.conn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		go serve(ctx, gb)

	}
}
