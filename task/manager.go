package task

import (
	"code.hzmantu.com/dts/mysql"
	"code.hzmantu.com/dts/mysql/compare"
	"code.hzmantu.com/dts/mysql/statement"
	"code.hzmantu.com/dts/structs"
	"code.hzmantu.com/dts/structs/db"
	"code.hzmantu.com/dts/utils"
	"code.hzmantu.com/dts/utils/report"
	"github.com/xelabs/go-mysqlstack/sqlparser/depends/sqltypes"
	"log"
	"os"
	"sync"
)

type Manager struct {
	config    *structs.Config
	wg        sync.WaitGroup
	storage   *utils.Storage
	canalChan chan *bool
}

func NewManager(config *structs.Config) *Manager {
	if !config.UseStorage {
		_ = os.RemoveAll("runtime/storage")
	}
	return &Manager{
		config:    config,
		storage:   utils.NewStorage("runtime/storage"),
		canalChan: make(chan *bool, 0),
	}
}

func (m *Manager) Start() {
	// output resource
	to := mysql.NewSource(m.config.Output, m.storage)
	// Disable foreign key, unique index checks
	to.DisableKey()
	var fromList []*mysql.Source
	// range inputs
	for _, input := range m.config.Inputs {
		// input resource
		from := mysql.NewSource(input, m.storage)
		fromList = append(fromList, from)
		// Record the location of binlog before synchronization
		_ = m.storage.PutPosition(input.Name, from.GetMasterPos())
		// Initialize the compare
		c := compare.NewCompare(from, to, m.config.SingleRowNums)
		// start comparing
		m.wg.Add(1)
		go func() {
			c.Start()
			m.wg.Done()
		}()
		// start export
		m.wg.Add(1)
		go func() {
			m.exportTableData(to, from, c.ExportTable)
			m.wg.Done()
		}()
	}
	m.wg.Wait()
	log.Printf("comparison synchronization task completed...")
	// append binlog
	m.RunCanal(fromList, to)
	// close storage
	m.storage.Close()
}

func (m *Manager) exportTableData(to, from *mysql.Source, exportTable <-chan *db.Table) {
	// range exportTable
	for table := range exportTable {
		log.Printf("syncing `%s`.`%s` data", table.Database, table.Name)

		// if there is no primary key, clear the data table
		if table.GetPrimaryKey() == "" {
			_ = to.GetConn().Exec(statement.TruncateTable(table))
		}
		// for
		for i := table.GetMinId(); i <= table.GetMaxId(); i += m.config.SingleRowNums {
			var list [][]sqltypes.Value
			// query
			rows := from.GetConn().GetTableRows(table, i, m.config.SingleRowNums)
			for rows.Next() {
				report.ReaderLength.WithLabelValues(table.Database, table.Name, "dump").Add(float64(len(rows.Datas())))
				list = append(list, rows.GetValues())
			}
			// has data
			if len(list) > 0 {
				// report metrics
				report.DumpCount.WithLabelValues(from.Name, table.Database, table.Name).Add(float64(len(list)))
				_ = to.GetConn().ExecInsertData(table, list)
			}
		}
		// Execute the remaining sql
		for _, sql := range table.GetToBeSql() {
			_ = to.GetConn().Exec(sql)
		}
		// Stored in the cache
		m.storage.SaveFinishTable(table)
		log.Printf("Dump `%s`.`%s` Success", table.Database, table.Name)
	}
	log.Printf("dump %s all data completed", from.Name)
}

func (m *Manager) RunCanal(fromList []*mysql.Source, to *mysql.Source) {
	// filtered
	for _, from := range fromList {
		// start binlog
		can, eventChan := mysql.NewCanal(from, m.storage, m.config.SyncBinlog)
		if can == nil {
			continue
		}
		m.wg.Add(1)
		go func() {
			for event := range eventChan {
				_ = to.GetConn().Exec(event.Sql)
			}
			can.Close()
			m.wg.Done()
		}()
	}
	m.wg.Wait()
	log.Println("Incremental synchronization completed...")
}
