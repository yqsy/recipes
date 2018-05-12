<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 编译](#2-编译)

<!-- /TOC -->

<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://tools.ietf.org/html/rfc3875
* https://github.com/yqsy/tinyhttpd

```
webserver --> [interface(cgi):webframe]
```

特点:

输入:
* GET: 环境变量作为输入(`固定的key值`让cgi去解析)
* POST: 带有CONTENT,那么把`CONTENT内容`写入到`stdin`中

输出:
* 输出到stdout

<a id="markdown-2-编译" name="2-编译"></a>
# 2. 编译

```bash
# 编译
mkdir -p bin

cd cgi
go build cgi.go
mv cgi ../bin

cd ../webserver
go build webserver.go
mv webserver ../bin

cd ..


# 静态资源
cp static/index.html bin


# 刷库脚本
# db目录下

```
