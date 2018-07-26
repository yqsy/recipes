package main

import (
	"io/ioutil"
	"github.com/yqsy/recipes/dht/bencode"
	"reflect"
	"github.com/pkg/errors"
	"encoding/json"
	"bytes"
	"fmt"
	"os"
	"encoding/hex"
	"crypto/md5"
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
func ParseTorrentFile(path string) (*TorrentMeta, error) {
	if torrentByte, err := ioutil.ReadFile(path); err != nil {
		return nil, errors.New(".torrent invalid")
	} else {
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
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	if torrentMeta, err := ParseTorrentFile(arg[1]); err != nil {
		panic(err)
	} else {
		jsonRaw := []byte (bencode.Prettify(torrentMeta.MetaWithOutPieces))
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, jsonRaw, "", "    "); err != nil {
			panic(err)
		} else {
			fmt.Printf("%v\n", string(prettyJSON.Bytes()))

			pieceLen := len(torrentMeta.Pieces)
			if pieceLen%20 != 0 {
				panic("invalid pieceLen")
			}

			fmt.Printf("piecesLen: %v\n", len(torrentMeta.Pieces))
			md5Hash := md5.Sum([]byte(torrentMeta.Pieces))
			fmt.Printf("hash: %v\n", hex.EncodeToString(md5Hash[:]))
			for i := 0; i < pieceLen/20; i += 20 {
				curPiece := torrentMeta.Pieces[i : i+20]
				pieceHex := hex.EncodeToString([]byte(curPiece))
				fmt.Printf("%v %v\n", i/20, pieceHex)
			}
		}
	}
}
