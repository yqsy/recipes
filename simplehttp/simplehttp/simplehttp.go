package main

import (
	"net/http"
	"fmt"
	"os"
	"net/http/httputil"
)

func handler(w http.ResponseWriter, r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(requestDump))
}

func main() {

	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :8080\n", arg[0], arg[0])
		return
	}

	http.HandleFunc("/", handler)
	err := http.ListenAndServe(arg[1], nil)

	if err != nil {
		panic(err)
	}
}
