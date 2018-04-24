package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"bufio"
)

func main() {
	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf("Usage:\n%v connectAddr\n", arg[0])
		return
	}

	config := &tls.Config{InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", arg[1], config)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	wn, err := conn.Write([]byte("hello\n"))
	_ = wn
	if err != nil {
		return
	}

	bufReader := bufio.NewReader(conn)

	line, err := bufReader.ReadString('\n')
	if err != nil {
		return
	}

	fmt.Print(line)
}
