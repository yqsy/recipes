<!-- TOC -->

- [1. 说明](#1-说明)
- [2. benchmark sudoku](#2-benchmark-sudoku)

<!-- /TOC -->



<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

* https://thrift.apache.org/tutorial/go (go说明)
* https://thrift.apache.org/docs/install/ (安装说明)
* https://github.com/apache/thrift
* http://diwakergupta.github.io/thrift-missing-guide/ 
* https://github.com/apache/thrift/tree/master/tutorial/go (参考例子)

我的deepin安装实践
```bash
sudo apt-get install automake bison flex g++ git libboost-all-dev libevent-dev libssl-dev libtool make pkg-config -y

cd /media/yq/ST1000DM003/linux/reference/refer

git clone https://github.com/apache/thrift.git
cd thrift

./bootstrap.sh
./configure
make
sudo make install
sudo make -k check

go get git.apache.org/thrift.git/lib/go/thrift

thrift -r --gen go sudoku_protocol.thrift
```

<a id="markdown-2-benchmark-sudoku" name="2-benchmark-sudoku"></a>
# 2. benchmark sudoku
