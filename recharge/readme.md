<!-- TOC -->

- [1. 说明](#1-说明)

<!-- /TOC -->



<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://www.zhihu.com/question/35552800/answer/140722893


业务场景: 

服务A提供入金接口(假设单纯只是资金表++):

对外问题
* 用户点击入金按钮,请求堵塞在服务端接收缓冲区,客户端重发?`重复入金`

对内问题
* rpc调用超时,重复调用?`重复入金`
* mq补偿,重复调用?`重复入金`


可见不管对内还是对外,都有重复入金的问题.原因是什么?`分布式系统的本身不可靠`,永远无法知道超时是对方没有收到,还是收到了

提供接口:

* 注册 (邮箱/手机号 保证唯一性)
* 入金 (rest/rpc/mq)  (都必须是调用方生成调用流水)


```bash

# 生成go代码
protoc -I /usr/local/include -I recharge_protocol \
    -I $GOPATH/src \
    -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --go_out=plugins=grpc:recharge_protocol \
    recharge_protocol/recharge_protocol.proto

# 生成reverse-proxy
protoc -I /usr/local/include -I recharge_protocol \
  -I $GOPATH/src \
  -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=logtostderr=true:recharge_protocol \
   recharge_protocol/recharge_protocol.proto 


```
