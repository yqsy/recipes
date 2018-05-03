package main

import (
	"net/http"
	"encoding/base64"
	"fmt"
	"crypto/tls"
	"io"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func main() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", "https://localhost:20001/123/abc", nil)

	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", "Basic "+basicAuth("admin", "123456"))

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	buf := make([]byte, resp.ContentLength)
	io.ReadFull(resp.Body, buf)
	fmt.Println(string(buf))
}
