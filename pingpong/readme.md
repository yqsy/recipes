<!-- TOC -->

- [1. 说明](#1-说明)
- [测试脚本](#测试脚本)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

参考
* http://think-async.com/Asio/LinuxPerformanceImprovements
* https://sourceforge.net/projects/asio/ (asio单独库的下载)



```bash
cd /media/yq/ST1000DM003/linux/reference/refer/asio-1.12.1
./configure
make
sudo make install 
```

<a id="markdown-测试脚本" name="测试脚本"></a>
# 测试脚本

```bash
cat > ./bench.sh << EOF
#!/bin/bash

killall server
timeout=5
for bufsize in 16384 32768 65536
do
  for nothreads in 1 2 4 
  do
    for nosessions in 1 10 100 
    do
      echo "Bufsize: \$bufsize Threads: \$nothreads Sessions: \$nosessions"
      ./server 0.0.0.0 55555 \$nothreads \$bufsize & srvpid=\$!
      ./client localhost 55555 \$nothreads \$bufsize \$nosessions \$timeout 
      kill -9 \$srvpid
    done
  done
done

EOF

chmod +x bench.sh


```
