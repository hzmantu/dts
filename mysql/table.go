package mysql

import (
	"code.hzmantu.com/dts/structs/db"
	"fmt"
	"github.com/xelabs/go-mysqlstack/driver"
	"strings"
)

func (c *Conn) GetTableRows(t *db.Table, min, step int64) *Rows {
	var rows driver.Rows

	if t.GetPrimaryKey() != "" {
		rows = c.Query(fmt.Sprintf(
			"SELECT `%s` FROM `%s`.`%s` WHERE `%s` BETWEEN %d AND %d;",
			strings.Join(t.GetFields(), "`, `"),
			t.Database,
			t.Name,
			t.GetPrimaryKey(),
			min,
			min+step-1,
		))
	} else {
		rows = c.Query(fmt.Sprintf(
			"SELECT `%s` FROM `%s`.`%s` LIMIT %d, %d;",
			strings.Join(t.GetFields(), "`, `"),
			t.Database,
			t.Name,
			min,
			step,
		))
	}

	return NewRows(rows, t, c)
}
