<!-- TOC -->

- [1. 如何写出正确的netcat?](#1-如何写出正确的netcat)
- [2. 测试吞吐量](#2-测试吞吐量)

<!-- /TOC -->



<a id="markdown-1-如何写出正确的netcat" name="1-如何写出正确的netcat"></a>
# 1. 如何写出正确的netcat?

netcat要做到的事情
* [stdin] -> remote
* stdout <- [remote]

3.5个事件
* 连接建立
* 连接关闭
* 收到数据
* 数据成功拷贝到缓冲区

在本个应用中,因为tcp有半关闭的特性,所以有难度的是`连接关闭`的事件

正确关闭的方式应该是两种情况:

情况一(主动):  
`[stdin]` ->shutdown remote  
stdout <-shutdown `[remote]`  
close socket  


情况二(被动):  
stdout <-shutdown `[remote]`  
`[stdin]` ->shutdown remote  
close socket  


程序应该被设计成`两个并发单元`,并发单元结束后`按照关闭的两种情况去shutdown write`

注意 情况二被动接收到shutdown write时,没有好的办法`从stdin中唤醒`,所以`直接close`.

由上所述,程序只能说是`支持主动半关闭`,`不能支持被动半关闭`


<a id="markdown-2-测试吞吐量" name="2-测试吞吐量"></a>
# 2. 测试吞吐量

```bash
python3 netcat.py -l 50000 > /dev/null
dd if=/dev/zero bs=1MB count=4096 | python3 netcat.py localhost 50000

go run netcat.go -l 50000 > /dev/null
dd if=/dev/zero bs=1MB count=4096 | go run netcat.go localhost 50000
```