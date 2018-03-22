<!-- TOC -->

- [1. httpproxy](#1-httpproxy)
- [2. 其他](#2-其他)
- [3. 单元测试](#3-单元测试)
    - [3.1. http](#31-http)
    - [3.2. https](#32-https)
- [4. 吞吐量测试](#4-吞吐量测试)

<!-- /TOC -->


<a id="markdown-1-httpproxy" name="1-httpproxy"></a>
# 1. httpproxy

* https://tools.ietf.org/html/rfc2616
* https://en.wikipedia.org/wiki/HTTP_tunnel
* https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol#Request_methods

```
                    ===
http === http proxy ===
                    ===
```

http proxy会对第一行以及header进行解析,解析的`目的是得到ip和端口`,然后转发`第一行`,`header`,其余当作tcp流量转发
```
GET http://baidu.com/ HTTP/1.1\r\n
User-Agent: curl/7.29.0\r\n
Host: baidu.com\r\n
Accept: */*\r\n
Proxy-Connection: Keep-Alive\r\n
\r\n
```

https proxy只会获取`第一行即可得知ip和端口`,然后进行tcp流量转发(注意要把CONNECT的包头消费掉,以及给一个应答)
```
CONNECT baidu.com:443 HTTP/1.1\r\n
Host: baidu.com:443\r\n
User-Agent: curl/7.29.0\r\n
Proxy-Connection: Keep-Alive\r\n
\r\n

HTTP/1.1 200 Connection established\r\n
\r\n
```

<a id="markdown-2-其他" name="2-其他"></a>
# 2. 其他

简单bash测试
```bash
export http_proxy=http://host1:20001
export https_proxy=http://host1:20001


curl -v https://baidu.com
curl -v http://baidu.com
```

<a id="markdown-3-单元测试" name="3-单元测试"></a>
# 3. 单元测试

<a id="markdown-31-http" name="31-http"></a>
## 3.1. http

正确:
* GET http://xxx.com/ HTTP/1.1\r\n
* GET / HTTP/1.1\r\n 域名在header
* 包头一个一个字节传

错误:
* 第一行不满足3个
* 没有端口 也没有host
* 第一行长度攻击 (暂时没有,go的http解析模块暂时是无限制增长的)
* header长度攻击 (暂时没有,go的http解析模块暂时是无限制增长的)


<a id="markdown-32-https" name="32-https"></a>
## 3.2. https

正确:
* 普通 CONNECT

错误:
* 第一行长度攻击 (暂时没有,go的http解析模块暂时是无限制增长的)

<a id="markdown-4-吞吐量测试" name="4-吞吐量测试"></a>
# 4. 吞吐量测试

```
# 直接连3W qps
ab -n 100000 -c 10 http://localhost:20001/hello


# 通过proxy 1W qps
go run httpproxy.go :1080 > /dev/zero
ab -n 100000 -c 10 -X localhost:1080  http://localhost:20001/hello
```
