<!-- TOC -->

- [1. 说明](#1-说明)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

参考
* https://en.wikipedia.org/wiki/Proxy_server#Transparent_proxy
* https://gist.github.com/wen-long/8644243 (ss-redir)
* http://f.ip.cn/rt/chnroutes.txt (国内ip地址)
* https://github.com/ryanchapman/go-any-proxy/blob/master/any_proxy.go (看下获取ip地址的源码吧)

本透明代理其实不是标准意义上的`透明代理`即`不修改包头,包体`

其描述的是一种`特殊`的代理方式: 在路由器上用`iptables`重定向到`透明代理程序`,提取ip包头的ip,tcp包头的port,再包装好socks包以`加密`方式发送到socks服务器

比之程序层面支持socks代理来说,其优点是`程序不需要直接支持socks`,缺点是`dns需要本地解析`.

其唯一的缺点有幸的是国内有纯净的dns,省去了再转发dns查询的烦恼(udp丢包率高时效果懂得)

