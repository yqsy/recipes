
<!-- TOC -->

- [1. 说明](#1-说明)
- [2. basic authentication](#2-basic-authentication)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

实现两个版本:
* 基于https的basic authentication
* 基于tls的自定义二进制协议的认证


<a id="markdown-2-basic-authentication" name="2-basic-authentication"></a>
# 2. basic authentication

经测试,账号密码是通过base64传输的.每次通讯都会重复传输账号和密码.

```
Authorization: Basic YWRtaW46MTIzNDU2
```

![](http://ouxarji35.bkt.clouddn.com/basic.png)


命令
```
# 生成私钥
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out ./fd.key

# 自签证书
openssl req -new -x509 -days 365 -key fd.key -out fd.crt \
-subj "/C=GB/L=London/O=Feisty Duck Ltd/CN=www.example.com"

# 启动
sudo go run https_basic_authentication.go :20001
```

