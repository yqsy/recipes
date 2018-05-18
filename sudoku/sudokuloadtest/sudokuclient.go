package main

import (
	"os"
	"fmt"
	"strconv"
	"net/http"
	"time"
	"sync/atomic"
	"io"
	"io/ioutil"
	"sync"
	"sort"
	"log"
)

var usage = `Usage:
%v connectAddr qps(each connection) connections problem
`

type Context struct {
	// 每1s生成一个快照打印到stdout以及flush到磁盘上
	latencies []int

	// 保护latencies
	mtx sync.Mutex

	qps int

	req *http.Request

	client *http.Client

	// 当前要被发送的请求数
	currentSend *int64
}

func (ctx *Context) addLatency(latency int) {
	ctx.mtx.Lock()
	ctx.latencies = append(ctx.latencies, latency)
	ctx.mtx.Unlock()
}

func (ctx *Context) getSnapshot(dst *[]int) {
	ctx.mtx.Lock()
	*dst = append(*dst, ctx.latencies...)
	ctx.latencies = []int{}
	ctx.mtx.Unlock()
}

func runClient(ctx *Context) {

	// 每秒钟增加发送数
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		for {
			select {
			case <-ticker.C:
				atomic.AddInt64(ctx.currentSend, int64(ctx.qps/10))
			}
		}
	}()

	time.Sleep(time.Millisecond * 10)

	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		select {
		case <-ticker.C:
			reqs := atomic.SwapInt64(ctx.currentSend, 0)

			for i := 0; i < int(reqs); i++ {
				t1 := time.Now()
				res, err := ctx.client.Do(ctx.req)
				if err != nil {
					panic(err)
				}

				io.Copy(ioutil.Discard, res.Body)
				res.Body.Close()
				escaped := int(time.Since(t1).Nanoseconds()) / 1000
				ctx.addLatency(escaped)
			}
		}
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 4 {
		fmt.Printf(usage)
		return
	}

	connectAddr := arg[1]
	qps, err := strconv.Atoi(arg[2])
	if err != nil {
		panic(err)
	}
	connections, err := strconv.Atoi(arg[3])
	if err != nil {
		panic(err)
	}

	problem := arg[4]

	log.Printf("connectAddr: %v qps: %v connections:%v\n", connectAddr, qps, connections)

	ctxs := make([]*Context, 0)
	for i := 0; i < connections; i++ {

		req, err := http.NewRequest("GET", "http://"+connectAddr+"/sudoku/"+problem, nil)

		if err != nil {
			panic(err)
		}

		ctx := &Context{
			latencies:   []int{},
			qps:         qps,
			req:         req,
			client:      &http.Client{},
			currentSend: new(int64),
		}

		ctxs = append(ctxs, ctx)

		go runClient(ctx)
	}

	ticker := time.NewTicker(time.Second)

	// 每一次快照的分布都写到文件内
	var fileCount int

	for {
		select {
		case <-ticker.C:
			var latencies []int
			for i := 0; i < connections; i++ {
				ctxs[i].getSnapshot(&latencies)
			}

			sort.Ints(latencies)
			min, max, mean, median, p90, p99 := getQuota(latencies)
			report := fmt.Sprintf("recv %v min %v max %v mean %v median %v p90 %v p99 %v\n",
				len(latencies), min, max, mean, median, p90, p99)
			log.Printf(report)

			fileCount += 1
			fileName := fmt.Sprintf("r%04d", fileCount)
			writeToFile(fileName, report, latencies)
		}
	}
}

func writeToFile(fileName string, report string, latencies []int) {
	if len(latencies) < 1 {
		return
	}

	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.Write([]byte("# " + report + "\n"))

	const interval = 5
	low := latencies[0] / interval * interval

	var count int
	var sum int
	for _, latency := range latencies {
		if latency < low+interval {
			count += 1
		} else {
			sum += count
			line := fmt.Sprintf("%4d %5d %5.2f\n", low, count, 100*float64(sum)/float64(len(latencies)))
			f.Write([]byte(line))
			low = latency / interval * interval
			count = 1
		}
	}

	sum += count
	if sum != len(latencies) {
		panic("error")
	}

	line := fmt.Sprintf("%4d %5d %5.2f\n", low, count, 100*float64(sum)/float64(len(latencies)))
	f.Write([]byte(line))
}

func getQuota(latencies []int) (min, max, mean, median, p90, p99 int) {
	min = getMinFloat(latencies)
	max = getMaxFloat(latencies)
	mean = getMean(latencies)
	median = getPercent(latencies, 50)
	p90 = getPercent(latencies, 90)
	p99 = getPercent(latencies, 99)
	return min, max, mean, median, p90, p99
}

func getMinFloat(s []int) int {
	if len(s) < 1 {
		return 0
	}

	min := s[0]
	for _, e := range s {
		if e < min {
			min = e
		}
	}
	return min
}

func getMaxFloat(s []int) int {
	if len(s) < 1 {
		return 0
	}

	max := s[0]
	for _, e := range s {
		if e > max {
			max = e
		}
	}
	return max
}

func getSum(s []int) int {
	var sum int
	for _, e := range s {
		sum += e
	}
	return sum
}

func getMean(s []int) int {
	return getSum(s) / int(len(s))
}

func getPercent(s []int, percent int) int {
	if len(s) < 1 {
		return 0
	}

	var idx int
	if percent > 0 {
		idx = len(s)*percent/100 - 1

		if idx < 0 {
			idx = 0
		}
	}
	return s[idx]
}
