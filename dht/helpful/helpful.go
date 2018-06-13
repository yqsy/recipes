package helpful

import (
	"crypto/rand"
	"github.com/Sirupsen/logrus"
	"runtime"
	"path"
	"fmt"
)

func RandomString(len int) string {
	buf := make([]byte, len)
	rand.Read(buf)
	return string(buf)
}

type ContextHook struct{}

func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook ContextHook) Fire(entry *logrus.Entry) error {
	if _, file, line, ok := runtime.Caller(8); ok {
		entry.Data["file"] = path.Base(file)
		entry.Data["line"] = line
	}

	return nil
}

func GetHex(str string) (rtn string) {
	buf := []byte(str)
	for i := 0; i < len(buf); i++ {
		rtn += fmt.Sprintf("%02x", buf[i])
	}
	return rtn
}
