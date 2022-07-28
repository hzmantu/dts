package compare

import (
	"code.hzmantu.com/dts/mysql"
	"code.hzmantu.com/dts/mysql/statement"
	"code.hzmantu.com/dts/structs/db"
	"log"
)

type Compare struct {
	from        *mysql.Source
	to          *mysql.Source
	step        int64
	ExportTable chan *db.Table
	sql         []string
}

func NewCompare(from, to *mysql.Source, step int64) *Compare {
	return &Compare{
		from:        from,
		to:          to,
		step:        step,
		ExportTable: make(chan *db.Table, 999),
	}
}

func (c *Compare) Start() {
	for _, database := range c.from.Databases {
		toDatabase, exist := c.to.Databases[database.Name]
		if !exist {
			log.Printf("the database %s does not exist \n", database.Name)
			_ = c.to.GetConn().Exec(statement.CreateDatabase(database.Name))
			toDatabase = db.NewDatabase(c.to.Name, database.Name, []*db.Table{})
		}
		c.compareTable(database.Tables, toDatabase.Tables)
	}
	conn := c.to.GetConn()
	conn.Run(statement.EnableForeignKeyChecks())
	for _, sql := range c.sql {
		conn.Run(sql)
	}
	_ = conn.Exec(statement.DisableForeignKeyChecks())
	close(c.ExportTable)
	log.Printf("completion of %s database data", c.from.Name)
}

func (c *Compare) AppendSql(sql string) {
	c.sql = append(c.sql, sql)
}
