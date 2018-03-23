<!-- TOC -->

- [1. netcat](#1-netcat)
- [2. 测试吞吐量](#2-测试吞吐量)

<!-- /TOC -->



<a id="markdown-1-netcat" name="1-netcat"></a>
# 1. netcat

```
[stdin] -> remote
stdout <- [remote]
```

正确关闭的方式应该是两种情况:

情况一(主动):  
`[stdin]` ->shutdown remote  
stdout <-shutdown `[remote]`  
close socket  


情况二(被动):  
stdout <-shutdown `[remote]`  
`[stdin]` ->shutdown remote  
close socket  

情况二的问题是,要`等到`用户按下`ctrl+d`才能够完整的进行完`半关闭`

但是对于工具的使用者来说,我想达到`对方shutdown wr或者close了连接`,我方`看看是否还有要发送的消息`,如果没有,就`shutdown wr且close`.

但是没有好的方法可以做到`看看是否还有要发送的消息`,所以我直接os.close了

<a id="markdown-2-测试吞吐量" name="2-测试吞吐量"></a>
# 2. 测试吞吐量

```bash
python3 netcat.py -l 50000 > /dev/null
dd if=/dev/zero bs=1MB count=4096 | python3 netcat.py localhost 50000

go run netcat.go -l 50000 > /dev/null
dd if=/dev/zero bs=1MB count=4096 | go run netcat.go localhost 50000
```
