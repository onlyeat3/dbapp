package main

import (
	"dbapp/dbapp"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	config := &dbapp.DBAppConfig{
		ServerPort:            4000,
		ServerDBName:          "",
		ServerUser:            "dbapp",
		ServerPassword:        "dbapp",
		MySQLConnPoolMinALive: 10,
		MySQLConnPoolMaxAlive: 500,
		MySQLConnPoolMaxIdle:  120,
		MySQLConnPoolAddress:  "127.0.0.1:3306",
		RedisAddress:          "10.91.14.186:32239",
		RedisPoolSize:         48,
	}
	dbapp.Start(config)
}
