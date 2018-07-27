package main

import (
	"github.com/yqsy/recipes/dht/bencode"
	"reflect"
	"github.com/pkg/errors"
	"encoding/json"
	"bytes"
	"fmt"
	"os"
	"io/ioutil"
	"crypto/sha1"
	"encoding/hex"
)

var usage = `Usage:
%v .torrent(file)
`

type TorrentMeta struct {
	MetaWithOutPieces map[string]interface{}
	Pieces            string
}

// parse (bencode) to TorrentMeta
// {
//     "info": {
//         "pieces": ""
//     }
// }
func ParseTorrentBytes(torrentByte []byte) (*TorrentMeta, error) {
	if torrentDecoded, err := bencode.Decode(string(torrentByte)); err != nil || reflect.TypeOf(torrentDecoded).Kind() != reflect.Map {
		return nil, errors.New(".torrent invalid")
	} else {

		if torrentInfo, ok := torrentDecoded.(map[string]interface{})["info"]; !ok || reflect.TypeOf(torrentInfo).Kind() != reflect.Map {
			return nil, errors.New(".torrent invalid")
		} else {

			torrentInfoMap := torrentInfo.(map[string]interface{})

			if torrentPieces, ok := torrentInfoMap["pieces"]; ! ok || reflect.TypeOf(torrentPieces).Kind() != reflect.String {
				return nil, errors.New(".torrent invalid")
			} else {

				torrentInfoMap["pieces"] = ""

				torrentMeta := &TorrentMeta{
					MetaWithOutPieces: torrentDecoded.(map[string]interface{}),
					Pieces:            torrentPieces.(string),
				}
				return torrentMeta, nil
			}
		}
	}
}

// https://stackoverflow.com/questions/2572521/extract-the-sha1-hash-from-a-torrent-file
// get "info" : { }
func ExtractTorrentBytes(torrentByte []byte) ([]byte, error) {
	if torrentDecoded, err := bencode.Decode(string(torrentByte)); err != nil || reflect.TypeOf(torrentDecoded).Kind() != reflect.Map {
		return nil, errors.New(".torrent invalid")
	} else {
		if torrentInfo, ok := torrentDecoded.(map[string]interface{})["info"]; !ok || reflect.TypeOf(torrentInfo).Kind() != reflect.Map {
			return nil, errors.New(".torrent invalid")
		} else {
			infoEncoded := []byte(bencode.Encode(torrentInfo))
			return infoEncoded, nil
		}
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	torrentByte, err := ioutil.ReadFile(arg[1])
	if err != nil {
		panic(err)
	}

	if torrentMeta, err := ParseTorrentBytes(torrentByte); err != nil {
		panic(err)
	} else {
		jsonRaw := []byte (bencode.Prettify(torrentMeta.MetaWithOutPieces))
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, jsonRaw, "", "    "); err != nil {
			panic(err)
		} else {
			// print json pretty with out Pieces bytes
			//fmt.Printf("%v\n", string(prettyJSON.Bytes()))

			// print pieces
			//pieceLen := len(torrentMeta.Pieces)
			//if pieceLen%20 != 0 {
			//	panic("invalid pieceLen")
			//}
			//fmt.Printf("piecesNum: %v\n", pieceLen/20)
			//for i := 0; i < pieceLen/20; i += 1 {
			//	curPiece := torrentMeta.Pieces[i*20 : i*20+20]
			//	pieceHex := hex.EncodeToString([]byte(curPiece))
			//	fmt.Printf("%v %v\n", i, pieceHex)
			//}

			//print hash
			extracted, err := ExtractTorrentBytes(torrentByte)
			if err != nil {
				panic(err)
			}
			sha1Sum := sha1.Sum(extracted)
			fmt.Printf("hash: %v\n", hex.EncodeToString(sha1Sum[:]))

		}
	}
}
