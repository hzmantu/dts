package db

type Database struct {
	SourceName string            `json:"source_name"`
	Name       string            `json:"name"`
	Alias      string            `json:"alias"`
	Tables     map[string]*Table `json:"tables"`
	CreateSql  string            `json:"createSql"`
}

func NewDatabase(sourceName string, name string, Tables []*Table) *Database {
	db := &Database{
		SourceName: sourceName,
		Name:       name,
		Tables:     make(map[string]*Table),
	}
	for _, table := range Tables {
		// Assign database name (beware of omission
		table.Database = name
		// turn to hash
		db.Tables[table.Name] = table
	}
	return db
}
