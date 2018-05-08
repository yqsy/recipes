
<!-- TOC -->

- [1. 说明](#1-说明)

<!-- /TOC -->



<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://stackoverflow.com/questions/14388706

我梳理下来几个观点:
1. 没有SO_REUSEADDR, BIND A `0.0.0.0:21`,BIND B `192.168.0.1:21`会出错
2. 处于TIME_WAIT的socket的BIND
