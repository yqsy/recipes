<!-- TOC -->

- [1. ttcp](#1-ttcp)
- [2. 安全问题](#2-安全问题)

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

<a id="markdown-2-安全问题" name="2-安全问题"></a>
# 2. 安全问题
fileld 1: number 4 bytes  
fileld 2: length 4 bytes  

决定了服务端进行`多少次循环`与`分配多少内存`,如果放到公网上使用,需要检验范围