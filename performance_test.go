package main

import (
	"context"
	"dbapp/dbapp"
	"dbapp/performance"
	"fmt"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/siddontang/go-log/log"
	"golang.org/x/sync/errgroup"
	"testing"
	"time"
)

const (
	minALive = 10
	maxAlive = 100
	maxIdle  = 60
)

func TestCompareDuration(t *testing.T) {
	count := 10000
	log.Infoln("start")
	testProxy(count)

	testDirectMySQL(count)

	testDirectRedis(count)
}

func testProxy(count int) {
	mysqlAddress := fmt.Sprintf("127.0.0.1:%v", DefaultServerPort)
	sql := "select SQL_CACHE * from article limit 10000,2"
	password := "dbapp"
	user := "dbapp"
	dbName := "test"
	connPool := client.NewPool(log.Debugf, minALive, maxAlive, maxIdle, mysqlAddress, user, password, dbName)

	monitor := performance.StartNewMonitorWithTimeUnit("proxy", time.Millisecond)
	runSelect(count, connPool, sql)
	monitor.End()
}

func testDirectMySQL(count int) {
	mysqlAddress := "127.0.0.1:3306"
	sql := "select SQL_CACHE * from article limit 10000,2"
	password := "root"
	user := "root"
	dbName := "test"
	connPool := client.NewPool(log.Debugf, minALive, maxAlive, maxIdle, mysqlAddress, user, password, dbName)

	monitor := performance.StartNewMonitorWithTimeUnit("MySQL", time.Millisecond)
	runSelect(count, connPool, sql)
	monitor.End()
}

func testDirectRedis(count int) {
	sql := "select SQL_CACHE * from article limit 10000,2"
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
		RedisPassword:         "",
	}
	redisClient := dbapp.NewGenericRedisClient(config)
	monitor := performance.StartNewMonitorWithTimeUnit("direct Redis", time.Millisecond)
	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)

	for i := 0; i < count; i++ {
		g.Go(func() error {
			_, err := redisClient.Get(ctx, sql).Bytes()
			if err != nil {
				return err
			}
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		fmt.Println(err)
	}
	monitor.End()
}

func runSelect(count int, connPool *client.Pool, sql string) {
	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	for i := 0; i < count; i++ {
		g.Go(func() error {
			conn, err := connPool.GetConn(ctx)
			//fmt.Printf("conn:%v,err:%v\n", conn, err)
			if err != nil {
				fmt.Println("get conn fail", err)
				return err
			}
			_, e := conn.Execute(sql)
			if e != nil {
				log.Errorf("%v", e)
				return e
			}
			connPool.PutConn(conn)
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
