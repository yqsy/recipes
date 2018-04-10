<!-- TOC -->

- [1. multiplexer dmux](#1-multiplexer-dmux)
- [2. 流量控制](#2-流量控制)
- [3. 包头定义](#3-包头定义)
- [4. 单元测试](#4-单元测试)
- [5. 项目需要改进的地方](#5-项目需要改进的地方)

<!-- /TOC -->


<a id="markdown-1-multiplexer-dmux" name="1-multiplexer-dmux"></a>
# 1. multiplexer dmux

```
input1 ======>|-------------|                   |------|======> output1
input1 ======>| multiplexer |====> channel ====>| dmux |======> output2
input1 ======>|-------------|                   |------|======> output3
```

这个项目比简单的proxy增加了一些难度,所切入的思考点如下:


multiplexer关于连接相关思考点:  
* input方 `SYN`,`FIN` 连接, `包装,穿透给`channel
* channel方 `FIN` 连接, 辨别是哪个input方, 发送shutdown write

multiplexer关于数据相关思考点:  
* input方 `send` payload , send channel payload 
* channel方 `send` payload , 辨别是哪个input方, send input payload

---

dmux关于连接相关思考点(对应上方):  
* channel方 `SYN`,`FIN`, 连接/shut down write指定的output
* output方 `FIN`, `包装,穿透给`channel

dmux关于数据相关思考点(对应上方):  
* channel方 `send`, send到指定的output
* output方 `send`, send channel

---

channel连接本身的思考:  
* multiplexer <- dmux 收到`FIN`, 关闭所有的input, `重连`
* multiplexer -> dmux 收到`FIN`, 关闭所有的output,  继续`accept` 新连接 处理


<a id="markdown-2-流量控制" name="2-流量控制"></a>
# 2. 流量控制

* 从input或者output收数据之前判断`发送字节数`是否达到`高水位`,如果达到`等待条件变量`直至水位之下
* 向output或者input发送数据完毕之后,判断`发送字节数`是否达到`consume位置`,如果达到`向channel`发送consume消息降低水位

<a id="markdown-3-包头定义" name="3-包头定义"></a>
# 3. 包头定义

* len 4 byte
* id 4 byte
* cmd 1 byte (bool)
* if cmd == true, command\r\n

---
command
* SYN ip:port\r\n
* FIN\r\n


<a id="markdown-4-单元测试" name="4-单元测试"></a>
# 4. 单元测试

* 同时开10个连接,每个连接收/发1G数据(并行)(1.正确完整收发 2.正确关闭连接 3.正确回收id) --> 需要肉眼配合观察日志,或者程序提供查询接口(比较麻烦)
* 包头长度攻击
* cmd长度攻击
* 错误cmd攻击
* 错误的id

<a id="markdown-5-项目需要改进的地方" name="5-项目需要改进的地方"></a>
# 5. 项目需要改进的地方

* 单通道支持多连接
* 单通道双向连接
* 心跳包支持
* 类似jmx的控制面板观察错误

命令
```bash
cd common
go test common.go common_test.go
```
