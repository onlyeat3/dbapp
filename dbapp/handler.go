package dbapp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-redis/redis/v8"
	"github.com/siddontang/go-log/log"
	"github.com/xwb1989/sqlparser"
	"strings"
	"time"
)

type CustomMySQLHandler struct {
	ctx         context.Context
	connPool    *client.Pool
	redisClient *redis.Client
	dbName      string
}

func (h CustomMySQLHandler) GetDBConn() (*client.Conn, error) {
	log.Infof("id:%v,handler%v", h.ctx.Value("id"), 1)
	conn, err := h.connPool.GetConn(h.ctx)
	log.Infof("id:%v,handler%v", h.ctx.Value("id"), 2)
	if err != nil {
		return nil, err
	}
	log.Infof("id:%v,handler%v", h.ctx.Value("id"), 3)
	if h.dbName != "" {
		err := conn.UseDB(h.dbName)
		log.Infof("id:%v,handler%v", h.ctx.Value("id"), 4)
		if err != nil {
			log.Infof("id:%v,handler%v", h.ctx.Value("id"), 5)
			return nil, err
		}
	}
	log.Infof("id:%v,handler%v", h.ctx.Value("id"), 6)
	return conn, err
}

func (h CustomMySQLHandler) ReturnDBConn(conn *client.Conn) {
	conn.UseDB("")
	h.connPool.PutConn(conn)
	log.Infof("id:%v,handler%v", h.ctx.Value("id"), 12)
}

func (h CustomMySQLHandler) GetRedisConn() *redis.Conn {
	return h.redisClient.Conn(h.ctx)
}

func (h *CustomMySQLHandler) UseDB(dbName string) error {
	conn, err := h.GetDBConn()
	if err != nil {
		return err
	}
	defer h.ReturnDBConn(conn)
	err = conn.UseDB(dbName)
	log.Infof("id:%v,handler%v", h.ctx.Value("id"), 11)
	if err == nil {
		h.dbName = dbName
	}
	return err
}

func (h *CustomMySQLHandler) handleQuery(query string, binary bool) (*mysql.Result, error) {
	//log.Infoln("sql", query)
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		// Do something with the err
		log.Warnln("sql parse fail", err)
	}
	dbConn, err := h.GetDBConn()
	if err != nil {
		return nil, err
	}
	defer h.ReturnDBConn(dbConn)
	log.Infof("id:%v,handler%v", h.ctx.Value("id"), 7)
	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		useCache := strings.ToLower(strings.TrimSpace(stmt.Cache)) == "sql_cache"
		isRedisValid := true
		if useCache {
			exists, _ := h.redisClient.Exists(h.ctx, query).Result()

			if exists == 1 {
				redisResult, redisGetErr := h.redisClient.Get(h.ctx, query).Bytes()
				if redisGetErr != nil {
					log.Errorln(redisGetErr)
					isRedisValid = false
				} else {
					r := &mysql.Result{}
					err := json.Unmarshal(redisResult, r)
					if err != nil {
						log.Errorln(err)
					} else {
						return r, nil
					}
				}
			}
		}
		dbResult, error := dbConn.Execute(query)
		if useCache && isRedisValid {
			if dbResult != nil && len(dbResult.RowDatas) > 0 {
				encoded, err := json.Marshal(dbResult)

				statusCmd := h.redisClient.Set(h.ctx, query, encoded, time.Second*60)
				if statusCmd.Err() != nil {
					fmt.Errorf("error:%v", err)
					return nil, err
				}
			}
		}
		log.Infof("id:%v,handler%v", h.ctx.Value("id"), 8)
		return dbResult, error
	default:
		dbResult, error := dbConn.Execute(query)
		log.Infof("id:%v,handler%v", h.ctx.Value("id"), 9)
		return dbResult, error
	}
}

func (h *CustomMySQLHandler) HandleQuery(query string) (*mysql.Result, error) {
	return h.handleQuery(query, false)
}

func (h *CustomMySQLHandler) HandleFieldList(table string, fieldWildcard string) ([]*mysql.Field, error) {
	return nil, nil
}

func (h *CustomMySQLHandler) HandleStmtPrepare(sql string) (params int, columns int, ctx interface{}, err error) {
	ss := strings.Split(sql, " ")
	switch strings.ToLower(ss[0]) {
	case "select":
		params = 1
		columns = 2
	case "insert":
		params = 2
		columns = 0
	case "replace":
		params = 2
		columns = 0
	case "update":
		params = 1
		columns = 0
	case "delete":
		params = 1
		columns = 0
	default:
		err = fmt.Errorf("invalid prepare %s", sql)
	}
	return params, columns, nil, err
}

func (h *CustomMySQLHandler) HandleStmtClose(context interface{}) error {
	return nil
}

func (h *CustomMySQLHandler) HandleStmtExecute(ctx interface{}, query string, args []interface{}) (*mysql.Result, error) {
	return h.handleQuery(query, true)
}

func (h *CustomMySQLHandler) HandleOtherCommand(cmd byte, data []byte) error {
	return mysql.NewError(mysql.ER_UNKNOWN_ERROR, fmt.Sprintf("command %d is not supported now", cmd))
}
