package db

import (
	"code.hzmantu.com/dts/utils/tea"
	"fmt"
	"strings"
)

const PRIMARY = "PRIMARY"

type Column struct {
	Field      string
	Type       string
	Collation  *string
	Null       string
	Key        string
	Default    *string
	Extra      string
	Privileges string
	Comment    string
	// filter rule (nil is nothing)
	Mask  *string
	After string
}

func (c *Column) GetSql() string {
	// sql
	sql := fmt.Sprintf("`%s` %s", c.Field, c.Type)

	if c.Collation != nil {
		sql += fmt.Sprintf(" COLLATE %s", *c.Collation)
	}

	if c.Null == "NO" {
		sql += " NOT NULL"
		if c.Type == "timestamp" && (c.Default == nil || *c.Default == "0000-00-00 00:00:00") {
			c.Default = tea.String("CURRENT_TIMESTAMP")
		}
	} else {
		if c.Type == "timestamp" {
			sql += " NULL"
		}
		if c.Default == nil {
			sql += " DEFAULT NULL"
		}
	}

	if c.Default != nil {
		if *c.Default == "CURRENT_TIMESTAMP" {
			sql += fmt.Sprintf(" DEFAULT %s", *c.Default)
		} else {
			sql += fmt.Sprintf(" DEFAULT '%s'", *c.Default)
		}
	}
	// auto_increment
	if c.Extra == "auto_increment" {
		sql += " AUTO_INCREMENT"
	}
	// has COMMENT
	if c.Comment != "" {
		comment := strings.Replace(c.Comment, "'", "\\'", -1)
		sql += fmt.Sprintf(" COMMENT '%s'", comment)
	}
	return sql
}
