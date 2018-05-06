<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 实践](#2-实践)
- [3. 反射的思路(三个思考切入点)](#3-反射的思路三个思考切入点)
    - [3.1. 如何通过类型获得类型名称](#31-如何通过类型获得类型名称)
    - [3.2. 如何通过类型名称创建类型并反序列化](#32-如何通过类型名称创建类型并反序列化)
    - [3.3. 如何通过类型dispatch到不同的入口函数内](#33-如何通过类型dispatch到不同的入口函数内)

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

<a id="markdown-3-反射的思路三个思考切入点" name="3-反射的思路三个思考切入点"></a>
# 3. 反射的思路(三个思考切入点)

 
<a id="markdown-31-如何通过类型获得类型名称" name="31-如何通过类型获得类型名称"></a>
## 3.1. 如何通过类型获得类型名称

c++
```c++
// [类型的descriptor:typeName]
typedef codec::Query T;
std::string type_name = T::descriptor()->full_name();
```

go
```go
// [类型变量:typeName]
typeName := proto.MessageName(&codec.Query{})
fmt.Println(typeName)
```


<a id="markdown-32-如何通过类型名称创建类型并反序列化" name="32-如何通过类型名称创建类型并反序列化"></a>
## 3.2. 如何通过类型名称创建类型并反序列化

c++
```c++
// c++ 用到了prototype设计模式

// 通过hash的手段完成 typeName -> descriptor -> prototype -> New对象
// 存在的键值对为
// [Name:Descriptor] # DescriptorPool
// [Descriptor:Prototype] # MessageFactory

google::protobuf::Message *createMessage(const std::string &typeName) {
    google::protobuf::Message *message = nullptr;

    auto descriptor = google::protobuf::DescriptorPool::generated_pool()->FindMessageTypeByName(typeName);
    if (descriptor) {
        auto prototype = google::protobuf::MessageFactory::generated_factory()->GetPrototype(descriptor);
        if (prototype) {
            message = prototype->New();
        }
    }
    return message;
}
```

go 
```go
// go并没有使用设计模式,因为其可以将类型当做变量一样使用
// 建立映射 [typeName:类型变量]
// 直接可以从类型变量生成新的对象

func createMessage(typeName string) (interface{}, error) {
	mt := proto.MessageType(typeName)
	if mt == nil {
		fmt.Errorf("unknown message type %q", typeName)
	}

	return reflect.New(mt.Elem()).Interface(), nil
}


// 映射关系
func init() {
	proto.RegisterType((*Query)(nil), "codec.Query")
	proto.RegisterType((*Answer)(nil), "codec.Answer")
	proto.RegisterType((*Empty)(nil), "codec.Empty")
}
```


<a id="markdown-33-如何通过类型dispatch到不同的入口函数内" name="33-如何通过类型dispatch到不同的入口函数内"></a>
## 3.3. 如何通过类型dispatch到不同的入口函数内

第二个步骤之后c++得到`google::protobuf::Message *`,而go得到`interface{}`.然而这只是消息的一个`抽象`,如何知道它实际是什么?从而分发到不同的函数去处理?

c++
```c++
// 基类带有一个获取到Descriptor指针的函数
google::protobuf::Message::GetDescriptor()

// 那么就只要做一个Descriptor与函数的映射关系即可
std::unordered_map<const google::protobuf::Descriptor *, Callback *> callbacksMap;

// 回调函数
void fooQuery(int sockfd, codec::Query *query)

// 用模板保存类型
gb.callbacksMap[codec::Query::descriptor()] = new CallbackT<codec::Query>(fooQuery)
```

go
```go
// 在go里面,类型也是变量,所以直接用类型作为key,回调作为接口
type Callback func(message proto.Message)
callbacks := make(map[interface{}]Callback)
callbacks[reflect.TypeOf((*codec.Query)(nil))] = fooQuery

cb := callbacks[reflect.TypeOf(msg)]
    
cb(msg.(proto.Message))    


// 美中不足的是go没有泛型,在回调函数中自己向下吧!
// 或者像grpc一样自动生成模板代码
func fooQuery(message proto.Message) {
	query := message.(*codec.Query)
	_ = query
	fmt.Println("query")
}
```
