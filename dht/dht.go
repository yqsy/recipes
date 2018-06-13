package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/yqsy/recipes/dht/helpful"
	"github.com/yqsy/recipes/dht/dht"
	"github.com/gin-gonic/gin"
	"github.com/yqsy/recipes/dht/inspector"
)

const (
	InspectorPort = 20001
)

func main() {
	logrus.AddHook(helpful.ContextHook{})

	d := dht.NewDht()

	go func() {
		if err := d.Serve(); err != nil {
			panic(err)
		}
	}()

	helpInspector := inspector.HelpInspect{Ins: &d.Ins}

	r := gin.Default()
	r.GET("/BasicInfo", helpInspector.BasicInfo())

}
