<!-- TOC -->

- [1. ttcp](#1-ttcp)
- [2. 使用](#2-使用)
- [3. 安全问题](#3-安全问题)

<!-- /TOC -->


<a id="markdown-1-ttcp" name="1-ttcp"></a>
# 1. ttcp


protocol:

client -> server:  
fileld 1: number 4 bytes  
fileld 2: length 4 bytes  
fileld 3: payload [length(4 bytes) + body(length bytes), ...]  

server -> client:  
field 1: ack 4 bytes  

<a id="markdown-2-使用" name="2-使用"></a>
# 2. 使用

```bash
# 默认监听5003
go run ttcp.go server

# 自定义payload大小(默认65535)
go run ttcp.go client -l 1048576

# 自定义payload数目(默认8192)
go run ttcp.go client -n 16384
```


<a id="markdown-3-安全问题" name="3-安全问题"></a>
# 3. 安全问题
fileld 1: number 4 bytes  
fileld 2: length 4 bytes  

决定了服务端进行`多少次循环`与`分配多少内存`,如果放到公网上使用,需要检验范围

