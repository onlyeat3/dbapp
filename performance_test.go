package main

import (
	"context"
	"fmt"
	"testing"
	"time"
	"virtdb/performance"
	"virtdb/virtdb"

	"github.com/go-mysql-org/go-mysql/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	minALive = 10
	maxAlive = 100
	maxIdle  = 60
)

func TestCompareDuration(t *testing.T) {
	count := 10000
	sql := "select SQL_CACHE * from article limit 10000,2"

	log.Infoln("start")

	testProxy(count, sql)
	testDirectRedis(count, sql)
	testDirectMySQL(count, sql)
}

func testProxy(count int, sql string) {
	mysqlAddress := fmt.Sprintf("127.0.0.1:%v", DefaultServerPort)
	password := "virtdb"
	user := "virtdb"
	dbName := "test"
	connPool := client.NewPool(log.Debugf, minALive, maxAlive, maxIdle, mysqlAddress, user, password, dbName)

	monitor := performance.StartNewMonitorWithTimeUnit("proxy", time.Millisecond)
	runSelect(count, connPool, sql)
	monitor.End()
}

func testDirectMySQL(count int, sql string) {
	mysqlAddress := "127.0.0.1:3306"
	password := "root"
	user := "root"
	dbName := "test"
	connPool := client.NewPool(log.Debugf, minALive, maxAlive, maxIdle, mysqlAddress, user, password, dbName)

	monitor := performance.StartNewMonitorWithTimeUnit("MySQL", time.Millisecond)
	runSelect(count, connPool, sql)
	monitor.End()
}

const (
	RedisAddress  = "127.0.0.1:6379"
	RedisPoolSize = 10000
	RedisPassword = ""
)

func testDirectRedis(count int, sql string) {
	redisClient := virtdb.NewGenericRedisClientWithConfig(RedisAddress, RedisPoolSize, RedisPassword)
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
