package mysql

import (
	"code.hzmantu.com/dts/mysql/statement"
	"code.hzmantu.com/dts/structs"
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils"
	"code.hzmantu.com/dts/utils/tea"
	"errors"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/xelabs/go-mysqlstack/driver"
	"github.com/xelabs/go-mysqlstack/sqlparser/depends/sqltypes"
	"log"
	"syscall"
)

var (
	maxTryTime        = 1
	maxErrorSqlLength = 200
)

type Conn struct {
	Config  *structs.SourceConfig
	db      driver.Conn
	pool    *Pool
	tryTime int
}

func NewConn(pool *Pool) *Conn {
	conn := &Conn{
		pool:   pool,
		Config: pool.Config,
	}
	conn.connect()
	return conn
}

func (c *Conn) connect() {
	var err error
	c.db, err = driver.NewConn(
		c.pool.Config.User,
		c.pool.Config.Password,
		c.pool.Config.Address,
		"",
		c.pool.Config.Charset,
	)
	utils.PanicError(err)
}
func (c *Conn) Run(sql string) {
	if structs.TaskConfig.Debug {
		log.Println("Run:", sql)
	}
	err := c.db.Exec(sql)
	if c.retry(err, sql) {
		c.Run(sql)
		return
	}
}

func (c *Conn) Exec(sql string) error {
	if structs.TaskConfig.Debug {
		log.Println("Exec:", sql)
	}
	err := c.db.Exec(sql)
	if errors.Is(err, syscall.EPIPE) {
		c.db = c.pool.newConn().db
		return err
	}
	if c.retry(err, sql) {
		// 进行重试
		return c.Exec(sql)
	}
	c.Recycle()
	return err
}

func (c *Conn) ExecIgnore(sql string) {
	if structs.TaskConfig.Debug {
		log.Println("Exec:", sql)
	}
	err := c.db.Exec(sql)
	c.retry(err, sql)
	c.Recycle()
}

// Query the call needs to be put back into the connection after the use
func (c *Conn) Query(sql string) driver.Rows {
	if structs.TaskConfig.Debug {
		log.Println("Query:", sql)
	}
	rows, err := c.db.Query(sql)
	if c.retry(err, sql) {
		// 进行重试
		return c.Query(sql)
	}
	return rows
}

func (c *Conn) FetchAll(sql string) []map[string]*string {
	if structs.TaskConfig.Debug {
		log.Println("FetchAll:", sql)
	}

	result, err := c.db.FetchAll(sql, -1)
	if c.retry(err, sql) {
		// 进行重试
		return c.FetchAll(sql)
	}
	c.Recycle()

	var list []map[string]*string
	fields := result.Fields
	for _, rows := range result.Rows {
		hashMap := make(map[string]*string)
		for index, row := range rows {
			key := fields[index].Name
			if row.IsNull() {
				hashMap[key] = nil
			} else {
				hashMap[key] = tea.String(row.ToString())
			}
		}
		list = append(list, hashMap)
	}
	return list
}

func (c *Conn) ExecInsertData(table *db.Table, rows [][]sqltypes.Value) error {
	err := c.Exec(statement.InterAllData(table, rows))
	if e, ok := err.(*mysql.MyError); ok {
		if e.Code == mysql.ER_NET_PACKET_TOO_LARGE {
			i := len(rows) / 2
			_ = c.ExecInsertData(table, rows[:i])
			_ = c.ExecInsertData(table, rows[i:])
			return nil
		}
	}
	return err
}

func (c *Conn) ExecUpdateData(table *db.Table, rows [][]sqltypes.Value) error {
	err := c.Exec(statement.UpdateAllData(table, rows))
	if e, ok := err.(*mysql.MyError); ok {
		if e.Code == mysql.ER_NET_PACKET_TOO_LARGE {
			i := len(rows) / 2
			_ = c.ExecUpdateData(table, rows[:i])
			_ = c.ExecUpdateData(table, rows[i:])
			return nil
		}
	}
	return err
}

func (c *Conn) ExecDeleteData(table *db.Table, ids []string) error {
	err := c.Exec(statement.DeleteDataByIds(table, ids))
	if e, ok := err.(*mysql.MyError); ok {
		if e.Code == mysql.ER_NET_PACKET_TOO_LARGE {
			i := len(ids) / 2
			_ = c.ExecDeleteData(table, ids[:i])
			_ = c.ExecDeleteData(table, ids[i:])
			return nil
		}
	}
	return err
}

func (c *Conn) Close() error {
	return c.db.Close()
}

func (c *Conn) Recycle() {
	c.tryTime = 0
	c.pool.Put(c)
}

func (c *Conn) retry(err error, sql string) bool {
	if !utils.CheckError(err) {
		return false
	}
	if c.tryTime >= maxTryTime {
		length := len(sql)
		if length > maxErrorSqlLength {
			length = maxErrorSqlLength
		}
		log.Println(sql[:length])
		utils.PanicError(err)
	}
	c.tryTime++
	_ = c.Close()
	c.db = c.pool.newConn().db
	return true
}
