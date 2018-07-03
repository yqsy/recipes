<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 维度整理](#2-维度整理)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://github.com/go-redis/redis (连接的库)
* https://github.com/gomodule/redigo (!)
* https://godoc.org/github.com/gomodule/redigo/redis#hdr-Executing_Commands (document)

```bash
docker run --name redis1 \
    -p 6379:6379 \
    -d redis:4.0.10

```



<a id="markdown-2-维度整理" name="2-维度整理"></a>
# 2. 维度整理

* string 字符串
* hash 哈希
* list 列表
* set 集合
* zset 有序集合
