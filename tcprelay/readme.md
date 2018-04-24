<!-- TOC -->

- [1. tcprelay](#1-tcprelay)

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

