<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 我梳理的概要](#2-我梳理的概要)
- [3. find_node](#3-find_node)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://en.wikipedia.org/wiki/Mainline_DHT
* https://github.com/zrools/spider/tree/gh-pages/asyncDHT
* https://github.com/fanpei91/nodeDHT/blob/master/nodeDHT.js (nodejs)
* https://github.com/laomayi/simDHT (python)
* https://github.com/shiyanhui/dht (go)
* https://blog.csdn.net/xxxxxx91116/article/details/7970815 (csdn)

BEP协议
* http://www.bittorrent.org/beps/bep_0003.html (BT协议)
* http://www.bittorrent.org/beps/bep_0005.html (dhc)
* http://www.bittorrent.org/beps/bep_0009.html (hashinfo -> metadata)
* http://www.bittorrent.org/beps/bep_0010.html 

不错的文档:
* https://www.jianshu.com/p/5c8e1ef0e0c3 (文档1)
* https://www.jianshu.com/p/8229e6be1e23 (文档2)
* http://www.aneasystone.com/archives/2015/05/analyze-magnet-protocol-using-wireshark.html (抓包分析)
* http://www.aneasystone.com/archives/2015/05/how-does-magnet-link-work.html
* https://juejin.im/entry/576ecaa4d342d30057c88408

成品:
* http://www.baocaibt.org/
* http://bthub.io/

工具:
* https://tool.lu/torrent
* https://github.com/ngosang/trackerslist 

bt缓存站:
* http://storetorrents.com/hash/



<a id="markdown-2-我梳理的概要" name="2-我梳理的概要"></a>
# 2. 我梳理的概要

先对协议进行梳理吧:

* ping: 字面含义作用
* find_node: 寻找`最近`的node
* get_peers: 根据`info_hash`获取`peers`  
* announce_peer: 通知其他节点自己开始下载某个资源,用于构建peer列表


dht爬虫思路:
* 不断地find_node,加入到其他node的路由表中.让其他节点发送`announce_peer`给自己,获得`info_hash`
* 两种方式从info_hash到metadata ~~1. 去种子库获取~~ 2. bep_009

下载思路:
* 传统: .torrent文件 -> 固定的tracker -> peers  (peers提供文件下载)
* 分布式: hash值(磁力链接) -> get_peers 

---

* get_peers info_hash在该网络中不一定存在
* announce_peer info_hash在该网路中一定存在

<a id="markdown-3-find_node" name="3-find_node"></a>
# 3. find_node

python和go的example有些差异,两边各自都有些问题

语言|join|衍生|事务id
-|-|-|-
go|id:自身 target:自身|id:衍生地址+自身混合 target:衍生地址|全局递增的id取2字节
python|id:自身 target:随机|id:衍生地址+自身混合 target:随机|随机2字节



go

逻辑不正确之一:
id自身 target自身?下一节点遇到这样的奇葩查询还会返回附近的8个节点吗?

逻辑不正确之二:
target衍生地址? 你的目的是爬虫,下一个节点接收到你查询自身的请求时难道还会返回8个节点吗?


python

逻辑不正确之一:  
随机2字节的事务id导致,一请求一应答的逻辑打破.一请求多应答?同样可以处理了,不严谨!

