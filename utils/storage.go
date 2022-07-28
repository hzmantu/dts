package utils

import (
	"code.hzmantu.com/dts/structs/db"
	"encoding/json"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	finishTablePrefix    = "FINISH_TABLE_"
	finishDatabasePrefix = "FINISH_DATABASE_"
	BinlogPositionPrefix = "BINLOG_POSITION_"
)

type Storage struct {
	db *leveldb.DB
}

func NewStorage(path string) *Storage {
	// open leveldb's file
	db, err := leveldb.OpenFile(path, nil)
	PanicError(err)
	return &Storage{
		db: db,
	}
}

func (s *Storage) SaveFinishTable(table *db.Table) {
	_ = s.db.Put(s.getFinishTableKey(table), []byte("1"), nil)
}

func (s *Storage) IsFinishTable(table *db.Table) bool {
	ret, _ := s.db.Has(s.getFinishTableKey(table), nil)
	return ret
}

func (s *Storage) IsNewDatabase(database string) bool {
	ret, _ := s.db.Has(s.getFinishDatabaseKey(database), nil)
	return !ret
}

func (s *Storage) Close() {
	_ = s.db.Close()
}

func (s *Storage) getFinishTableKey(table *db.Table) []byte {
	return []byte(finishTablePrefix + table.Database + "_" + table.Name)
}

func (s *Storage) getFinishDatabaseKey(database string) []byte {
	return []byte(finishDatabasePrefix + database)
}

func (s *Storage) PutPosition(name string, pos *mysql.Position) error {
	if exists, _ := s.db.Has([]byte(BinlogPositionPrefix+name), nil); exists {
		return nil
	}
	ret, err := json.Marshal(pos)
	if err != nil {
		return nil
	}
	return s.db.Put([]byte(BinlogPositionPrefix+name), ret, nil)
}

func (s *Storage) GetPosition(name string) (*mysql.Position, error) {
	var pos *mysql.Position
	ret, err := s.db.Get([]byte(BinlogPositionPrefix+name), nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(ret, &pos)
	if err != nil {
		return nil, err
	}
	return pos, nil
}
