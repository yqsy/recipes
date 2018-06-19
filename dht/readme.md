<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 我梳理的概要](#2-我梳理的概要)
- [3. 实现思路](#3-实现思路)
- [4. find_node](#4-find_node)
- [5. 恶心的地方](#5-恶心的地方)
- [6. 实现bencode解析](#6-实现bencode解析)

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
* http://www.bittorrent.org/beps/bep_0010.html (hashinfo -> metadata)

rfc文档
* http://jonas.nitro.dk/bittorrent/bittorrent-rfc.html (还是这个靠谱)


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


<a id="markdown-3-实现思路" name="3-实现思路"></a>
# 3. 实现思路

* 发送find_node加入dht网络
* 接收find_node的应答 -> 继续发送find_node请求(来回往复)

---
* 接收get_peers请求,发送模拟空应答
* 接收announce_peer请求,存储hash_info

---
全双工协议,请求应答的思路:  
* 协议的"y"表示了是请求`"r"`还是应答`"q"`
* 用事务ID保持`请求应答匹配`的正确,唯一性,以及知道应答的type好做`dispatch`

<a id="markdown-4-find_node" name="4-find_node"></a>
# 4. find_node

语言|join|衍生|事务id
-|-|-|-
go|id:自身 target:自身|id:自身 target:自身+衍生地址混合|全局递增的id取2字节
python|id:自身 target:随机|id:衍生地址+自身混合 target:随机|随机2字节




<a id="markdown-5-恶心的地方" name="5-恶心的地方"></a>
# 5. 恶心的地方

恶心之一:

dht 协议恶心之处,就是他把example写成json这样,但其实和你想的差的很远.

例如`id,target`存储的是160bit长度的id号码.其存储方式不是字符串,虽然fuck tmd 的文档example里面写的是16进制字符.但是实际上是一串字节流,20个字节.

这样可以节省空间:如果用字节流传输,只需要20字节.如果使用16进制字符串(1字节=2个16进制字符串)则需要40个字符.字符串的存储空间是40字节.

如果是我来做这件事情,我绝对不会用 json + ""这种符号 的方式来误导大家,而是在文档的显著位置使用红色高亮字体告诫大家这里为了节省空间,使用20字节来表达出一个160位的号码,请大家编程的时候使用xxx(示例)函数将其转换成16进制的字符串增加它的可阅读性.

恶心之二:

http://www.bittorrent.org/beps/bep_0009.html


extension message  你搞一个类似json表达格式的协议也就算了,但是能把二进制数据好好表示吗,跟在数据后面是啥意思?

<a id="markdown-6-实现bencode解析" name="6-实现bencode解析"></a>
# 6. 实现bencode解析



把json拿来对比

数据类型|实例值|解析思路|存储|bencode
-|-|-|-|-
null|null|n开头|nil|
boolean|true false| t f 开头|bool|
string|""|"开头|string|数字:字符串内容 (4:spam)
number|浮点数|默认|float64| i开头e末尾 (i42e)
array|[]|[开头|[]interface{}| l开头e末尾 (l4:spami42ee)
object|{...}|{开头|map[string]interface{}| d开头e末尾 (d3:bar4:spam3:fooi42ee)

这个只是一个简单的工程,尽量简易以及容易使用,以下特性:

* 二进制与Value树的转换, Encode, Decode 函数
* 内置数据结构与Value树的转换
  * (NewString, NewNumber, NewArray, NewObject) 是内置数据结构转Value树
  * Value树转内置结构
