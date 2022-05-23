package main

import (
	"context"
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
}

func testProxy(count int) {
	mysqlAddress := fmt.Sprintf("127.0.0.1:%v", DefaultServerPort)
	sql := "select SQL_CACHE * from trade limit 2"
	password := "dbapp"
	user := "dbapp"
	dbName := "test"
	connPool := client.NewPool(log.Debugf, minALive, maxAlive, maxIdle, mysqlAddress, user, password, dbName)

	startTime := time.Now()
	runSelect(count, connPool, sql)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Infoln("proxy耗时:", duration.Milliseconds())
}

func testDirectMySQL(count int) {
	mysqlAddress := "127.0.0.1:3306"
	sql := "select * from trade limit 2"
	password := "root"
	user := "root"
	dbName := "test"
	connPool := client.NewPool(log.Debugf, minALive, maxAlive, maxIdle, mysqlAddress, user, password, dbName)

	startTime := time.Now()
	runSelect(count, connPool, sql)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Infoln("直接MySQL耗时:", duration.Milliseconds())
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
