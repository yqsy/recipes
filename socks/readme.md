<!-- TOC -->

- [1. socks](#1-socks)
- [2. 其他](#2-其他)
- [3. 单元测试](#3-单元测试)

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
echo -e '\x04\0x01\0x13\0x8B\0xC0\0xA8\0x2\0x99\0x0' | nc host1 20001

# socks4a -> vm1:5003
echo -e '\x04\0x01\0x13\0x8B\0x0\0x0\0x0\0xFF\0x0\0x76\0x6D\0x31\0x0' | nc host1 20001

# debian 
nc -X 4 -x 127.0.0.1:1080 127.0.0.1 5003

#centos
nc --proxy-type socks4 --proxy 127.0.0.1:1080 127.0.0.1 5003
```


<a id="markdown-3-单元测试" name="3-单元测试"></a>
# 3. 单元测试

* socks4 正确性测试
* socks4a 正确性测试
* 第一个字节不为4
* 第二个字节不为1
* 包头8字节一个一个字节发
* user ID string 长度攻击
* user ID 一个一个字节发 (验证复杂度和正确性)
* socks4a domain长度攻击
* socks4a domain 一个一个字节发 (验证复杂度和正确性)
* 只开连接不发包攻击,用户层心跳包验证正确性
* remote half close
* local half close
