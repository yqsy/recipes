<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 网络部分](#2-网络部分)
    - [2.1. proxy类型](#21-proxy类型)
    - [2.2. echo 类型](#22-echo-类型)
    - [2.3. chat类型](#23-chat类型)
    - [2.4. 其他](#24-其他)
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
* [tcprelay(代理/慢速/1bit反转)](tcprelay/readme.md)
* [netcat](netcat/readme.md)
* [slowrecv(多路汇聚慢速收)](slowrecv/readme.md)
* [multiplexer(N:1 and 1:n)](multiplexer/readme.md)
* [through-net(中继内网穿透)](multiplexer/readme.md)
* [transparent-proxy(透明代理)]()
* [tls-tun]()

<a id="markdown-22-echo-类型" name="22-echo-类型"></a>
## 2.2. echo 类型

* [pingpong](pingpong)
* [echo](echo)
* [chargen]()
* [time]()
* [daytime]()
* [discard](discard)
* [ttcp](ttcp/readme.md)
* [simplehttp](simplehttp)

<a id="markdown-23-chat类型" name="23-chat类型"></a>
## 2.3. chat类型

* [hub]()
* [pub-sub]()

<a id="markdown-24-其他" name="24-其他"></a>
## 2.4. 其他

* [incomplete-send(tcp发送数据不完整例子)]()
* [file-transfer(文件传输)]()
* [connection-pool(连接池)]()
* [identity-verification(通道身份验证)]()
* [maxconnection(最大连接数)]()
* [heart-beat(心跳包)]()
* [idleconnection(空闲连接时间轮/堆/升序链表/多线程做法)](idleconnection/readme.md)
* [roundtrip(测时间差)](roundtrip)
* [simple-rpc]()
* [sudoku(数独)]()
* [procmon(简易监控)]()
* [N皇后]()
* [多机求解中位数]()

<a id="markdown-3-底层部分" name="3-底层部分"></a>
# 3. 底层部分

* [自旋锁/读写锁/信号量/互斥量/条件变量](sync/readme.md)
* [lock-free/wait-free](sync/readme.md)
* [blockqueue](blockqueue)
* [mcoroutine(协程实现)]()
* [mtcp(可靠传输协议实现)]()
