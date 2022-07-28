package compare

import (
	"code.hzmantu.com/dts/mysql"
	"code.hzmantu.com/dts/mysql/statement"
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils/report"
	"log"
	"sync"
)

func (c *Compare) compareTable(fromTables, toTables map[string]*db.Table) {
	for _, table := range fromTables {
		toTable, exist := toTables[table.Name]
		if !exist {
			_ = c.to.GetConn().Exec(table.GetCreateSql())
			for _, constraint := range table.Constraints {
				c.AppendSql(statement.CreateConstraint(table, constraint))
			}
			c.ExportTable <- table
			continue
		}

		log.Printf("comparing `%s`.`%s` data table structure...", table.Database, table.Name)
		c.compareTableConstraints(table, toTable)
		c.compareTableKeys(table, toTable)
		c.compareTableColumns(table, toTable)

		toTable.DisableMask()
		toTable.NewColumns(table.Columns, table.GetFields())
		if table.GetPrimaryKey() != "" && table.AutoIncrement < toTable.AutoIncrement {
			toTable.AppendToBeSql(statement.AlterAutoIncrement(table, table.AutoIncrement))
		}

		c.compareTableData(table, toTable)
	}
}

func (c *Compare) compareTableColumns(table, toTable *db.Table) {
	toColumns := toTable.Columns
	fromColumns := table.Columns

	for field, column := range fromColumns {
		toColumn, exist := toColumns[field]
		if !exist {
			_ = c.to.GetConn().Exec(statement.AddColumn(table, column))
		} else if column.GetSql() != toColumn.GetSql() {
			log.Printf("difference field, original: %s, current: %s", column.GetSql(), toColumn.GetSql())
			_ = c.to.GetConn().Exec(statement.ChangeColumn(table, column))
		}
	}

	for field := range toColumns {
		if _, exist := fromColumns[field]; !exist {
			_ = c.to.GetConn().Exec(statement.DropColumn(table, field))
		}
	}
}

func (c *Compare) compareTableKeys(table, toTable *db.Table) {
	toKeys := toTable.Keys
	fromKeys := table.Keys
	for name, key := range fromKeys {
		toKey, exist := toKeys[name]
		if !exist {
			toTable.AppendToBeSql(statement.CreateIndex(table, key))
		} else if key.GetSql() != toKey.GetSql() {
			log.Printf("difference index，original：%s , current：%s", key.GetSql(), toKey.GetSql())
			c.to.GetConn().ExecIgnore(statement.DropIndex(table, name))
			toTable.AppendToBeSql(statement.CreateIndex(table, key))
		}
	}

	for field := range toKeys {
		if _, exist := fromKeys[field]; !exist {
			c.to.GetConn().ExecIgnore(statement.DropIndex(table, field))
		}
	}
}

func (c *Compare) compareTableConstraints(table, toTable *db.Table) {
	fromConstraints := table.Constraints
	toConstraints := toTable.Constraints
	for name, constraint := range fromConstraints {
		toConstraint, exist := toConstraints[name]
		if !exist {
			c.AppendSql(statement.CreateConstraint(table, constraint))
		} else if constraint.GetSql() != toConstraint.GetSql() {
			log.Printf("difference constraint, original: %s, current: %s", constraint.GetSql(), toConstraint.GetSql())
			c.to.GetConn().ExecIgnore(statement.DropConstraint(table, name))
			c.AppendSql(statement.CreateConstraint(table, constraint))
		}
	}
	for name := range toConstraints {
		if _, exist := fromConstraints[name]; !exist {
			c.to.GetConn().ExecIgnore(statement.DropConstraint(table, name))
		}
	}
}

func (c *Compare) compareTableData(fromTable, toTable *db.Table) {
	if c.to.Storage.IsFinishTable(fromTable) {
		for _, sql := range toTable.GetToBeSql() {
			_ = c.to.GetConn().Exec(sql)
		}
		return
	}
	if fromTable.GetMaxId() == 0 {
		_ = c.to.GetConn().Exec(statement.TruncateTable(fromTable))
		return
	}
	if fromTable.GetPrimaryKey() == "" || toTable.GetMaxId() == 0 {
		c.ExportTable <- fromTable
		return
	}
	log.Printf("comparing `%s`.`%s` table data...", fromTable.Database, fromTable.Name)
	// clear redundant data
	_ = c.to.GetConn().Exec(statement.DeleteRedundantData(fromTable))
	// wait group
	var wg sync.WaitGroup
	var fromRows, toRows *mysql.Rows
	for i := fromTable.GetMinId(); i <= fromTable.GetMaxId(); i += c.step {
		if i > toTable.GetMaxId() {
			fromTable.SaveDataInterval(i, fromTable.GetMaxId())
			c.ExportTable <- fromTable
			return
		}
		wg.Add(1)
		go func() {
			fromRows = c.from.GetConn().GetTableRows(fromTable, i, c.step)
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			toRows = c.to.GetConn().GetTableRows(toTable, i, c.step)
			wg.Done()
		}()
		wg.Wait()
		insert, update, deleteIds := c.compareRows(fromRows, toRows, fromTable)
		if len(deleteIds) > 0 {
			report.CompareCount.
				WithLabelValues(fromTable.Database, fromTable.Name, "delete").
				Add(float64(len(deleteIds)))
			_ = c.to.GetConn().ExecDeleteData(fromTable, deleteIds)
		}
		if len(update) > 0 {
			report.CompareCount.
				WithLabelValues(fromTable.Database, fromTable.Name, "update").
				Add(float64(len(update)))
			_ = c.to.GetConn().ExecUpdateData(fromTable, update)
		}
		if len(insert) > 0 {
			report.CompareCount.
				WithLabelValues(fromTable.Database, fromTable.Name, "insert").
				Add(float64(len(insert)))
			_ = c.to.GetConn().ExecInsertData(fromTable, insert)
		}
	}
	// execute the remaining sql
	for _, sql := range toTable.GetToBeSql() {
		_ = c.to.GetConn().Exec(sql)
	}
	// save cache
	c.to.Storage.SaveFinishTable(fromTable)
}
