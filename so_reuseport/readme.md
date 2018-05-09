<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 测试方法](#2-测试方法)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明


* https://stackoverflow.com/questions/14388706

我梳理下来的观点:
1. 可以bind任意次相同的ip和端口
2. Linux >= 3.9

我自己的浅薄观点:  
```
像使用libevent的memcached主线程accept到了sockfd,再以mutex生产者消费者队列交给工作线程去serve,这个过程消耗多少时间?

参考:
https://github.com/brpc/brpc/blob/master/docs/cn/benchmark.md

>>但如果多个线程协作，即使在极其流畅的系统中，也要付出3-5微秒的上下文切换代价和1微秒的cache同步代价

把自己线程上的变量同步到另外一个线程是需要时间的,如何节省这个时间呢?

SO_REUSEPORT正是解决这个问题的.特别适合多进程模型,不用去做master accept交给slave去处理的逻辑了,每个人都是master去获取sockfd去处理,减少应用层逻辑量,避免cache同步的代价.

```

<a id="markdown-2-测试方法" name="2-测试方法"></a>
# 2. 测试方法

```
python3 so_reuseport.py 127.0.0.1 20001

for i in `seq 1 1000`; do  
  nc localhost 20001 > /dev/null 2>&1 &
done
```
