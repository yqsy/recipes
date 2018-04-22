<!-- TOC -->

- [1. multiplexer dmux](#1-multiplexer-dmux)
    - [1.1. `ssh -NL`功能切入的思考点如下:](#11-ssh--nl功能切入的思考点如下)
    - [1.2. `ssh -NR`功能切入的思考点如下:](#12-ssh--nr功能切入的思考点如下)
- [2. 包头定义](#2-包头定义)

<!-- /TOC -->


<a id="markdown-1-multiplexer-dmux" name="1-multiplexer-dmux"></a>
# 1. multiplexer dmux

```
session ======>|-------------|          |------|======> session
session ======>| multiplexer |=========>| dmux |======> session
session ======>|-------------|          |------|======> session
```

这个项目比简单的proxy增加了一些难度(multiplexer维护一个客户端连接/dmux维护一组连接)


![](multiplexer.png)

<a id="markdown-11-ssh--nl功能切入的思考点如下" name="11-ssh--nl功能切入的思考点如下"></a>
## 1.1. `ssh -NL`功能切入的思考点如下:  

multiplexer:
* 发送`CONNECT ip:port`给channel握手,握手成功后本地开始监听accept session
* session发送`SYN`包装给channel
* 互相传递`FIN`,`payload`

与之相对的dmux:
* 返回`CONNECT OK`,并保存需要连接的`ip:port`
* `SYN` ->返回`SYN OK`或`SYN ERROR`
* 互相传递`FIN`,`payload`

<a id="markdown-12-ssh--nr功能切入的思考点如下" name="12-ssh--nr功能切入的思考点如下"></a>
## 1.2. `ssh -NR`功能切入的思考点如下:  

multiplexer:
* 发送`BIND ip:port`给channel握手,并保存需要连接的`ip:port`
* `SYN` ->返回`SYN OK`或`SYN ERROR`
* 互相传递`FIN`,`payload`

与之相对的dmux:
* BIND相应地址,并返回`BIND OK / BIND ERROR`
* session发送`SYN`包装给channel
* 互相传递`FIN`,`payload`

<a id="markdown-2-包头定义" name="2-包头定义"></a>
# 2. 包头定义

* len 4 byte
* id 4 byte
* cmd 1 byte (bool)
* if cmd == true, command

---

建立channel握手:
* CONNECT ip:port
* CONNECT OK
* BIND ip:port
* BIND OK / BIND ERROR
---

其他:
* FIN
* SYN
* SYN OK / SYN ERROR
* ACK xxx // xxx is bytes num
