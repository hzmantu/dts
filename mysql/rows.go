package mysql

import (
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils"
	"code.hzmantu.com/dts/utils/transform"
	"github.com/xelabs/go-mysqlstack/driver"
	"github.com/xelabs/go-mysqlstack/sqlparser/depends/sqltypes"
)

type Rows struct {
	rows        driver.Rows
	table       *db.Table
	currentData []sqltypes.Value
	currentId   int64
	data        []byte
	conn        *Conn
}

func NewRows(rows driver.Rows, table *db.Table, conn *Conn) *Rows {
	return &Rows{
		rows:  rows,
		table: table,
		conn:  conn,
	}
}

func (r *Rows) GetId() int64 {
	return r.currentId
}

func (r *Rows) Datas() []byte {
	return r.data
}

func (r *Rows) Next() bool {
	if !r.rows.Next() {
		r.Close()
		return false
	}
	r.data = r.data[:0]
	values, err := r.rows.RowValues()
	utils.PanicError(err)
	if r.table.GetPrimaryKey() != "" {
		index := r.table.GetPrimaryKeyIndex()
		r.currentId = transform.StringToInt64(values[index].ToString())
	}

	for i, field := range r.table.GetFields() {
		column := r.table.Columns[field]
		if r.table.HashMask() && column.Mask != nil && !values[i].IsNull() {
			value := utils.CallString(*column.Mask, values[i].ToString())
			values[i], _ = sqltypes.NewValue(values[i].Type(), []byte(value))
		}
		r.data = append(r.data, values[i].Raw()...)
	}
	r.currentData = values
	return true
}

func (r *Rows) GetValues() []sqltypes.Value {
	return r.currentData
}

func (r *Rows) Close() {
	err := r.rows.Close()
	utils.PanicError(err)
	r.conn.Recycle()
}
