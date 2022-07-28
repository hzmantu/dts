package structs

import (
	"code.hzmantu.com/dts/utils"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

var TaskConfig *Config

type Config struct {
	StatAddr   string `yaml:"stat_addr"`
	UseStorage bool   `yaml:"use_storage"`
	SyncBinlog bool   `yaml:"sync_binlog"`

	Inputs        []*SourceConfig              `yaml:"inputs"`
	Output        *SourceConfig                `yaml:"output"`
	SingleRowNums int64                        `yaml:"singleRowNums"`
	ReaderNums    int                          `yaml:"readerNums"`
	WriterNums    int                          `yaml:"writerNums"`
	Filters       map[string]map[string]string `yaml:"filters"`
	Debug         bool                         `yaml:"debug"`
}

type SourceConfig struct {
	Name             string            `yaml:"name"`
	Driver           string            `yaml:"driver"`
	Address          string            `yaml:"address"`
	User             string            `yaml:"user"`
	Password         string            `yaml:"password"`
	Charset          string            `yaml:"charset"`
	Only             []string          `yaml:"only"`
	Exclude          []string          `yaml:"exclude"`
	MaxConnNums      int               `yaml:"maxConnNums"`
	DbAlias          map[string]string `yaml:"dbAlias"`
	onlyDatabases    []string
	onlyTables       map[string][]string
	excludeDatabases []string
	excludeTables    map[string][]string
}

type Mask struct {
	Field string `yaml:"field"`
	Mask  string `yaml:"mask"`
}

func GetConfig(filepath string) *Config {
	content, err := ioutil.ReadFile(filepath)
	utils.PanicError(err)

	TaskConfig = &Config{}
	err = yaml.Unmarshal(content, TaskConfig)
	utils.PanicError(err)
	return TaskConfig
}

func (s *SourceConfig) InitSource() {
	s.excludeTables = make(map[string][]string)
	s.onlyTables = make(map[string][]string)

	var database, table string
	databaseMap := make(map[string]bool)
	for _, val := range s.Only {
		// split with .
		index := strings.Index(val, ".")
		if index > 0 {
			database = val[:index]
			table = val[index+1:]
			if table != "" && table != "*" {
				s.onlyTables[database] = append(s.onlyTables[database], table)
			}
		} else {
			database = val
		}

		if _, exist := databaseMap[database]; !exist {
			s.onlyDatabases = append(s.onlyDatabases, database)
		}
	}

	databaseMap = make(map[string]bool)
	for _, val := range s.Exclude {
		index := strings.Index(val, ".")
		if index > 0 {
			table = val[index+1:]
			database = val[:index]
			if table != "" && table != "*" {
				s.excludeTables[database] = append(s.excludeTables[database], table)
				continue
			}
		} else {
			database = val
		}

		if _, exist := databaseMap[database]; !exist {
			s.excludeDatabases = append(s.excludeDatabases, database)
		}
	}
}

func (s *SourceConfig) GetSqlOpenAddress() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/?charset=%s&timeout=3s",
		s.User,
		s.Password,
		s.Address,
		s.Charset,
	)
}

func (s *SourceConfig) GetOnlyDatabases() []string {
	return s.onlyDatabases
}

func (s *SourceConfig) GetExcludeDatabases() []string {
	return s.excludeDatabases
}

func (s *SourceConfig) GetOnlyTables(database string) []string {
	if tables, exist := s.onlyTables[database]; exist {
		return tables
	}
	return []string{}
}

func (s *SourceConfig) GetExcludeTables(database string) []string {
	if tables, exist := s.excludeTables[database]; exist {
		return tables
	}
	return []string{}
}
