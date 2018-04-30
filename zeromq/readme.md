<!-- TOC -->

- [1. 说明](#1-说明)
- [2. Latency benchmark](#2-latency-benchmark)

<!-- /TOC -->

<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明
参考
* http://zeromq.org/distro:debian (install)
* http://zeromq.org/area:results (benchmark)
* http://zeromq.org/results:perf-howto
* http://zguide.zeromq.org/ (document)
* http://zeromq.org/intro:get-the-software (从源码安装)

库安装的方法
```bash
sudo bash -c "cat >> /etc/apt/sources.list" << EOF
deb http://download.opensuse.org/repositories/network:/messaging:/zeromq:/git-stable/Debian_8.0/ ./
EOF

wget https://download.opensuse.org/repositories/network:/messaging:/zeromq:/git-stable/Debian_8.0/Release.key -O- | sudo apt-key add
sudo apt-get install libzmq3-dev -y

# 卸了把
sudo dpkg -r libzmq3-dev -y

```

不过这次要benchmark,所以要从源码编译出来


```bash
cd refer
git clone https://github.com/zeromq/libzmq
cd libzmq
./autogen.sh && ./configure && make -j 4
make check && sudo make install && sudo ldconfig
```

<a id="markdown-2-latency-benchmark" name="2-latency-benchmark"></a>
# 2. Latency benchmark

zeromq
```
cd /media/yq/ST1000DM003/linux/reference/refer/libzmq/perf

```
