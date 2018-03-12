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
* https://www.scadacore.com/tools/programming-calculators/online-hex-converter/ (大小端转换)
* https://golang.org/src/encoding/binary/binary_test.go (go byte包的测试用例)


```bash
# 发送二进制数据测试
echo -e '\x04\0x01\0x13\0x8B\0xC0\0xA8\0x2\0x99\0x0' | nc host1 20001
```


<a id="markdown-3-单元测试" name="3-单元测试"></a>
# 3. 单元测试
