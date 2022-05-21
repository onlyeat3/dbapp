package dbapp

import (
	"context"
	"fmt"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/server"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/siddontang/go-log/log"
	"net"
	"sync"
	"time"
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
	address := fmt.Sprintf("0.0.0.0:%v", config.ServerPort)
	l, _ := net.Listen("tcp", address)

	redisClient := redis.NewClient(&redis.Options{Addr: config.RedisAddress, PoolSize: config.RedisPoolSize})

	for {
		c, netAcceptError := l.Accept()
		if netAcceptError != nil {
			log.Errorf("accept net connection fail.%v", netAcceptError)
			continue
		}
		ctx := context.Background()

		svr := server.NewDefaultServer()

		dbAppProvider := &DBAppProvider{}
		redisConn := redisClient.Conn(ctx)
		handler := &CustomMySQLHandler{ctx: ctx, dbConn: nil, redisConn: redisConn}

		serverConn, err := server.NewCustomizedConn(c, svr, dbAppProvider, handler)

		user := dbAppProvider.username
		password := dbAppProvider.password

		connPool := client.NewPool(log.Debugf, config.MySQLConnPoolMinALive, config.MySQLConnPoolMaxAlive, config.MySQLConnPoolMaxIdle, config.MySQLConnPoolAddress, user, password, "")
		clientConn, connGetPoolError := connPool.GetConn(ctx)
		if connGetPoolError != nil {
			err := c.Close()
			log.Errorf("faild to getConn %v", err)
			if err != nil {
				continue
			}
		}

		if err != nil {
			log.Errorln(err)
			continue
		}
		go func() {
			for serverConn != nil && !serverConn.Closed() {
				err := serverConn.HandleCommand()
				if err != nil {
					log.Errorf("close check error,%v", err)
				}
			}
			defer func() {
				clientConn.Close()
				//connPool.PutConn(clientConn)
				redisConn.Close()
				for serverConn != nil && !serverConn.Closed() {
					serverConn.Close()
				}
				if c != nil {
					c.Close()
				}
			}()
		}()
	}
}
