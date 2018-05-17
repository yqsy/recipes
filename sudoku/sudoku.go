package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/yqsy/algorithm/sudoku/sudoku_extra"
)

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
	r := gin.Default()
	r.GET("/sudoku/:subject", serveSudoku)
	r.Run(":20000") // listen and serve on 0.0.0.0:8080
}
