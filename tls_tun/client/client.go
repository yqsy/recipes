package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"bufio"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
		log.Println(err)
		return
	}

	bufReader := bufio.NewReader(conn)

	line, err := bufReader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return
	}

	log.Print(line)
}
