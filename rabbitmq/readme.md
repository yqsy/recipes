<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 简单的示例](#2-简单的示例)
- [3. 消息的可靠性](#3-消息的可靠性)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://www.rabbitmq.com/tutorials/tutorial-one-go.html
* https://www.rabbitmq.com/getstarted.html
* https://www.rabbitmq.com/install-debian.html

```bash
sudo apt-get update
sudo apt-get install erlang -y

echo "deb https://dl.bintray.com/rabbitmq/debian stretch main" | sudo tee /etc/apt/sources.list.d/bintray.rabbitmq.list


wget -O- https://dl.bintray.com/rabbitmq/Keys/rabbitmq-release-signing-key.asc |
     sudo apt-key add -

sudo apt-get update
sudo apt-get install rabbitmq-server

sudo rabbitmq-server
```

<a id="markdown-2-简单的示例" name="2-简单的示例"></a>
# 2. 简单的示例

sender
```golang
// 和中间件取得联系
conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

// 打开一个channel
ch, err := conn.Channel()

// channel参数配置
q, err := ch.QueueDeclare(
	"hello",
	false, // durable
	false, // delete when unused
	false, // exclusive
	false, // no-wait
	nil,)  // arguments

// 发布消息
err = ch.Publish(
	"",     // exchange
	q.Name, // routing key
	false,  // mandatory
	false,  // immediate
	amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	})

```

receiver
```golang
// 和中间件取得联系
conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

// 打开一个channel
ch, err := conn.Channel()

// channel 参数配置
q, err := ch.QueueDeclare(
	"hello",
	false, // durable
	false, // delete when unused
	false, // exclusive
	false, // no-wait
	nil,)  // arguments


// 接收消息
msgs, err := ch.Consume(
	q.Name, // queue
	"",     // consumer
	true,   // auto-ack
	false,  // exclusive
	false,  // no-local
	false,  // no-wait
	nil,)   // args
```


<a id="markdown-3-消息的可靠性" name="3-消息的可靠性"></a>
# 3. 消息的可靠性

sender增加
```golang
q, err := ch.QueueDeclare(
    true // durable
)

amqp.Publishing{
		DeliveryMode: amqp.Persistent, // 持久化
}

```

receiver增加
```golang
q, err := ch.QueueDeclare(
    true // durable
)

err = ch.Qos(
	1,     // prefetch count
	0,     // prefetch size
	false, // global
)

msgs, err := ch.Consume(
    false, // auto-ack
)

// 主动ack
d.Ack(false)
```
