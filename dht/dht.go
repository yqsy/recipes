package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/yqsy/recipes/dht/hashinfo"
	"github.com/gin-gonic/gin"
	"github.com/yqsy/recipes/dht/inspector"
	"github.com/op/go-logging"
	"os"
	"io/ioutil"
)

const (
	InspectorAddr = ":20001"
)

var log = logging.MustGetLogger("dht")

var format = logging.MustStringFormatter(
	`%{color}%{time:20060102 15:04:05.000000} %{id} %{level:.4s}%{color:reset} %{message} - %{shortfile}`,
)

func main() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)

	ins := inspector.Inspector{UnReplyTid: make(map[string]struct{})}

	hashInfoGetter := hashinfo.NewHashInfoGetter(&ins)
	go func() {
		if err := hashInfoGetter.Run(); err != nil {
			panic(err)
		}
	}()

	//metaGetter := metadata.MetaGetter{Ins: &ins}
	//go func() {
	//	if err := metaGetter.Run(hashInfoGetter.MetaSourceChan); err != nil {
	//		panic(err)
	//	}
	//}()

	helpInspector := inspector.HelpInspect{Ins: &ins}
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r := gin.Default()
	r.GET("/BasicInfo", helpInspector.BasicInfo())
	r.Run(InspectorAddr)
}
