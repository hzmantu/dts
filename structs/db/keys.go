package db

import (
	"fmt"
	"strings"
)

type Keys struct {
	keys      []*Key
	KeyName   string
	NonUnique int
}

func (k *Keys) Append(key *Key) {
	if key.KeyName == k.KeyName {
		k.keys = append(k.keys, key)
		k.NonUnique = key.NonUnique
	}
}

func (k *Keys) GetSql() string {
	if k.KeyName == PRIMARY {
		return fmt.Sprintf("PRIMARY KEY (%s)", k.getFields())
	}
	if k.NonUnique == 0 {
		// 唯一索引
		return fmt.Sprintf("UNIQUE KEY `%s` (%s)", k.KeyName, k.getFields())
	}
	return fmt.Sprintf("KEY `%s` (%s)", strings.Trim(k.KeyName, "`"), k.getFields())
}

func (k *Keys) getFields() string {
	var columns []string
	for _, key := range k.keys {
		columns = append(columns, key.ColumnName)
	}
	return fmt.Sprintf("`%s`", strings.Join(columns, "`, `"))
}

func (k *Keys) GetIndexSql() string {
	if k.KeyName == PRIMARY {
		return fmt.Sprintf("PRIMARY INDEX (%s)", k.getFields())
	}
	if k.NonUnique == 0 {
		// 唯一索引
		return fmt.Sprintf("UNIQUE INDEX `%s` (%s)", strings.Trim(k.KeyName, "`"), k.getFields())
	}
	return fmt.Sprintf("INDEX `%s` (%s)", k.KeyName, k.getFields())
}
