<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 网络部分](#2-网络部分)
    - [2.1. proxy](#21-proxy)
    - [2.2. echo](#22-echo)
    - [2.3. benchmark](#23-benchmark)
    - [2.4. chat](#24-chat)
    - [2.5. 其他](#25-其他)
    - [p2p](#p2p)
    - [2.6. 测试框架](#26-测试框架)
    - [2.7. 分布式系统](#27-分布式系统)
    - [2.8. 特性](#28-特性)
- [3. 底层部分](#3-底层部分)

<!-- /TOC -->

<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

一些关于网络/并发的演示代码

<a id="markdown-2-网络部分" name="2-网络部分"></a>
# 2. 网络部分

<a id="markdown-21-proxy" name="21-proxy"></a>
## 2.1. proxy

* [socks(4|4a|5)proxy](socks)
* [http(s)proxy](httpproxy)
* [proxy_all(proxy集合)](proxy_all)
* [tcprelay(代理/慢速/1bit反转)](tcprelay)
* [netcat](netcat)
* [slowrecv(多路汇聚慢速收)](slowrecv)
* [multiplexer(N:1 and 1:n)](multiplexer)
* [through_net(中继内网穿透)](multiplexer)
* [tls_tun/tls_netcat](tls_tun)
* [transparent_proxy(透明代理)](transparent_proxy)

<a id="markdown-22-echo" name="22-echo"></a>
## 2.2. echo

* [echo](echo)
* [discard](discard)
* [chargen](chargen)
* [time](time)
* [daytime](daytime)
* [ttcp](ttcp)
* [simplehttp(s)](simplehttp)
* [simpleudp](simpleudp)

<a id="markdown-23-benchmark" name="23-benchmark"></a>
## 2.3. benchmark

* [pingpong](pingpong)
* [httpbench](http_bench)
* [zeromq(延迟)](zeromq)

<a id="markdown-24-chat" name="24-chat"></a>
## 2.4. chat

* [hub_pub_sub](hub)
* [chat(聊天室)](chat)

<a id="markdown-25-其他" name="25-其他"></a>
## 2.5. 其他

* [incomplete_send(tcp发送数据不完整例子)](incomplete_send)
* [identity_verification(通道身份验证)](verification)
* [maxconnection(最大连接数)](maxconnection)
* [codec(编解码器)](codec)
* heart_beat(心跳包)
* connection_pool(连接池)
* [idleconnection(空闲连接时间轮/堆/升序链表/多线程做法)](idleconnection)
* [roundtrip(测时间差)](roundtrip)
* simple_rpc
* [sudoku负载均衡](sudoku)
* [procmon(简易监控)](procmon)
* [cgi留言板](cgi)
* [fastcgi](fastcgi)
* [websocket](websocket)
* [mariadb](mariadb)
* [redis](redis)


<a id="markdown-p2p" name="p2p"></a>
## p2p

* [torrentparse(种子文件解析器)](torrentparse)
* [dht(嗅探)](dht)
* p2pdownload(p2p下载器)(p2pdownload)

<a id="markdown-26-测试框架" name="26-测试框架"></a>
## 2.6. 测试框架

* [grpc_sudoku](grpc/sudoku)
* [thrift_sudoku](thrift/sudoku)
* [rabbitmq](rabbitmq)
* [grpc_rest](grpc_rest)

<a id="markdown-27-分布式系统" name="27-分布式系统"></a>
## 2.7. 分布式系统

* [recharge](recharge)
* [mimeralpool](mineralpool)

<a id="markdown-28-特性" name="28-特性"></a>
## 2.8. 特性

* [SO_REUSEADDR/TIME_WAIT](so_reuseaddr)
* [SO_REUSEPORT](so_reuseport)

<a id="markdown-3-底层部分" name="3-底层部分"></a>
# 3. 底层部分

* [自旋锁/互斥量/条件变量](sync)
* [lock_free/wait_free](sync)
* [blockqueue](blockqueue)
* mcoroutine(协程实现)
* mtcp(可靠传输协议实现)
