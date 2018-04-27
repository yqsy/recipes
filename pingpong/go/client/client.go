package main

import (
	"net"
	"os"
	"fmt"
)

var usage = `Usage:
%v connectAddr`

func serve(conn net.Conn) {
	defer conn.Close()

}

func main() {
	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf(usage, arg[0])
		return
	}

	conn, err := net.Dial("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	serve(conn)
}
