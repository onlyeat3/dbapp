package main

import (
	"dbapp/dbapp"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dbapp.Start()
}
