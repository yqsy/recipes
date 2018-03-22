package main

import (
	"net"
	"fmt"
)

func main() {

	remoteConn, err := net.Dial("tcp", "vm1:5003")

	if err != nil {
		panic(err)
	}

	bytes := make([]byte, 20 * 1024 * 1024)

	wn,err := remoteConn.Write(bytes)

	fmt.Println(wn)

	if err != nil {
		panic(err)
	}

}
