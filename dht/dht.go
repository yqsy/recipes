package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/Sirupsen/logrus"
	"github.com/yqsy/recipes/dht/helpful"
	"github.com/yqsy/recipes/dht/hashinfo"
	"github.com/gin-gonic/gin"
	"github.com/yqsy/recipes/dht/inspector"
	"github.com/yqsy/recipes/dht/metadata"
)

const (
	InspectorAddr = ":20001"
)

func main() {
	logrus.AddHook(helpful.ContextHook{})

	ins := inspector.Inspector{UnReplyTid: make(map[string]struct{})}

	hashInfoGetter := hashinfo.NewHashInfoGetter(&ins)
	go func() {
		if err := hashInfoGetter.Run(); err != nil {
			panic(err)
		}
	}()

	metaGetter := metadata.MetaGetter{Ins: &ins}
	go func() {
		if err := metaGetter.Run(hashInfoGetter.MetaSourceChan); err != nil {
			panic(err)
		}
	}()

	helpInspector := inspector.HelpInspect{Ins: &ins}
	r := gin.Default()
	r.GET("/BasicInfo", helpInspector.BasicInfo())
	r.GET("/AllNodes", helpInspector.AllNodes())
	r.Run(InspectorAddr)
}
