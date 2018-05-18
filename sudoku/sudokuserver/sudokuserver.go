package main

import (
	"github.com/gin-gonic/gin"
	"github.com/yqsy/algorithm/sudoku/sudoku_extra"
	"net/http"
	"os"
	"fmt"
)

var usage = `Usage:
%v listenAddr
`

func serveSudoku(c *gin.Context) {
	subject := c.Param("subject")
	table, err := sudoku_extra.ConvertLineToTable(subject)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "error",
			"error":  err.Error()})
		return
	}

	if sudoku_extra.Solve(table, 0) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"result": table.GetLine()})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": "error",
			"error":  "solve error"})
	}
}

func main() {
	arg := os.Args
	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	r := gin.Default()
	r.GET("/sudoku/:subject", serveSudoku)
	r.Run(arg[1])
}
