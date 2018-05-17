<!-- TOC -->

- [1. 选择](#1-选择)
- [2. 测量](#2-测量)
- [3. 测量维度](#3-测量维度)

<!-- /TOC -->


<a id="markdown-1-选择" name="1-选择"></a>
# 1. 选择

陈硕的实例有5种模型

模型|特点|优缺点
-|-|-
server_basic.cc|单eventloop|发挥单核性能
server_threadpool.cc|单eventloop,线程池|发挥多核性能,不过延时会增加一些
server_multiloop.cc|多eventloop|发挥多核性能,延时不增加
server_hybrid.cc|动态eventloop,线程池|处理带宽单event loop应付不过来的情况
server_prod.cc|过载保护|1. 队列太大,返回server too busy 2.发送缓冲区水位

这5种模型我统统都不实验^ - ^. 我的选择是用go来做实验.

主要有几点考虑:
* 并发模型的不同,不需要去思考`(单/多)`event loop 以及线程池
* 不需要去思考过载保护,因为数据积压在了内核缓冲区

大大的减少了心智负担.

<a id="markdown-2-测量" name="2-测量"></a>
# 2. 测量


陈硕的测量方式有5种

文件|说明
-|-
batch.cc|直接求解
batch.cc|通过网络求解,得到网络占比
pipeline.cc|测量最大容量,流水线深度多少时把服务器cpu占满
sudoku_stress.cc|压力测试
loadtest.cc|性能测试,延迟,最小,最大,平均,中位数, 延迟分布


<a id="markdown-3-测量维度" name="3-测量维度"></a>
# 3. 测量维度

延迟维度

描述|变量名
-|-
延迟累加|latency_sum_us
60秒内每秒延迟|latency_sum_us_per_second 
60秒内总计延迟|latency_sum_us_60s



请求维度

描述|变量名
-|-
请求累加|total_requests/total_responses/total_solved/bad_requests/dropped_requests
60秒内每秒请求|requests_per_second
60秒内总计请求|requests_60s
平均延迟|latency_us_60s_avg / latency_us_avg


延迟维度2

* min
* max
* sum
* average
* median
* p90
* p99
