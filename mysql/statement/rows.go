package statement

import (
	"code.hzmantu.com/dts/structs/db"
	"fmt"
	"github.com/xelabs/go-mysqlstack/sqlparser/depends/sqltypes"
	"strings"
)

type KeyValue struct {
	Id    string
	Value string
}

func InterAllData(table *db.Table, rows [][]sqltypes.Value) string {
	var insertValues []string
	for _, row := range rows {
		var values []string
		for _, value := range row {
			values = append(values, getValueInsertString(value))
		}
		insertValues = append(insertValues, strings.Join(values, ", "))
	}
	return fmt.Sprintf(
		"INSERT IGNORE INTO `%s`.`%s` (`%s`) VALUES (%s);",
		table.Database,
		table.Name,
		strings.Join(table.GetFields(), "`, `"),
		strings.Join(insertValues, "), ("),
	)
}

func UpdateAllData(table *db.Table, rows [][]sqltypes.Value) string {
	var ids, setString []string
	// hash
	updateHash := make(map[string][]KeyValue)
	fields := table.GetFields()
	for _, row := range rows {
		id := row[table.GetPrimaryKeyIndex()].ToString()
		ids = append(ids, id)
		for index, value := range row {
			if index == table.GetPrimaryKeyIndex() {
				continue
			}
			field := fields[index]
			updateHash[field] = append(updateHash[field], KeyValue{
				Id:    id,
				Value: getValueInsertString(value),
			})
		}
	}
	for field, data := range updateHash {
		sql := fmt.Sprintf("`%s` = CASE `%s`", field, table.GetPrimaryKey())
		for _, keyValue := range data {
			sql += fmt.Sprintf(" WHEN %s THEN %s", keyValue.Id, keyValue.Value)
		}
		sql += fmt.Sprintf(" ELSE `%s` END", field)
		setString = append(setString, sql)
	}
	return fmt.Sprintf(
		"UPDATE IGNORE `%s`.`%s` SET %s WHERE `%s` IN (%s);",
		table.Database,
		table.Name,
		strings.Join(setString, ", "),
		table.GetPrimaryKey(),
		strings.Join(ids, ", "),
	)
}

func DeleteDataByIds(table *db.Table, ids []string) string {
	return fmt.Sprintf(
		"DELETE FROM `%s`.`%s` WHERE `%s` IN (%s);",
		table.Database,
		table.Name,
		table.GetPrimaryKey(),
		strings.Join(ids, ", "),
	)
}

func getValueInsertString(value sqltypes.Value) string {
	if value.IsNull() {
		return "NULL"
	}
	str := value.ToString()
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "'", "\\'", -1)
	return fmt.Sprintf("'%s'", str)
}

func DeleteRedundantData(table *db.Table) string {
	return fmt.Sprintf(
		"DELETE FROM `%s`.`%s` WHERE `%s` < %d OR `%s` > %d;",
		table.Database,
		table.Name,
		table.GetPrimaryKey(),
		table.GetMinId(),
		table.GetPrimaryKey(),
		table.GetMaxId(),
	)
}
