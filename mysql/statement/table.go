package statement

import (
	"code.hzmantu.com/dts/structs/db"
	"fmt"
)

func DropTable(table *db.Table) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s`;", table.Database, table.Name)
}

func ChangeColumn(table *db.Table, column *db.Column) string {
	return fmt.Sprintf(
		"ALTER TABLE `%s`.`%s` CHANGE `%s` %s;",
		table.Database, table.Name, column.Field, column.GetSql(),
	)
}

func AddColumn(table *db.Table, column *db.Column) string {
	return fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD %s;", table.Database, table.Name, column.GetSql())
}

func DropColumn(table *db.Table, name string) string {
	return fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP COLUMN `%s`;", table.Database, table.Name, name)
}

func DropIndex(table *db.Table, name string) string {
	return fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP INDEX `%s`;", table.Database, table.Name, name)
}

func CreateIndex(table *db.Table, keys *db.Keys) string {
	return fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD %s;", table.Database, table.Name, keys.GetIndexSql())
}

func TruncateTable(table *db.Table) string {
	return fmt.Sprintf("TRUNCATE TABLE `%s`.`%s`;", table.Database, table.Name)
}

func DropConstraint(table *db.Table, name string) string {
	return fmt.Sprintf("ALTER TABLE `%s`.`%s` DROP FOREIGN KEY `%s`;", table.Database, table.Name, name)
}

func CreateConstraint(table *db.Table, constraint *db.Constraint) string {
	return fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD %s;", table.Database, table.Name, constraint.GetSql())
}

func AlterAutoIncrement(table *db.Table, number int64) string {
	return fmt.Sprintf("ALTER TABLE `%s`.`%s` AUTO_INCREMENT=%d;", table.Database, table.Name, number)
}
