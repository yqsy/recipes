<!-- TOC -->

- [1. multiplexer dmux](#1-multiplexer-dmux)
- [2. 包头定义](#2-包头定义)
- [3. 单元测试](#3-单元测试)

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




<a id="markdown-2-包头定义" name="2-包头定义"></a>
# 2. 包头定义

* len 4 byte
* id 4 byte
* cmd 1 byte (bool)
* if cmd == true, command\r\n

---
command
* SYN to ip:port\r\n
* FIN\r\n


<a id="markdown-3-单元测试" name="3-单元测试"></a>
# 3. 单元测试

```bash
go test multiplexer.go multiplexer_test.go
```
