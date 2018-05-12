package main

import (
	"os"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"strconv"
	"io"
	"encoding/json"
	"time"
)

type PostMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func main() {
	file, _ := os.Create("/tmp/log1")

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	method := os.Getenv("REQUEST_METHOD")
	logger.Printf("REQUEST_METHOD: %v\n", method)

	fmt.Printf(`Content-type:text/html

<html>
<head>
<meta charset="utf-8">
<title></title>
</head>
<body>
`)

	dispatch(method, logger)

	fmt.Printf(`</body>
</html>
`)

}

func dispatch(method string, logger *log.Logger) {
	if method == "GET" {
		queryStr := os.Getenv("QUERY_STRING")
		logger.Printf("QUERY_STRING: %v\n", queryStr)

		db, err := sql.Open("sqlite3", "./cgi.db")
		if err != nil {
			fmt.Printf("<p>open sqlite3 error: %v</p>\n", err)
			return
		}

		defer db.Close()



	} else if method == "POST" {
		contentLengthStr := os.Getenv("CONTENT_LENGTH")
		logger.Printf("contentLength: %v\n", contentLengthStr)

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			fmt.Printf("<p>CONTENT_LENGTH: %v error</p>\n", contentLength)
			return
		}

		buf := make([]byte, contentLength)
		rn, err := io.ReadFull(os.Stdin, buf)
		if err != nil || rn != len(buf) {
			fmt.Printf("<p>cgi read error: %v</p>\n", err)
			return
		}

		var postMessage PostMessage

		err = json.Unmarshal(buf, &postMessage)
		if err != nil {
			fmt.Printf("<p>json content error: %v</p>\n", err)
			return
		}

		if postMessage.Title == "" {
			fmt.Printf("<p>title is empty</p>\n")
			return
		}

		db, err := sql.Open("sqlite3", "./cgi.db")
		if err != nil {
			fmt.Printf("<p>open sqlite3 error: %v</p>\n", err)
			return
		}

		defer db.Close()

		tx, _ := db.Begin()
		stmt, _ := tx.Prepare("insert into message(title,body, create_time) values(?,?,?)")
		defer stmt.Close()
		_, err = stmt.Exec(postMessage.Title, postMessage.Body, time.Now().Unix())
		tx.Commit()

		if err != nil {
			fmt.Printf("<p>insert error: %v</p>\n", err)
			return
		}

		fmt.Printf("<p>insert ok</p>\n")
	}
}
