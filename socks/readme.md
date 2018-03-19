<!-- TOC -->

- [1. socks](#1-socks)
- [2. 其他](#2-其他)
- [3. 单元测试](#3-单元测试)
    - [3.1. socks4a](#31-socks4a)
    - [3.2. socks5](#32-socks5)
    - [3.3. 命令](#33-命令)

<!-- /TOC -->


<a id="markdown-1-socks" name="1-socks"></a>
# 1. socks
```
                         === raw === server1  
=== raw === socks server === raw === server2  
                         === raw === server3  
```

* https://www.openssh.com/txt/rfc1928.txt (socks5)
* https://www.openssh.com/txt/socks4.protocol (socks4)
* https://www.openssh.com/txt/socks4a.protocol (socks4a)
* https://en.wikipedia.org/wiki/SOCKS (wikipedia)


<a id="markdown-2-其他" name="2-其他"></a>
# 2. 其他
* https://golang.org/src/encoding/binary/binary_test.go (go byte包的测试用例)


简单bash单元测试
```bash
# socks4 -> 192.168.2.153:5003
echo -en '\x04\0x01\0x13\0x8B\0xC0\0xA8\0x2\0x99\0x0' | nc host1 20001

# socks4a -> vm1:5003
echo -en '\x04\0x01\0x13\0x8B\0x0\0x0\0x0\0xFF\0x0\0x76\0x6D\0x31\0x0' | nc host1 20001

# debian 
nc -X 4 -x 127.0.0.1:1080 127.0.0.1 5003
nc -X 5 -x 127.0.0.1:1080 127.0.0.1 5003

# centos(不支持5)
nc --proxy-type socks4 --proxy 127.0.0.1:1080 127.0.0.1 5003
```


<a id="markdown-3-单元测试" name="3-单元测试"></a>
# 3. 单元测试

<a id="markdown-31-socks4a" name="31-socks4a"></a>
## 3.1. socks4a

正确:
* socks4 正确性测试
* socks4a 正确性测试
* 包头一个一个字节发
* local half close
* remote half close


错误:
* Version不为4
* CommandCode不为1
* user ID string 长度攻击
* socks4a domain长度攻击

<a id="markdown-32-socks5" name="32-socks5"></a>
## 3.2. socks5

正确:
* ipv4
* ipv6 (暂时不测试,冒烟已通过)
* domain
* Greeting 多种认证方式(只支持No authentication)
* Greeting 和 heder一起过来了
* 包头一个一个字节发
* 头部是定长协议,如果超过协议长度的buffer过来,是否保存等连接到对方时再发送

错误:
* Version不为5
* CommandCode不为1
* AddressType不为 1 3 4

<a id="markdown-33-命令" name="33-命令"></a>
## 3.3. 命令

```bash
go run common.go socks4.go socks5.go socks.go :1080

go test common.go socks4.go socks5.go socks4_test.go 
go test common.go socks4.go socks5.go socks5_test.go 
```
