package main

import (
	"net/http"
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"io"
)

// 查询留言
func handleMessageGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	env := os.Environ()
	env = append(env, fmt.Sprintf("REQUEST_METHOD=%v", r.Method))

	if r.Method == "GET" {
		env = append(env, fmt.Sprintf("QUERY_STRING=%v", r.URL.RawQuery))
	} else {
		fmt.Printf("unsupported method: %v\n", r.Method)
		return
	}

	cmd := exec.Command("./cgi")
	cmd.Env = env
	cmdOut, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	stdOut, _ := ioutil.ReadAll(cmdOut) // read until eof

	w.Write(stdOut)
}

// 留言
func handlerMessagePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	env := os.Environ()
	env = append(env, fmt.Sprintf("REQUEST_METHOD=%v", r.Method))

	if r.Method == "POST" {
		env = append(env, fmt.Sprintf("CONTENT_LENGTH=%v", r.ContentLength))
	} else {
		fmt.Printf("unsupported method: %v\n", r.Method)
		return
	}

	cmd := exec.Command("./cgi")
	cmd.Env = env
	cmdIn, _ := cmd.StdinPipe()
	cmdOut, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	if r.Method == "POST" {
		io.Copy(cmdIn, r.Body)
	}

	stdOut, _ := ioutil.ReadAll(cmdOut) // read until eof
	w.Write(stdOut)
}

func main() {
	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :8080\n", arg[0], arg[0])
		return
	}

	// 按照restful的接口,应该是同一个名词,不同的动词
	// 但是我估计原声的框架不支持同一个URL不同method的分发
	// 所以这里作为一个example就不讲究了
	http.HandleFunc("/messagepost", handlerMessagePost)
	http.HandleFunc("/messageget", handleMessageGet)

	err := http.ListenAndServe(arg[1], nil)

	if err != nil {
		panic(err)
	}
}
