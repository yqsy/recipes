<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 实践](#2-实践)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

技术选型很关键.兼容性,通用性是选型的重要指标,如果选择了只能给一门语言使用的某种技术,那么后继发展将会相当痛苦.

例如使用c struct或gob作为rpc框架的协议.结果虽然不是不能使用任何其他语言了,但是使用其他语言的成本会很高.

所以这个示例使用性能和通用性兼备的\protobuf. 做一个codec的演示.思考点:  
* 长度,把包截成流
* 类型,将二进制数据反序列化成对象
* check sum
* type name\0 结束,方便tcpdump观察
* 没有版本号

对于`类型`,有一些错误的做法:
* 在header中放int typeid,接收放用switch-case来选择对应的消息类型和处理函数.容易产生id的分配冲突.
* 在header中放string typeName,然后look-up table.缺点是每次都要修改table,其实protobuf自带了解决方案


<a id="markdown-2-实践" name="2-实践"></a>
# 2. 实践

* https://github.com/google/protobuf/tree/master/examples (示例目录)
* https://github.com/google/protobuf/blob/master/src/README.md (编译protobuf)
* https://developers.google.com/protocol-buffers/docs/downloads (二进制)
* https://developers.google.com/protocol-buffers/docs/gotutorial (go tutorial)
* https://developers.google.com/protocol-buffers/docs/proto3 (proto3 的说明)

```bash
cd /media/yq/ST1000DM003/linux/reference/refer
wget https://github.com/google/protobuf/archive/v3.5.1.tar.gz
tar -xvzf v3.5.1.tar.gz
cd protobuf-3.5.1
./autogen.sh
./configure
make
make check
sudo make install
sudo ldconfig 

go get -u github.com/golang/protobuf/protoc-gen-go

# protoc -I=$SRC_DIR --go_out=$DST_DIR $SRC_DIR/addressbook.proto

# 生成go的代码
protoc -I=./ --go_out=./go/ ./proto/query.proto

# 生成c艹的代码
protoc -I=./ --cpp_out=./cplusplus/ ./proto/query.proto
```
