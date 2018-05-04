package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"bufio"
	"log"
	"github.com/yqsy/recipes/verification/simple_tls_authentication/common"
	"encoding/gob"
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

	// verify identify
	loginReq := common.LoginReq{User: "admin", Password: "12345"}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(loginReq)

	decoder := gob.NewDecoder(conn)
	var loginRes common.LoginRes
	err = decoder.Decode(&loginRes)
	if err != nil {
		log.Printf("err: %v", err)
		return
	}

	if !loginRes.LoginResult {
		log.Printf("login error\n")
		return
	}

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
