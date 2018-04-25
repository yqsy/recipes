<!-- TOC -->

- [1. tcprelay](#1-tcprelay)
- [命令](#命令)

<!-- /TOC -->


<a id="markdown-1-tcprelay" name="1-tcprelay"></a>
# 1. tcprelay

```
local -> proxy -> remote
local <- proxy <- remote
```

正确关闭的方式应该是(不分先后):

local ->shutdown proxy ->shutdown remote  
local shutdown<- proxy shutdown<- remote  
close  

<a id="markdown-命令" name="命令"></a>
# 命令

```bash
# local -> pfoxy ->remote 方向每秒限制1024B
go run tcprelay.go -ListenAddr :5000 -RemoteAddr localhost:5001 -readLocalLimit 1M

# local -> pfoxy ->remote 方向随机一字节,1bit反转
go run tcprelay.go -ListenAddr :5000 -RemoteAddr localhost:5001 -readLocalReverse

```
