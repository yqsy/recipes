package main

import (
	"github.com/streadway/amqp"
	"os"
	"strings"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"task_queue",
		true, // 持久化
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	body := bodyFrom(os.Args)

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // 持久化
			ContentType:  "text/plain",
			Body:         []byte(body),
		})

	if err != nil {
		panic(err)
	}
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}
