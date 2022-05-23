package dbapp

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/siddontang/go-log/log"
	"strings"
	"time"
)

type CustomMySQLHandler struct {
	ctx         context.Context
	connPool    *client.Pool
	redisClient *GenericRedisClient
	dbName      string
}

func (h CustomMySQLHandler) GetDBConn() (*client.Conn, error) {
	conn, err := h.connPool.GetConn(h.ctx)
	if err != nil {
		return nil, err
	}
	if h.dbName != "" && conn.GetDB() != h.dbName {
		err := conn.UseDB(h.dbName)
		if err != nil {
			return nil, err
		}
	}
	return conn, err
}

func (h CustomMySQLHandler) ReturnDBConn(conn *client.Conn) {
	h.connPool.PutConn(conn)
}

func (h *CustomMySQLHandler) UseDB(dbName string) error {
	conn, err := h.GetDBConn()
	if err != nil {
		return err
	}
	defer h.ReturnDBConn(conn)
	err = conn.UseDB(dbName)
	//log.Infof("id:%v,handler%v", h.ctx.Value("id"), 11)
	if err == nil {
		h.dbName = dbName
	}
	return err
}

func (h *CustomMySQLHandler) handleQuery(query string, binary bool) (*mysql.Result, error) {
	//log.Infoln("sql", query)
	//stmt, err := sqlparser.Parse(query)
	//if err != nil {
	//	Do something with the err
	//log.Warnln("sql parse fail", err)
	//}
	//log.Infof("id:%v,handler%v", h.ctx.Value("id"), 7)
	ss := strings.Split(query, " ")
	switch strings.ToLower(ss[0]) {
	case "select":
		useCache := strings.Contains(strings.ToLower(strings.TrimSpace(query)), "sql_cache")
		isRedisValid := true
		if useCache {
			redisResult, redisGetErr := h.redisClient.Get(h.ctx, query).Bytes()
			if redisGetErr != nil && redisGetErr.Error() != "redis: nil" {
				log.Errorln(redisGetErr)
				isRedisValid = false
			} else {
				if redisResult != nil {
					r := &mysql.Result{}
					decoder := gob.NewDecoder(bytes.NewReader(redisResult))
					err := decoder.Decode(r)
					if err != nil {
						log.Errorln(err)
					} else {
						return r, nil
					}
				}
			}
		}

		dbConn, err := h.GetDBConn()
		if err != nil {
			return nil, err
		}
		defer h.ReturnDBConn(dbConn)
		dbResult, error := dbConn.Execute(query)
		if useCache && isRedisValid {
			if dbResult != nil && len(dbResult.RowDatas) > 0 {
				var buff bytes.Buffer
				encoder := gob.NewEncoder(&buff)
				err := encoder.Encode(dbResult)
				if err != nil {
					log.Errorf("encode fail", err)
				}

				statusCmd := h.redisClient.Set(h.ctx, query, buff.Bytes(), time.Second*60)
				if statusCmd.Err() != nil {
					log.Errorf("error:%v", err)
					return nil, err
				}
			}
		}
		//log.Infof("id:%v,handler%v", h.ctx.Value("id"), 8)

		return dbResult, error
	default:
		dbConn, err := h.GetDBConn()
		if err != nil {
			return nil, err
		}
		defer h.ReturnDBConn(dbConn)
		dbResult, error := dbConn.Execute(query)
		//log.Infof("id:%v,handler%v", h.ctx.Value("id"), 9)
		return dbResult, error
	}

	return nil, nil
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
