package dbapp

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"net"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/server"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/siddontang/go-log/log"
)

type DBAppProvider struct {
	userPool sync.Map // username -> password
	username string
	password string
}

func (m *DBAppProvider) CheckUsername(username string) (found bool, err error) {
	time.Sleep(time.Millisecond * time.Duration(50))
	m.username = username
	return true, nil
}

func (m *DBAppProvider) GetCredential(username string) (password string, found bool, err error) {
	time.Sleep(time.Millisecond * time.Duration(50))
	m.password = password
	return password, true, nil
}

func (m *DBAppProvider) AddUser(username, password string) {
	m.userPool.Store(username, password)
}

func Start(config *DBAppConfig) {
	conns := make([]*server.Conn, 0)
	//go func() {
	//	for {
	//		time.Sleep(time.Second)
	//		aliveConnCount := 0
	//		for _, conn := range conns {
	//			if !conn.Closed() {
	//				aliveConnCount++
	//			}
	//		}
	//		log.Infof("Current Connection count:%v,aliveConnCount:%v", len(conns), aliveConnCount)
	//	}
	//}()
	// Create a new Node with a Node number of 1
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	address := fmt.Sprintf("0.0.0.0:%v", config.ServerPort)
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorln(err)
		return
	}

	redisClient := redis.NewClient(&redis.Options{Addr: config.RedisAddress, PoolSize: config.RedisPoolSize})
	connPool := client.NewPool(log.Debugf, config.MySQLConnPoolMinALive, config.MySQLConnPoolMaxAlive, config.MySQLConnPoolMaxIdle, config.MySQLAddress, config.MySQLUser, config.MySQLPassword, "")
	for {
		id := node.Generate().String()
		log.Infof("id:[%v],%v\n", id, 1)
		c, netAcceptError := l.Accept()
		log.Infof("id:[%v],%v\n", id, 2)

		if netAcceptError != nil {
			log.Errorf("accept net connection fail.%v", netAcceptError)
			log.Infof("id:[%v],%v\n", id, 3)
			continue
		}

		go func() {

			ctx := context.WithValue(context.Background(), "id", id)
			handler := &CustomMySQLHandler{ctx: ctx, connPool: connPool, redisClient: redisClient}

			log.Infof("id:[%v],%v\n", id, 4)
			serverConn, err := server.NewConn(c, config.ServerUser, config.ServerPassword, handler)
			log.Infof("id:[%v],%v\n", id, 5)

			if err != nil {
				log.Infof("id:[%v],%v\n", id, 6)
				log.Errorln(err)
				return
			}
			log.Infof("id:[%v],%v\n", id, 7)

			conns = append(conns, serverConn)
			for {
				log.Infof("id:[%v],%v\n", id, 8)
				err := serverConn.HandleCommand()
				log.Infof("id:[%v],%v\n", id, 9)
				if err != nil {
					// log.Errorf("handle serverConn error,closed:%v,%v", serverConn.Closed(), err)
				}
				if serverConn.Closed() {
					break
				}
				log.Infof("id:[%v],%v\n", id, 10)
			}
		}()
	}

}

func logN(i int) {

}
