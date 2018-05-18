package main

import (
	"os"
	"fmt"
	"strconv"
	"github.com/yqsy/algorithm/sudoku/sudoku_extra"
	"time"
)

var usage = `Usage:
%v num problem
`

func main() {
	arg := os.Args
	usage = fmt.Sprintf(usage, arg[0])
	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	num, err := strconv.Atoi(arg[1])
	if err != nil {
		panic(err)
	}

	problem := arg[2]

	t1 := time.Now()

	for i := 0; i < num; i ++ {
		table1Extra, _ := sudoku_extra.ConvertLineToTable(problem)
		var pos int
		sudoku_extra.Solve(table1Extra, pos)
	}

	t2 := float64(time.Since(t1).Nanoseconds()) / 1000 / 1000

	fmt.Printf("cost: %.2f ms\n", t2)
}
