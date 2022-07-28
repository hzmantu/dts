package db

import (
	"fmt"
	"strings"
)

type Table struct {
	Database        string
	Name            string
	Engine          string
	AutoIncrement   int64
	Collation       string
	CreateOptions   string
	Comment         string
	Columns         map[string]*Column
	Keys            map[string]*Keys
	Constraints     map[string]*Constraint
	primaryKey      *string
	primaryKeyIndex int
	fields          []string
	maxId           int64
	minId           int64
	hashMask        bool
	hasUnique       bool
	toBeSql         []string
}

func (t *Table) AppendElement(columns []*Column, keys []*Key, constraints []*Constraint, mask map[string]string) {
	t.Columns = make(map[string]*Column)
	t.Keys = make(map[string]*Keys)
	t.Constraints = make(map[string]*Constraint)
	var after string

	for index, column := range columns {
		t.Columns[column.Field] = column
		t.fields = append(t.fields, column.Field)
		if column.Key == "PRI" {
			t.primaryKey = &column.Field
			t.primaryKeyIndex = index
		}
		if rule, exist := mask[column.Field]; exist {
			column.Mask = &rule
			t.hashMask = true
		}
		column.After = after
		after = column.Field
	}

	for _, key := range keys {
		if _, exist := t.Keys[key.KeyName]; !exist {
			t.Keys[key.KeyName] = &Keys{
				KeyName: key.KeyName,
			}
			if key.NonUnique == 0 && key.KeyName != "PRIMARY" {
				t.hasUnique = true
			}
		}
		t.Keys[key.KeyName].Append(key)
	}

	for _, constraint := range constraints {
		t.Constraints[constraint.Name] = constraint
	}

	if key, exist := t.Keys["PRIMARY"]; !exist || len(key.keys) > 1 {
		t.primaryKey = nil
	}
}

func (t *Table) SaveDataInterval(min int64, max int64) {
	t.minId = min
	t.maxId = max
}

func (t *Table) GetMaxId() int64 {
	return t.maxId
}

func (t *Table) GetMinId() int64 {
	return t.minId
}

func (t *Table) GetPrimaryKey() string {
	if t.primaryKey == nil {
		return ""
	}
	return *t.primaryKey
}

func (t *Table) GetPrimaryKeyIndex() int {
	return t.primaryKeyIndex
}

func (t *Table) GetFields() []string {
	return t.fields
}

func (t *Table) GetCreateSql() string {
	sep := ", \n"
	var columnSql, keySql []string
	for _, field := range t.GetFields() {
		columnSql = append(columnSql, t.Columns[field].GetSql())
	}
	for _, keys := range t.Keys {
		keySql = append(keySql, keys.GetSql())
	}
	contentSql := strings.Join(columnSql, sep)
	if len(keySql) > 0 {
		contentSql += sep + strings.Join(keySql, sep)
	}
	sql := fmt.Sprintf(
		"CREATE TABLE `%s`.`%s` (%s)",
		t.Database,
		t.Name,
		contentSql,
	)
	if t.Engine != "" {
		sql += fmt.Sprintf(" ENGINE=%s", t.Engine)
	}
	if t.AutoIncrement != 0 {
		sql += fmt.Sprintf(" AUTO_INCREMENT=%d", t.AutoIncrement)
	}
	if t.Collation != "" {
		sql += fmt.Sprintf(
			" DEFAULT CHARSET=%s COLLATE=%s",
			GetCharsetFromCollate(t.Collation),
			t.Collation,
		)
	}

	if t.CreateOptions != "" {
		sql += fmt.Sprintf(" %s", t.CreateOptions)
	}
	return sql
}

func (t *Table) HashMask() bool {
	return t.hashMask
}

func (t *Table) DisableMask() {
	t.hashMask = false
}

func (t *Table) NewColumns(columns map[string]*Column, fields []string) {
	t.Columns = columns
	t.fields = fields
}

func (t *Table) AppendToBeSql(sql string) {
	t.toBeSql = append(t.toBeSql, sql)
}

func (t *Table) GetToBeSql() []string {
	defer func() {
		t.toBeSql = []string{}
	}()
	return t.toBeSql
}

func GetCharsetFromCollate(collate string) string {
	index := strings.Index(collate, "_")
	if index > 0 {
		// 返回
		return collate[:index]
	}
	return collate
}
