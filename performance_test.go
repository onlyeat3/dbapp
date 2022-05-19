package main

import (
	"context"
	"fmt"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/siddontang/go-log/log"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestProxy(t *testing.T) {
	//count := 100
	count := 1000
	mysqlAddress := "127.0.0.1:4000"
	runSelect(count, mysqlAddress)
	log.Infof("end func")
}

//func BenchmarkMySQL(b *testing.B) {
//	//count := 100
//	count := b.N
//	//fmt.Printf("count:%v,b.N:%v\n", count, b.N)
//	mysqlAddress := "127.0.0.1:3306"
//	runSelect(count, mysqlAddress)
//}

func runSelect(count int, mysqlAddress string) {
	password := "root"
	user := "root"

	dbName := "test"
	minALive := 100
	maxAlive := 1000
	maxIdle := 120
	connPool := client.NewPool(func(format string, args ...interface{}) {
		fmt.Printf(format, args)
	}, minALive, maxAlive, maxIdle, mysqlAddress, user, password, dbName)
	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)
	for i := 0; i < count; i++ {
		g.Go(func() error {
			conn, err := connPool.GetConn(ctx)
			fmt.Printf("conn:%v,err:%v\n", conn, err)
			if err != nil {
				fmt.Println("get conn fail", err)
				return err
			}
			_, e := conn.Execute("select * from trade limit 2")
			if e != nil {
				log.Errorf("%v", e)
				return e
			}
			fmt.Println("1")
			conn.Close()
			connPool.PutConn(conn)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("end.")
}
