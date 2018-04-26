<!-- TOC -->

- [说明](#说明)

<!-- /TOC -->

<a id="markdown-说明" name="说明"></a>
# 说明

A <==encrypted==> B

参考:
* https://gist.github.com/denji/12b3a568f092ab951456

```bash
# 生成私钥
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out ./fd.key

# 自签证书
openssl req -new -x509 -days 365 -key fd.key -out fd.crt \
-subj "/C=GB/L=London/O=Feisty Duck Ltd/CN=www.example.com"

```
