package main

import (
	"net/http"
	"os"
	"fmt"
	"net"
	"net/http/fcgi"
)

var usage = `Usage:
%v listenAddr
`

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("This is a FastCGI example server.\n"))

	fmt.Println("an req")
}

func main() {
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

	b := new(FastCGIServer)
	err = fcgi.Serve(listener, b)
	if err != nil {
		panic(err)
	}
}
