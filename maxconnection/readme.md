<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 错误的提示](#2-错误的提示)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

最大连接数在应用层的考虑点有两个
* 作为客户端,connect  最大连接数多少?无法分配新的fd会怎么样?
* 作为服务端,accept 最大连接数多少?无法分配新的fd会怎么样?

参考:
* http://www.ideawu.net/blog/archives/740.html#q1


相关配置大概有:

```bash
# 全局限制
# 查看
cat /proc/sys/fs/file-nr
# 修改
sudo bash -c "cat >> /etc/sysctl.conf" << EOF
fs.file-max = 1020000
EOF


# 进程限制
# 查看
ulimit -n
# 修改
sudo bash -c "cat >> /etc/security/limits.conf" << EOF
*      soft  nofile    1020000
*      hard  nofile    1020000
EOF

# port range
# 查看
sysctl -a | grep net.ipv4.ip_local_port_range
# 修改
sudo bash -c "cat >> /etc/sysctl.conf" << EOF
net.ipv4.ip_local_port_range = 1024 65535
EOF


# 读写缓冲区(内存角度)
cat /proc/sys/net/ipv4/tcp_wmem
cat /proc/sys/net/ipv4/tcp_rmem

```

<a id="markdown-2-错误的提示" name="2-错误的提示"></a>
# 2. 错误的提示

文件描述符软限制不足够accept了:  
too many open files

端口数量不足够生成新的套接字去connect了:  
dial tcp 127.0.0.1:5001: connect: cannot assign requested address
