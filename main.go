package main

import (
	"dbapp/dbapp"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	_ "net/http/pprof"
)

const (
	DefaultServerPort = 4001
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	config := &dbapp.DBAppConfig{
		ServerPort:            DefaultServerPort,
		ServerDBName:          "",
		ServerUser:            "dbapp",
		ServerPassword:        "dbapp",
		MySQLConnPoolMinALive: 10,
		MySQLConnPoolMaxAlive: 100,
		MySQLConnPoolMaxIdle:  120,
		MySQLAddress:          "127.0.0.1:3306",
		MySQLUser:             "root",
		MySQLPassword:         "root",
		RedisAddress:          "127.0.0.1:6379",
		RedisPoolSize:         1000,
	}
	dbapp.Start(config)
}
