<!-- TOC -->

- [1. slowrecv](#1-slowrecv)
- [指令](#指令)

<!-- /TOC -->


<a id="markdown-1-slowrecv" name="1-slowrecv"></a>
# 1. slowrecv

```
    input ===>|-----------|     |----|     |-----------|
==> input ===>| slowrecv  | ==> | pv | ==> | /dev/zero |
    input ===>|-----------|     |----|     |-----------|
```

作用: 将所有流量汇聚到一个小管道,以测试input那端是否有`流量控制`


<a id="markdown-指令" name="指令"></a>
# 指令

```
go run slowrecv.go -l 5001 | pv -L 1K | cat > /dev/null
```
