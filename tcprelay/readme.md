<!-- TOC -->

- [1. tcprelay](#1-tcprelay)
- [2. 测试半关闭](#2-测试半关闭)

<!-- /TOC -->


<a id="markdown-1-tcprelay" name="1-tcprelay"></a>
# 1. tcprelay

tcprelay要做到的事情
* local -> proxy -> remote
* local <- proxy <- remote

正确关闭的方式应该是(不分先后):

local ->shutdown proxy ->shutdown remote  
local shutdown<- proxy shutdown<- remote  
close  

<a id="markdown-2-测试半关闭" name="2-测试半关闭"></a>
# 2. 测试半关闭

```
go run netcat-half.go -l 20000
go run tcprelay.go 0.0.0.0 10000 localhost 20000
go run netcat-half.go localhost 10000

# 或
python3 tcprelay.py 0.0.0.0 10000 localhost 20000
```
