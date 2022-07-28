package mysql

import (
	"code.hzmantu.com/dts/mysql/statement"
	"code.hzmantu.com/dts/structs"
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils"
	"github.com/go-mysql-org/go-mysql/mysql"
	"strconv"
)

type Source struct {
	pool      *Pool
	Name      string
	Databases map[string]*db.Database
	Config    *structs.SourceConfig
	Storage   *utils.Storage
	Pos       *mysql.Position
	export    *Export
}

func NewSource(config *structs.SourceConfig, storage *utils.Storage) *Source {
	pool := NewPool(config)
	config.InitSource()

	source := &Source{
		pool:      pool,
		Name:      config.Name,
		Databases: make(map[string]*db.Database),
		Config:    config,
		Storage:   storage,
		export:    NewExport(pool),
	}

	databases := source.export.Databases()
	for _, database := range databases {
		source.Databases[database.Name] = database
	}

	return source
}

func (s *Source) GetConn() *Conn {
	return s.pool.Get()
}

func (s *Source) DisableKey() {
	s.pool.AppendInitSql(statement.DisableForeignKeyChecks())
}

func (s *Source) GetMasterPos() *mysql.Position {
	conn := s.GetConn()
	row := conn.Query("SHOW MASTER STATUS;")

	defer func() {
		_ = row.Close()
		conn.Recycle()
	}()

	if row.Next() == false {
		return nil
	}
	values, _ := row.RowValues()
	name := values[0].ToString()
	pos, _ := strconv.ParseInt(values[1].ToString(), 10, 64)
	return &mysql.Position{Name: name, Pos: uint32(pos)}
}
