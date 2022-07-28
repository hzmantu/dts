package mysql

import (
	"code.hzmantu.com/dts/structs"
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils/transform"
	"fmt"
	"log"
	"strings"
)

var innerDatabases = []string{"information_schema", "mysql", "performance_schema", "sys"}

type Export struct {
	pool        *Pool
	config      *structs.SourceConfig
	constraints map[string][]*db.Constraint
}

func NewExport(pool *Pool) *Export {
	return &Export{
		pool:        pool,
		config:      pool.Config,
		constraints: make(map[string][]*db.Constraint),
	}
}

func (e *Export) Databases() []*db.Database {
	var list []*db.Database
	databases := e.getDatabases()
	e.initConstraints()
	for _, name := range databases {
		log.Printf("%s is collecting %s database corresponding table structure data...\n", e.config.Name, name)
		database := db.NewDatabase(e.config.Name, name, e.exportTables(name))
		list = append(list, database)
	}
	return list
}

func (e *Export) getDatabases() []string {
	sql := "SHOW DATABASES"
	var databases, where []string
	if onlyDatabases := e.config.GetOnlyDatabases(); len(onlyDatabases) > 0 {
		where = append(where, fmt.Sprintf(
			"`database` IN ('%s')",
			strings.Join(onlyDatabases, "', '"),
		))
	}
	if excludeDatabases := e.config.GetExcludeDatabases(); len(excludeDatabases) > 0 {
		where = append(where, fmt.Sprintf(
			"`database` NOT IN ('%s')",
			strings.Join(excludeDatabases, "', '"),
		))
	}
	where = append(where, fmt.Sprintf(
		"`database` NOT IN ('%s')",
		strings.Join(innerDatabases, "', '"),
	))
	if len(where) > 0 {
		sql += " WHERE " + strings.Join(where, " AND ")
	}

	result := e.getConn().FetchAll(sql)
	for _, data := range result {
		databases = append(databases, *data["Database"])
	}
	return databases
}

func (e *Export) getTables(databaseName string) []*db.Table {
	var tables []*db.Table
	var where []string
	sql := fmt.Sprintf("SHOW TABLE STATUS FROM `%s`", databaseName)
	where = append(where, "Engine IS NOT NULL")

	if list := e.config.GetOnlyTables(databaseName); len(list) > 0 {
		where = append(where, fmt.Sprintf(
			"`Name` IN ('%s')",
			strings.Join(list, "', '"),
		))
	}

	if list := e.config.GetExcludeTables(databaseName); len(list) > 0 {
		where = append(where, fmt.Sprintf(
			"`Name` NOT IN ('%s')",
			strings.Join(list, "', '"),
		))
	}

	if len(where) > 0 {
		sql += " WHERE " + strings.Join(where, " AND ")
	}

	result := e.getConn().FetchAll(sql)

	for _, data := range result {
		var autoIncrement int64
		if data["Auto_increment"] != nil {
			autoIncrement = transform.StringToInt64(*data["Auto_increment"])
		}
		tables = append(tables, &db.Table{
			Database:      databaseName,
			Name:          *data["Name"],
			Engine:        *data["Engine"],
			AutoIncrement: autoIncrement,
			Collation:     *data["Collation"],
			CreateOptions: *data["Create_options"],
		})
	}
	return tables
}

func (e *Export) getIntervalId(table *db.Table) (int64, int64) {
	var field string
	query := "SELECT %s FROM `%s`.`%s`;"
	if table.GetPrimaryKey() != "" {
		field = fmt.Sprintf("MIN(`%s`), MAX(`%s`)", table.GetPrimaryKey(), table.GetPrimaryKey())
	} else {
		field = "0, COUNT(1)"
	}
	query = fmt.Sprintf(query, field, table.Database, table.Name)
	conn := e.getConn()
	row := conn.Query(query)
	defer func() {
		_ = row.Close()
		conn.Recycle()
	}()
	if row.Next() == false {
		return 0, 0
	}
	values, _ := row.RowValues()
	return transform.StringToInt64(values[0].ToString()), transform.StringToInt64(values[1].ToString())
}

func (e *Export) getTableColumns(table *db.Table) []*db.Column {
	var columns []*db.Column
	sql := "SHOW FULL COLUMNS FROM `%s`.`%s`;"
	result := e.getConn().FetchAll(fmt.Sprintf(sql, table.Database, table.Name))
	for _, data := range result {
		columns = append(columns, &db.Column{
			Field:      *data["Field"],
			Type:       *data["Type"],
			Collation:  data["Collation"],
			Null:       *data["Null"],
			Key:        *data["Key"],
			Default:    data["Default"],
			Extra:      *data["Extra"],
			Privileges: *data["Privileges"],
			Comment:    *data["Comment"],
		})
	}
	return columns
}

func (e *Export) getTableIndex(table *db.Table) []*db.Key {
	var keys []*db.Key
	sql := "SHOW INDEX FROM `%s`.`%s`;"
	result := e.getConn().FetchAll(fmt.Sprintf(sql, table.Database, table.Name))
	for _, data := range result {
		keys = append(keys, &db.Key{
			Table:      *data["Table"],
			NonUnique:  transform.StringToInt(*data["Non_unique"]),
			KeyName:    strings.Replace(*data["Key_name"], "`", "", -1),
			SeqInIndex: transform.StringToInt(*data["Seq_in_index"]),
			ColumnName: *data["Column_name"],
		})
	}
	return keys
}

func (e *Export) initConstraints() {
	result := e.getConn().FetchAll("SELECT U.*, R.UPDATE_RULE, R.DELETE_RULE " +
		"FROM information_schema.REFERENTIAL_CONSTRAINTS R " +
		"LEFT JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE U " +
		"ON U.CONSTRAINT_NAME = R.CONSTRAINT_NAME AND R.CONSTRAINT_SCHEMA = U.CONSTRAINT_SCHEMA")
	for _, data := range result {
		key := fmt.Sprintf("%s.%s", *data["TABLE_SCHEMA"], *data["TABLE_NAME"])
		e.constraints[key] = append(e.constraints[key], &db.Constraint{
			Name:                  *data["CONSTRAINT_NAME"],
			ColumnName:            *data["COLUMN_NAME"],
			ReferencedTableSchema: *data["REFERENCED_TABLE_SCHEMA"],
			ReferencedColumnName:  *data["REFERENCED_COLUMN_NAME"],
			ReferencedTableName:   *data["REFERENCED_TABLE_NAME"],
			UpdateRule:            *data["UPDATE_RULE"],
			DeleteRule:            *data["DELETE_RULE"],
		})
	}
}

func (e *Export) getTableConstraints(table *db.Table) []*db.Constraint {
	key := fmt.Sprintf("%s.%s", table.Database, table.Name)
	if constraints, exist := e.constraints[key]; exist {
		return constraints
	}
	return []*db.Constraint{}
}

func (e *Export) exportTables(databaseName string) []*db.Table {
	tables := e.getTables(databaseName)
	for _, table := range tables {
		columns := e.getTableColumns(table)
		keys := e.getTableIndex(table)
		constraints := e.getTableConstraints(table)
		masks := getFieldMasks(table.Database, table.Name)
		table.AppendElement(columns, keys, constraints, masks)
		minId, maxId := e.getIntervalId(table)
		table.SaveDataInterval(minId, maxId)
		log.Printf(
			"%s table：%s.%s, id: %d, maxId: %d",
			e.config.Name, table.Database, table.Name,
			table.AutoIncrement, maxId,
		)
	}
	return tables
}

func (e *Export) getConn() *Conn {
	return e.pool.Get()
}

func getFieldMasks(databaseName string, tableName string) map[string]string {
	filters := structs.TaskConfig.Filters
	if masks, exists := filters[databaseName+"."+tableName]; exists {
		return masks
	}
	return nil
}
