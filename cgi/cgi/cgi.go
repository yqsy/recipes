package main

import (
	"os"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"strconv"
	"io"
	"encoding/json"
	"time"
	"fmt"
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

	dispatch(method, logger)
}

func dispatch(method string, logger *log.Logger) {

	if method == "GET" {

		fmt.Printf(`<meta charset="utf-8">
`)

		queryStr := os.Getenv("QUERY_STRING")
		logger.Printf("QUERY_STRING: %v\n", queryStr)

		db, err := sql.Open("sqlite3", "./cgi.db")
		if err != nil {
			fmt.Printf("<p>open sqlite3 error: %v</p>\n", err)
			return
		}

		defer db.Close()

		rows, err := db.Query("select title,body,create_time from message order by create_time desc")
		if err != nil {
			fmt.Printf("<p>db.Query error: %v</p>\n", err)
			return
		}

		defer rows.Close()

		var result []string

		for rows.Next() {
			var title string
			var body string
			var createTime int
			err = rows.Scan(&title, &body, &createTime)

			if err != nil {
				fmt.Printf("<p>rows.Scan: %v</p>\n", err)
				return
			}
			result = append(result, fmt.Sprintf("title:%v body:%v time:%v", title, body, time.Unix(int64(createTime), 0).Format(time.RFC822)))
		}

		for _, val := range result {
			fmt.Printf("<p>%v</p>\n", val)
		}

	} else if method == "POST" {
		fmt.Printf(`<meta charset="utf-8">
`)

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

		tx, err := db.Begin()
		if err != nil {
			fmt.Printf("<p>db.Begin error: %v</p>\n", err)
			return
		}

		stmt, err := tx.Prepare("insert into message(title,body, create_time) values(?,?,?)")
		if err != nil {
			fmt.Printf("<p>tx.Prepare error: %v</p>\n", err)
			return
		}

		defer stmt.Close()
		_, err = stmt.Exec(postMessage.Title, postMessage.Body, time.Now().Unix())
		if err != nil {
			fmt.Printf("<p>stmt.Exec error: %v</p>\n", err)
			return
		}

		err = tx.Commit()

		if err != nil {
			fmt.Printf("<p>insert error: %v</p>\n", err)
			return
		}

		fmt.Printf("<p>insert ok</p>\n")
	} else {
		fmt.Printf("<p>unsoppered</p>\n")
	}
}
