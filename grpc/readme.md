<!-- TOC -->

- [1. 说明](#1-说明)
- [2. benchmark sudoku](#2-benchmark-sudoku)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明


```bash
# 生成
protoc -I sudoku_protocol/ sudoku_protocol/sudoku_protocol.proto --go_out=plugins=grpc:sudoku_protocol

# -I 表示import的根目录
# 参数1 表示元数据文件
# go_out=插件:生成的目录


# 服务器
go run sudoku_server.go :20000

# 客户端
go run sudoku_client.go 127.0.0.1:20000 080001030500804706000270000920400003103958402400002089000029000305106008040300010

```

<a id="markdown-2-benchmark-sudoku" name="2-benchmark-sudoku"></a>
# 2. benchmark sudoku


