<!-- TOC -->

- [说明](#说明)

<!-- /TOC -->

<a id="markdown-说明" name="说明"></a>
# 说明
参考
* http://zeromq.org/distro:debian

```bash

sudo bash -c "cat >> /etc/apt/sources.list" << EOF
deb http://download.opensuse.org/repositories/network:/messaging:/zeromq:/git-stable/Debian_8.0/ ./
EOF

wget https://download.opensuse.org/repositories/network:/messaging:/zeromq:/git-stable/Debian_8.0/Release.key -O- | sudo apt-key add
sudo apt-get install libzmq3-dev -y

```

