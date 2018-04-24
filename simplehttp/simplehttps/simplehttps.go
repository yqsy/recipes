package main

import (
	"os"
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", r.URL.Path[1:])
}

func main() {

	arg := os.Args

	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :8080\n", arg[0], arg[0])
		return
	}

	http.HandleFunc("/", handler)
	err := http.ListenAndServeTLS(arg[1], "fd.crt", "fd.key", nil)

	if err != nil {
		panic(err)
	}
}
