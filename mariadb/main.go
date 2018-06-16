package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"os"
)

var usage = `Usage:
%v dbpassword
`

func main() {

	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("root:%v@/hashinfo?charset=utf8", arg[1]))

	if err != nil {
		panic(err)
	}

	defer db.Close()

	Insert(db)
	fmt.Println("query after insert:")
	Query(db)

	Update(db)
	fmt.Println("query after update:")
	Query(db)

	Delete(db)
	fmt.Println("query after delete:")
	Query(db)

	Transaction(err, db)
	fmt.Println("query after Transaction insert commit:")
	Query(db)
	Delete(db)

	TransactionRollback(err, db)
	fmt.Println("query after Transaction insert rollback:")
	Query(db)
}

func TransactionRollback(err error, db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	stmt, err := tx.Prepare("insert hashinfos set hashinfo = ?")
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec("poiuytrewqasdfghjklz")
	if err != nil {
		panic(err)
	}
	err = tx.Rollback()
	if err != nil {
		panic(err)
	}
}

func Transaction(err error, db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	stmt, err := tx.Prepare("insert hashinfos set hashinfo = ?")
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec("poiuytrewqasdfghjklz")
	if err != nil {
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

func Update(db *sql.DB) {
	stmt, err := db.Prepare("update hashinfos set hashinfo= ? where hashinfo='09876543210987654321'")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec("abcdabcdabcdabcdabcd")
	if err != nil {
		panic(err)
	}
}

func Delete(db *sql.DB) {
	stmt, err := db.Prepare("delete from hashinfos")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}
}

func Insert(db *sql.DB) {
	stmt, err := db.Prepare("insert hashinfos set hashinfo = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec("09876543210987654321")
	if err != nil {
		panic(err)
	}
}

func Query(db *sql.DB) {
	stmt, err := db.Prepare("select hashinfo from hashinfos order by id")

	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var hashinfo string

		err = rows.Scan(&hashinfo)
		if err != nil {
			panic(err)
		}

		fmt.Println(hashinfo)
	}
}
