<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 模式](#2-模式)

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

go get github.com/streadway/amqp
```


<a id="markdown-2-模式" name="2-模式"></a>
# 2. 模式

对官网的示例做一个简单的功能说明:
* hello world 简单的示例
* work queues 任务`分`给`多个`worker,消息持久,应用层ack
* publish/subscribe 消息发布给`每个`sub
* Routing 过滤消息级别
* Topics 有选择的接收消息
* rpc 字面意思

