<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 简单冒烟测试](#2-简单冒烟测试)
- [3. 全面透明代理](#3-全面透明代理)

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

<a id="markdown-2-简单冒烟测试" name="2-简单冒烟测试"></a>
# 2. 简单冒烟测试


```bash
# 得到无污染的ip
dig @101.6.6.6  www.google.com

# 查询
sudo iptables -nv -L -t nat

# 增加一项
sudo iptables -t nat -N SOCKSs
sudo iptables -t nat -A SOCKS -d 172.217.161.164 -p tcp -j REDIRECT --to-ports 5001
sudo iptables -t nat -A SOCKS -j RETURN 

# 开启
sudo iptables -t nat -A OUTPUT -p tcp -j SOCKS
sudo iptables -t nat -A PREROUTING -p tcp -j SOCKS

# 开启程序
go run transparent.go :5001 127.0.0.1:1080
```

<a id="markdown-3-全面透明代理" name="3-全面透明代理"></a>
# 3. 全面透明代理

```bash

sudo apt-get install ipset -y
curl -sL http://f.ip.cn/rt/chnroutes.txt | egrep -v '^$|^#' > cidr_cn
sudo ipset -N cidr_cn hash:net
for i in `cat cidr_cn`; do echo ipset -A cidr_cn $i >> ipset.sh; done
chmod +x ipset.sh && sudo ./ipset.sh
sudo mkdir -p /etc/sysconfig
sudo ipset -S  | sudo tee /etc/sysconfig/ipset.cidr_cn
sudo touch /etc/rc.local
rm cidr_cn ipset.sh

sudo bash -c 'cat >> /etc/rc.local' << EOF
#!/bin/bash
# rc.local config file created by use
ipset restore < /etc/sysconfig/ipset.cidr_cn
exit 0
EOF

sudo iptables -t nat -N SOCKS
sudo iptables -t nat -A SOCKS -d 0.0.0.0/8 -j RETURN
sudo iptables -t nat -A SOCKS -d 10.0.0.0/8 -j RETURN
sudo iptables -t nat -A SOCKS -d 127.0.0.0/8 -j RETURN
sudo iptables -t nat -A SOCKS -d 169.254.0.0/16 -j RETURN
sudo iptables -t nat -A SOCKS -d 172.16.0.0/12 -j RETURN
sudo iptables -t nat -A SOCKS -d 192.168.0.0/16 -j RETURN
sudo iptables -t nat -A SOCKS -d 224.0.0.0/4 -j RETURN
sudo iptables -t nat -A SOCKS -d 240.0.0.0/4 -j RETURN
sudo iptables -t nat -A SOCKS -d 202.144.194.10 -j RETURN
sudo iptables -t nat -A SOCKS -m set --match-set cidr_cn dst -j RETURN
sudo iptables -t nat -A SOCKS -p tcp -j REDIRECT --to-ports 20018
sudo iptables -t nat -A OUTPUT -p tcp -j SOCKS
sudo iptables -t nat -A PREROUTING -p tcp -j SOCKS

```
