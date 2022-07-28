package db

import (
	"fmt"
)

type Constraint struct {
	Name                  string
	ColumnName            string
	ReferencedTableSchema string
	ReferencedTableName   string
	ReferencedColumnName  string
	UpdateRule            string
	DeleteRule            string
}

func (c *Constraint) GetSql() string {
	return fmt.Sprintf(
		"CONSTRAINT `%s` FOREIGN KEY (`%s`) REFERENCES `%s`.`%s` (`%s`) ON DELETE %s ON UPDATE %s",
		c.Name,
		c.ColumnName,
		c.ReferencedTableSchema,
		c.ReferencedTableName,
		c.ReferencedColumnName,
		c.DeleteRule,
		c.UpdateRule,
	)
}
