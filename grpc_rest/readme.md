<!-- TOC -->

- [1. 说明](#1-说明)

<!-- /TOC -->



<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://github.com/grpc-ecosystem/grpc-gateway

```
# 安装protoc见wiz (版本不能低于3.0)

go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get -u github.com/golang/protobuf/protoc-gen-go

```

```bash
# 生成go代码
protoc -I /usr/local/include -I service_protocol \
  -I $GOPATH/src \
  -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --go_out=plugins=grpc:service_protocol \
    service_protocol/service_protocol.proto


# 生成reverse-proxy
protoc -I /usr/local/include -I service_protocol \
  -I $GOPATH/src \
  -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=logtostderr=true:service_protocol \
   service_protocol/service_protocol.proto


# 启动
cd /home/yq/go/src/github.com/yqsy/recipes/grpc_rest/server
go run main.go :20001

# 开启proxy(默认:8080)
cd /home/yq/go/src/github.com/yqsy/recipes/grpc_rest/gw
go run main.go -echo_endpoint :20001


# swagger ? 
protoc -I /usr/local/include -I service_protocol \
  -I $GOPATH/src \
  -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=logtostderr=true:service_protocol \
  service_protocol/service_protocol.proto

```
