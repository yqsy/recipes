package main

import (
	"github.com/streadway/amqp"
	"log"
	"bytes"
	"time"
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

	q, err := ch.QueueDeclare(
		"task_queue",
		true, // 持久化
		false,
		false,
		false,
		nil,
	)

	err = ch.Qos(
		1,
		0,
		false,
	)

	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // 不自动ack
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dot_count := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dot_count)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false)
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C\n")
	<-forever
}
