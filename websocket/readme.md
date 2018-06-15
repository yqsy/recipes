<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 我的发现猜测](#2-我的发现猜测)

<!-- /TOC -->



<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://github.com/gorilla/websocket


在研究dht krpc的时候,突然发现udp是一个应用层全双工的协议, A -> B, A <- B 双向都可以同时发送请求

一般使用tcp的应用层(rpc)协议都是半双工的,`一请求一应答`.

```
A -> B
A <- B
```

半双工思考点:
* A处encode,发送,阻塞等待应答到来,`明确decode`应答类型
* B处使用原型模式,根据type字符串decode req类型再dispatch


但是全双工的协议就不是上面两个思考点那么简单了

```
(1)
A -> B  
A <- B

与

(2)
A <- B
A -> B

可能同时发生
```

思考点:
* (1) 如何在收到B的数据时,知道这是对应的A发送的请求的应答(可能这是B的请求的呢?)
* (2) 与上面的一样,如何收到A的数据时,知道这是对应的......

只要能够对应好,区分这是1.对方发送过来的应答(找到相应的上下文,然后`明确decode`) 2. 对方发送过来的请求(直接处理然后给应答)


<a id="markdown-2-我的发现猜测" name="2-我的发现猜测"></a>
# 2. 我的发现猜测

看了下go实现的websocket example目录,基本都是`消息推送`,消息推过去就不管应答了,所以它应该不能解决我的问题!

下次研究websocket.