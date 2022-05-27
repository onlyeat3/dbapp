package main

import (
	"dbapp/dbapp"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
)

const (
	DefaultServerPort = 4001
)

func initLogger() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {

	initLogger()
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
		RedisPoolSize:         1000,
		RedisPassword:         "",
	}
	dbapp.Start(config)
}
