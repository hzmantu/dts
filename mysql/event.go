package mysql

import (
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils"
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"strings"
)

type Event struct {
	table   *db.Table
	event   *canal.RowsEvent
	columns []string
}

func NewEvent(event *canal.RowsEvent, table *db.Table) *Event {
	var columns []string
	for _, TableColumn := range event.Table.Columns {
		columns = append(columns, TableColumn.Name)
	}
	return &Event{event: event, table: table, columns: columns}
}

func (e *Event) getPkColumns() []string {
	var pkColumns []string
	for _, idx := range e.event.Table.PKColumns {
		column := e.event.Table.GetPKColumn(idx)
		pkColumns = append(pkColumns, column.Name)
	}
	return pkColumns
}

func (e *Event) getColumnValueByIdx(value interface{}, idx int) string {
	if e.table.HashMask() && len(e.columns) > idx {
		return e.getColumnValue(value, e.columns[idx])
	}

	return getValueInsertString(value)
}

func (e *Event) getColumnValue(value interface{}, columnName string) string {
	if e.table.HashMask() {
		column, exist := e.table.Columns[columnName]
		if exist && column.Mask != nil && value != nil {
			value = utils.CallString(*column.Mask, fmt.Sprintf("%s", value))
		}
	}

	return getValueInsertString(value)
}

func (e *Event) hasPkColumns() bool {
	return len(e.event.Table.PKColumns) == 0
}

func (e *Event) GetSql() string {
	switch e.event.Action {
	case "insert":
		return e.InterAllData()
	case "update":
		if e.hasPkColumns() {
			return e.updateNotPkStatement()
		}
		return e.updatePkStatement()
	case "delete":
		if e.hasPkColumns() {
			return e.deleteNotPkStatement()
		}
		return e.deletePkStatement()
	}
	return ""
}

func (e *Event) InterAllData() string {
	var insertValues []string
	for _, row := range e.event.Rows {
		var values []string
		for idx, value := range row {
			values = append(values, e.getColumnValueByIdx(value, idx))
		}
		insertValues = append(insertValues, strings.Join(values, ", "))
	}

	return fmt.Sprintf(
		"INSERT IGNORE INTO `%s`.`%s` (`%s`) VALUES (%s);",
		e.event.Table.Schema, e.event.Table.Name,
		strings.Join(e.columns, "`, `"),
		strings.Join(insertValues, "), ("),
	)
}

func (e *Event) updatePkStatement() string {
	pkColumns := e.getPkColumns()

	var sql string
	for i := 0; i < len(e.event.Rows); i += 2 {
		var wheres []string
		for idx, column := range pkColumns {
			oldValue := e.event.Rows[i][idx]
			wheres = append(wheres, fmt.Sprintf("`%s` = %s", column, e.getColumnValue(oldValue, column)))
		}
		var values []string
		for idx, column := range e.columns {
			oldValue := e.getColumnValue(e.event.Rows[i][idx], column)
			newValue := e.getColumnValue(e.event.Rows[i+1][idx], column)
			if oldValue != newValue {
				values = append(values, fmt.Sprintf("`%s` = %s", column, newValue))
			}
		}
		if len(values) == 0 {
			continue
		}

		sql += fmt.Sprintf(
			"UPDATE `%s`.`%s` SET %s WHERE %s ;",
			e.event.Table.Schema, e.event.Table.Name,
			strings.Join(values, ","),
			strings.Join(wheres, " AND "),
		)
	}
	return sql
}

func (e *Event) updateNotPkStatement() string {
	var sql string

	for i := 0; i < len(e.event.Rows); i += 2 {
		var wheres []string
		var values []string
		for idx, column := range e.columns {
			oldValue := e.getColumnValue(e.event.Rows[i][idx], column)
			wheres = append(wheres, fmt.Sprintf("`%s` = %s", column, oldValue))
			newValue := e.getColumnValue(e.event.Rows[i+1][idx], column)
			if oldValue != newValue {
				values = append(values, fmt.Sprintf("`%s` = %s", column, newValue))
			}
		}
		if len(values) == 0 {
			continue
		}

		sql += fmt.Sprintf(
			"UPDATE `%s`.`%s` SET %s WHERE %s ;",
			e.event.Table.Schema, e.event.Table.Name,
			strings.Join(values, ","),
			strings.Join(wheres, " AND "),
		)
	}
	return sql
}

func (e *Event) deletePkStatement() string {
	pkColumns := e.getPkColumns()

	var sql string
	for _, rows := range e.event.Rows {
		var wheres []string
		values, _ := e.event.Table.GetPKValues(rows)
		for i, val := range values {
			wheres = append(wheres, fmt.Sprintf("`%s` = %s", pkColumns[i], getValueInsertString(val)))
		}
		sql += fmt.Sprintf(
			"DELETE FROM `%s`.`%s` WHERE %s ;",
			e.event.Table.Schema, e.event.Table.Name,
			strings.Join(wheres, " AND "),
		)
	}
	return sql
}

func (e *Event) deleteNotPkStatement() string {
	var sql string
	for _, rows := range e.event.Rows {
		var wheres []string
		for idx, val := range rows {
			wheres = append(wheres, fmt.Sprintf("`%s` = %s", e.columns[idx], e.getColumnValueByIdx(val, idx)))
		}
		sql += fmt.Sprintf(
			"DELETE FROM `%s`.`%s` WHERE %s ;",
			e.event.Table.Schema, e.event.Table.Name,
			strings.Join(wheres, " AND "),
		)
	}
	return sql
}

func getValueInsertString(value interface{}) string {
	if value == nil {
		return "NULL"
	}
	switch value.(type) {
	case string:
		return insertStringReplace(value.(string))
	case float64, float32:
		return fmt.Sprintf("%f", value)
	case uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", value)
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", value)
	default:
		return insertStringReplace(fmt.Sprintf("%s", value))
	}
}

func insertStringReplace(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "'", "\\'", -1)
	return fmt.Sprintf("'%s'", str)
}
