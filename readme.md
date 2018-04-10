<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 网络部分](#2-网络部分)
    - [2.1. proxy类型](#21-proxy类型)
    - [2.2. echo 类型](#22-echo-类型)
    - [2.3. 其他](#23-其他)
- [3. 底层部分](#3-底层部分)

<!-- /TOC -->

<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

一些关于网络/并发的演示代码

<a id="markdown-2-网络部分" name="2-网络部分"></a>
# 2. 网络部分

<a id="markdown-21-proxy类型" name="21-proxy类型"></a>
## 2.1. proxy类型

* [http(s)proxy](httpproxy/readme.md)
* [socks(4|4a|5)proxy](socks/readme.md)
* [tcprelay](tcprelay/readme.md)
* [netcat](netcat/readme.md)
* [multiplexer(N:1 and 1:n)](multiplexer/readme.md)
* [through-net(中继内网穿透)](through-net/readme.md)
* [slowrecv(慢速收)](slowrecv/readme.md)

<a id="markdown-22-echo-类型" name="22-echo-类型"></a>
## 2.2. echo 类型

* [echo](echo)
* [discard](discard)
* [ttcp](ttcp/readme.md)
* [simplehttp](simplehttp)


<a id="markdown-23-其他" name="23-其他"></a>
## 2.3. 其他

* [idleconnection(时间轮/堆/升序链表/多线程做法)](idleconnection/readme.md)


<a id="markdown-3-底层部分" name="3-底层部分"></a>
# 3. 底层部分

* [自旋锁/读写锁/信号量/互斥量/条件变量/lock-free/wait-free](sync/readme.md)
