package virtdb

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/server"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type VirtdbProvider struct {
	userPool sync.Map // username -> password
	username string
	password string
}

func (m *VirtdbProvider) CheckUsername(username string) (found bool, err error) {
	time.Sleep(time.Millisecond * time.Duration(50))
	m.username = username
	return true, nil
}

func (m *VirtdbProvider) GetCredential(username string) (password string, found bool, err error) {
	time.Sleep(time.Millisecond * time.Duration(50))
	m.password = password
	return password, true, nil
}

func (m *VirtdbProvider) AddUser(username, password string) {
	m.userPool.Store(username, password)
}

func Start(config *VirtdbConfig) {
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
	//		//log.Infof("Current Connection count:%v,aliveConnCount:%v", len(conns), aliveConnCount)
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
	log.Infof("Listening in %v", address)

	redisClient := NewGenericRedisClient(config)
	for {
		id := node.Generate().String()
		//log.Infof("id:[%v],%v\n", id, 1)
		c, netAcceptError := l.Accept()
		//log.Infof("id:[%v],%v\n", id, 2)

		if netAcceptError != nil {
			log.Errorf("accept net connection fail.%v", netAcceptError)
			//log.Infof("id:[%v],%v\n", id, 3)
			continue
		}

		go func() {
			ctx := context.WithValue(context.Background(), "id", id)
			conn, err := client.Connect(config.MySQLAddress, config.MySQLUser, config.MySQLPassword, config.ServerDBName)
			if err != nil {
				return
			}
			defer func(conn *client.Conn) {
				_ = conn.Close()
			}(conn)
			handler := &CustomMySQLHandler{ctx: ctx, conn: conn, redisClient: redisClient}

			//log.Infof("id:[%v],%v\n", id, 4)
			serverConn, err := server.NewConn(c, config.ServerUser, config.ServerPassword, handler)
			//log.Infof("id:[%v],%v\n", id, 5)

			if err != nil {
				//log.Infof("id:[%v],%v\n", id, 6)
				log.Errorln(err)
				return
			}
			//log.Infof("id:[%v],%v\n", id, 7)

			conns = append(conns, serverConn)
			for {
				//log.Infof("id:[%v],%v\n", id, 8)
				err := serverConn.HandleCommand()
				//log.Infof("id:[%v],%v\n", id, 9)
				if err != nil {
					// log.Errorf("handle serverConn error,closed:%v,%v", serverConn.Closed(), err)
				}
				if serverConn.Closed() {
					break
				}
				//log.Infof("id:[%v],%v\n", id, 10)
			}
		}()
	}
}
