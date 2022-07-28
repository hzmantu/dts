package compare

import (
	"bytes"
	"code.hzmantu.com/dts/mysql"
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils/report"
	"code.hzmantu.com/dts/utils/transform"
	"github.com/xelabs/go-mysqlstack/sqlparser/depends/sqltypes"
)

func (c *Compare) compareRows(from, to *mysql.Rows, table *db.Table) (
	insertData, updateData [][]sqltypes.Value,
	deleteIds []string,
) {
	var fromEnded, toEnded bool
	var fromId, toId int64
	formNext := func() {
		if from.Next() {
			fromId = from.GetId()
			report.ReaderCount.WithLabelValues(table.Database, table.Name, "from").Add(1)
			report.ReaderLength.WithLabelValues(table.Database, table.Name, "from").Add(float64(len(from.Datas())))
		} else {
			fromEnded = true
		}
	}
	toNext := func() {
		if to.Next() {
			toId = to.GetId()
			report.ReaderCount.WithLabelValues(table.Database, table.Name, "to").Add(1)
			report.ReaderLength.WithLabelValues(table.Database, table.Name, "to").Add(float64(len(to.Datas())))
		} else {
			toEnded = true
		}
	}
	formNext()
	toNext()
	for {
		if !fromEnded && !toEnded {
			if fromId == toId {
				if bytes.Compare(from.Datas(), to.Datas()) != 0 {
					updateData = append(updateData, from.GetValues())
				}
				formNext()
				toNext()
			} else if fromId < toId {
				insertData = append(insertData, from.GetValues())
				formNext()
			} else if fromId > toId {
				deleteIds = append(deleteIds, transform.InterfaceToString(toId))
				toNext()
			}
		} else if !fromEnded && toEnded {
			insertData = append(insertData, from.GetValues())
			formNext()
		} else if fromEnded && !toEnded {
			deleteIds = append(deleteIds, transform.InterfaceToString(toId))
			toNext()
		} else {
			return
		}
	}
}
