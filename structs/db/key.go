package db

type Key struct {
	Table      string
	NonUnique  int
	KeyName    string
	SeqInIndex int
	ColumnName string
}
