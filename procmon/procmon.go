package main

import (
	"net/http"
	"fmt"
	"os"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
)

func cpuHandler(w http.ResponseWriter, r *http.Request) {
	infoStat, err := cpu.Info()

	if err != nil {
		return
	}

	for i := 0; i < len(infoStat); i++ {
		fmt.Fprint(w, infoStat[i])
	}
}

func memHandler(w http.ResponseWriter, r *http.Request) {
	virtualMemoryStat, err := mem.VirtualMemory()

	if err != nil {
		return
	}

	fmt.Fprint(w, virtualMemoryStat)
}

func diskHandler(w http.ResponseWriter, r *http.Request) {

	paritions, err := disk.Partitions(true)
	if err != nil {
		return
	}
	for i := 0; i < len(paritions); i++ {
		fmt.Fprint(w, paritions[i])
	}
}

func netHandler(w http.ResponseWriter, r *http.Request) {
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return
	}
	for i := 0; i < len(ioCounters); i++ {
		fmt.Fprint(w, ioCounters[i])
	}
}

func main() {

	arg := os.Args
	if len(arg) < 2 {
		fmt.Printf("Usage:\n %v listenaddr\nExample:\n %v :8080\n", arg[0], arg[0])
		return
	}

	http.HandleFunc("/cpu", cpuHandler)
	http.HandleFunc("/mem", memHandler)
	http.HandleFunc("/disk", diskHandler)
	http.HandleFunc("/net", netHandler)
	err := http.ListenAndServe(arg[1], nil)

	if err != nil {
		panic(err)
	}
}
