<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 我梳理的概要](#2-我梳理的概要)

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
