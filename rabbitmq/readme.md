<!-- TOC -->

- [1. 说明](#1-说明)

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

