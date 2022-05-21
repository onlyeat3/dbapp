package dbapp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-redis/redis/v8"
	"github.com/siddontang/go-log/log"
	"strings"
	"time"
)

type CustomMySQLHandler struct {
	ctx       context.Context
	dbConn    *client.Conn
	redisConn *redis.Conn
}

func (h CustomMySQLHandler) DelayExecute(command string) (*mysql.Result, error) {
	time.Sleep(time.Millisecond * 10)
	return h.dbConn.Execute(command)
}

func (h *CustomMySQLHandler) UseDB(dbName string) error {
	_, error := h.dbConn.Execute("use " + dbName)
	return error
}

func (h *CustomMySQLHandler) handleQuery(query string, binary bool) (*mysql.Result, error) {
	log.Infoln("sql", query)
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		// Do something with the err
		log.Warnln("sql parse fail", err)
	}
	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		useCache := strings.TrimSpace(stmt.Cache) == "sql_cache"
		if useCache {
			redisResult, redisGetErr := h.redisConn.Get(h.ctx, query).Bytes()
			if redisGetErr == nil {
				r := &mysql.Result{}
				err := json.Unmarshal(redisResult, r)
				if err != nil {
					return nil, err
				}
				return r, nil
			}
		}
		dbResult, error := h.dbConn.Execute(query)
		if useCache {
			if dbResult != nil && len(dbResult.RowDatas) > 0 {
				encoded, err := json.Marshal(dbResult)

				statusCmd := h.redisConn.Set(h.ctx, query, encoded, time.Second*60)
				if statusCmd.Err() != nil {
					fmt.Errorf("error:%v", err)
				}
			}
		}
		if error != nil {
			return nil, error
		}
		return dbResult, nil
	default:
		dbResult, error := h.dbConn.Execute(query)
		if error != nil {
			return nil, error
		}
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
