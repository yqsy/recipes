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
		if err := d.Run(); err != nil {
			panic(err)
		}
	}()

	go func() {
		uniqDict := make(map[string]struct{})

		// TODO d.Ins.HashInfoNumberAll read from db
		for {
			hashInfo := <-d.HashInfoChan
			if _, ok := uniqDict[hashInfo]; !ok {
				uniqDict[hashInfo] = struct{}{}

				d.Ins.SafeDo(func() {
					d.Ins.HashInfoNumberSinceStart += 1
					d.Ins.HashInfoNumberAll += 1
				})
			}
		}
	}()

	helpInspector := inspector.HelpInspect{Ins: &d.Ins}

	r := gin.Default()
	r.GET("/BasicInfo", helpInspector.BasicInfo())

}
