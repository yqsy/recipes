package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/Sirupsen/logrus"
	"github.com/yqsy/recipes/dht/helpful"
	"github.com/yqsy/recipes/dht/dht"
	"github.com/gin-gonic/gin"
	"github.com/yqsy/recipes/dht/inspector"
	"os"
	"database/sql"
	"fmt"
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

		mysqlpassword := os.Getenv("MYSQL_PASSWORD")
		db, err := sql.Open("mysql", fmt.Sprintf("root:%v@/dht?charset=utf8", mysqlpassword))

		if err != nil {
			panic(err)
		}

		// TODO d.Ins.HashInfoNumberAll read from db
		for {
			hashInfo := <-d.HashInfoChan
			if _, ok := uniqDict[hashInfo]; !ok {
				uniqDict[hashInfo] = struct{}{}

				stmt, err := db.Prepare("insert hashinfos set hashinfo=?")
				if err != nil {
					logrus.Warnf("inser error: %v", err)
				}

				_, err = stmt.Exec(hashInfo)

				if err != nil {
					logrus.Warnf("exec error: %v", err)
				}

				stmt.Close()

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
