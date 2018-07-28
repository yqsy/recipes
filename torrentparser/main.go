package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"github.com/yqsy/recipes/dht/bencode"
	"github.com/yqsy/typechekcer"
	"reflect"
	"bytes"
	"encoding/json"
	"crypto/sha1"
	"errors"
	"encoding/hex"
)

var usage = `Usage:
%v .torrent(file)
`

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	rawFlag := false
	if len(arg) > 2 {
		if arg[2] == "--raw" {
			rawFlag = true
		}
	}

	torrentByte, err := ioutil.ReadFile(arg[1])
	if err != nil {
		panic(err)
	}

	if rawFlag {
		rawPrint(torrentByte)
	} else {
		deepPrint(torrentByte)
	}
}

// 调用之前需要保证slice的每个元素都是string
func CombinePath(paths []interface{}, rootPath string) string {
	wholePath := rootPath
	for i := 0; i < len(paths); i++ {
		node := paths[i].(string)
		wholePath = wholePath + "/" + node
	}
	return wholePath
}

// "  files": [
//    {
//        "length": 2342023084,
//        "path": [
//            "The.Avengers.2012.复仇者联盟.双语字幕.HR-HDTV.AC3.1024X576.x264-人人影视制作.mkv"
//        ]
//    },
//    ...
//    ]
// 转换成
//    {
//      "path": "复仇者联盟DVD包/The.Avengers.2012.复仇者联盟.双语字幕.HR-HDTV.AC3.1024X576.x264-人人影视制作.mkv",
//      "name": "The.Avengers.2012.复仇者联盟.双语字幕.HR-HDTV.AC3.1024X576.x264-人人影视制作.mkv",
//      "length": 2342023084,
//      "offset": 0
//    },
func deepPrint(torrentByte []byte) {
	objInterface, err := bencode.Decode(string(torrentByte))
	if err != nil {
		panic(err)
	}

	if err := typechekcer.CheckMapValue(objInterface, "info", reflect.Map, reflect.Invalid); err != nil {
		panic(err)
	}

	// calc hash
	info := objInterface.(map[string]interface{})["info"].(map[string]interface{})
	infoEncoded := []byte(bencode.Encode(info))
	sha1Sum := sha1.Sum(infoEncoded)
	info["sha1_hash"] = hex.EncodeToString(sha1Sum[:])
	//fmt.Printf("hash: %v\n", hex.EncodeToString(sha1Sum[:]))

	// 清空
	info["pieces"] = ""

	// 计算文件
	if err := typechekcer.CheckMapValue(info, "files", reflect.Slice, reflect.Map); err != nil {
		panic(err)
	}

	// 获取根目录以作组合
	if err := typechekcer.CheckMapValue(info, "name", reflect.String, reflect.Invalid); err != nil {
		panic(err)
	}

	// 偏移
	var offset int

	files := info["files"].([]interface{})
	for i := 0; i < len(files); i++ {
		err := typechekcer.CheckMapValue(files[i], "length", reflect.Int, reflect.Invalid)
		if err != nil {
			panic(err)
		}

		err = typechekcer.CheckMapValue(files[i], "path", reflect.Slice, reflect.String)
		if err != nil {
			panic(err)
		}

		file := files[i].(map[string]interface{})
		path := file["path"].([]interface{})

		if len(path) < 1 {
			panic(errors.New("invalid path"))
		}

		// 全路径
		wholePath := CombinePath(path, info["name"].(string))

		// 文件名
		fileName := path[len(path)-1].(string)

		// 长度
		length := file["length"].(int)

		// 重写
		file["path"] = wholePath
		file["name"] = fileName
		file["length"] = length
		file["offset"] = offset

		// 偏移
		offset = offset + length
	}

	jsonRaw := []byte(bencode.Prettify(objInterface))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonRaw, "", "    "); err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", string(prettyJSON.Bytes()))
}

func rawPrint(torrentByte []byte) {
	objInterface, err := bencode.Decode(string(torrentByte))
	if err != nil {
		panic(err)
	}
	// 把pieces提取出来,其他保留不变
	if err := typechekcer.CheckMapValue(objInterface, "info.pieces", reflect.String, reflect.Invalid); err != nil {
		panic(err)
	}
	info := objInterface.(map[string]interface{})["info"].(map[string]interface{})
	piecesStorage := info["pieces"].(string)
	_ = piecesStorage
	info["pieces"] = ""

	// objInterface + piecesStorage 构成了torrent文件
	jsonRaw := []byte(bencode.Prettify(objInterface))
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonRaw, "", "    "); err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", string(prettyJSON.Bytes()))
}
