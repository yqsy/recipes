package main

import (
	"net/http"
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"io"
	"net/http/httputil"
)

// GET POST
func handleMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	requestDump, err := httputil.DumpRequest(r, true)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(requestDump))

	env := os.Environ()
	env = append(env, fmt.Sprintf("REQUEST_METHOD=%v", r.Method))

	if r.Method == "GET" {
		env = append(env, fmt.Sprintf("QUERY_STRING=%v", r.URL.RawQuery))
	} else if r.Method == "POST" {
		env = append(env, fmt.Sprintf("CONTENT_LENGTH=%v", r.ContentLength))
	} else {
		fmt.Printf("unsupported method: %v\n", r.Method)
		return
	}

	cmd := exec.Command("./cgi")
	cmd.Env = env
	cmdIn, _ := cmd.StdinPipe()
	cmdOut, _ := cmd.StdoutPipe()

	err = cmd.Start()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	if r.Method == "POST" {
		io.Copy(cmdIn, r.Body)
	}

	stdOut, _ := ioutil.ReadAll(cmdOut) // read until eof

	w.Header().Set("Content-Type", "text/html")
	w.Write(stdOut)
}

func main() {
	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :8080\n", arg[0], arg[0])
		return
	}

	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/message", handleMessage)

	err := http.ListenAndServe(arg[1], nil)

	if err != nil {
		panic(err)
	}
}
