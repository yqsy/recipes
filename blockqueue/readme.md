<!-- TOC -->

- [1. 命令](#1-命令)
- [2. 简单性能报告](#2-简单性能报告)

<!-- /TOC -->

<a id="markdown-1-命令" name="1-命令"></a>
# 1. 命令



```bash
cd cplusplus_mutex
cmake -DCMAKE_BUILD_TYPE=RELEASE
make

cd cplusplus_spinlock
cmake -DCMAKE_BUILD_TYPE=RELEASE
make
```

<a id="markdown-2-简单性能报告" name="2-简单性能报告"></a>
# 2. 简单性能报告

|go blockqueue|go chan|c++ mutex|boost lock-free queue|
|:-:|:-:|:-:|:-:|
|350W|300W|500W|600W|

