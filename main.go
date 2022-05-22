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
		ServerDBName:          "test",
		ServerUser:            "dbapp",
		ServerPassword:        "dbapp",
		MySQLConnPoolMinALive: 10,
		MySQLConnPoolMaxAlive: 10000,
		MySQLConnPoolMaxIdle:  10,
		MySQLAddress:          "127.0.0.1:3306",
		MySQLUser:             "root",
		MySQLPassword:         "root",
		RedisAddress:          "127.0.0.1:6379",
		RedisPoolSize:         10000,
	}
	dbapp.Start(config)
}
