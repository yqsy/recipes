
<!-- TOC -->

- [1. 说明](#1-说明)

<!-- /TOC -->



<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://stackoverflow.com/questions/14388706

我梳理下来几个观点:
1. 没有SO_REUSEADDR, BIND A `0.0.0.0:21`,BIND B `192.168.0.1:21`会出错
2. 处于TIME_WAIT的socket的BIND会出错

我的实践
1. 不管怎么样BIND相同端口,任何地址都会和`0.0.0.0`冲突
2. 我的内核版本Linux 4.14.0-deepin2-amd64,主动关闭不会出现TIME_WAIT额,需要找份TCP代码梳理一下

