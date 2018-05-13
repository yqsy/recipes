<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 指令](#2-指令)
- [3. 细节](#3-细节)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://en.wikipedia.org/wiki/FastCGI
* https://gist.github.com/atomaths/5403041 (参考了一个韩国人的..)

简单来说,CGI会不断地创建释放进程,造成大量不必要的开销,以及不能复用数据库连接,内存缓存

本程序用来测试nginx+fastcgi的负载均衡手段

<a id="markdown-2-指令" name="2-指令"></a>
# 2. 指令
```
wget https://raw.githubusercontent.com/nginx/nginx/master/conf/nginx.conf

# 修改
sudo mv  /etc/nginx/nginx.conf /etc/nginx/nginx.conf.bak
sudo cp ./nginx.conf /etc/nginx/nginx.conf

go run fastcgi.go :20001
go run fastcgi.go :20002

# 观察 是否能轮询方式的负载均衡
curl http://127.0.0.1:20000/foo/123
```


<a id="markdown-3-细节" name="3-细节"></a>
# 3. 细节

```bash
# 先读这样一个包头
type header struct {
	Version       uint8
	Type          recType
	Id            uint16
	ContentLength uint16
	PaddingLength uint8
	Reserved      uint8
}


# 再读满buf,长度如下
n := int(rec.h.ContentLength) + int(rec.h.PaddingLength)

```