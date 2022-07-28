package mysql

import (
	"code.hzmantu.com/dts/utils"
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"log"
	"math/rand"
	"time"
)

type EventData struct {
	Schema string
	Table  string
	Sql    string
}

func NewCanal(from *Source, storage *utils.Storage, syncBinlog bool) (*canal.Canal, chan EventData) {
	pos, err := storage.GetPosition(from.Name)
	if err != nil || pos == nil {
		return nil, nil
	}

	if currentPos := from.GetMasterPos(); currentPos.String() == pos.String() && syncBinlog == false {
		log.Printf("source %s binlog no changes", from.Name)
		return nil, nil
	}
	now := time.Now().Unix()

	cfg := canal.NewDefaultConfig()
	cfg.Addr = from.Config.Address
	cfg.User = from.Config.User
	cfg.Password = from.Config.Password
	cfg.Flavor = mysql.MySQLFlavor
	cfg.Charset = mysql.DEFAULT_CHARSET
	cfg.ServerID = uint32(rand.New(rand.NewSource(now)).Intn(1000)) + 1001

	if syncBinlog == true {
		now = 0
	}

	channel := make(chan EventData, 10)
	can, err := canal.NewCanal(cfg)
	utils.PanicError(err)
	eventHandler := &EventHandler{
		e:        channel,
		endTime:  now,
		startPos: pos,
		source:   from,
		storage:  storage,
	}
	can.SetEventHandler(eventHandler)
	go func() {
		if utils.CheckError(can.RunFrom(*pos)) {
			_ = eventHandler.finish()
		}
	}()

	return can, channel
}

type EventHandler struct {
	e        chan EventData
	close    bool
	endTime  int64
	startPos *mysql.Position
	source   *Source
	storage  *utils.Storage
	canal.DummyEventHandler
}

func (h *EventHandler) String() string {
	return "DDLEventHandler"
}

func (h *EventHandler) OnXID(nextPos mysql.Position) error {
	_ = h.storage.PutPosition(h.source.Name, &nextPos)
	return nil
}

func (h *EventHandler) OnDDL(_ mysql.Position, event *replication.QueryEvent) error {
	if h.close == true {
		return nil
	}
	p := parser.New()
	stmts, _, err := p.Parse(string(event.Query), "", "")
	if err != nil {
		fmt.Errorf("parse query(%s) err %v, will skip this e", event.Query, err)
		return err
	}

	for _, stmt := range stmts {
		nodes := parseStmt(stmt)
		for _, node := range nodes {
			if node.db == "" {
				node.db = string(event.Schema)
			}
			h.e <- EventData{
				Schema: node.db,
				Table:  node.table,
				Sql:    string(event.Query),
			}
		}
	}
	return nil
}

func (h *EventHandler) OnRow(event *canal.RowsEvent) error {
	if h.close == true {
		return nil
	}
	// if the binlog deadline is exceeded
	if h.endTime > 0 && int64(event.Header.Timestamp) > h.endTime {
		return h.finish()
	}

	database, exist := h.source.Databases[event.Table.Schema]
	if !exist {
		return nil
	}
	tableInfo, exist := database.Tables[event.Table.Name]
	if !exist {
		return nil
	}

	if sql := NewEvent(event, tableInfo).GetSql(); sql != "" {
		h.e <- EventData{
			Schema: event.Table.Schema,
			Table:  event.Table.Name,
			Sql:    sql,
		}
	}
	return nil
}

func (h *EventHandler) finish() error {
	h.close = true
	close(h.e)
	return nil
}

type node struct {
	db    string
	table string
}

func parseStmt(stmt ast.StmtNode) (ns []*node) {
	switch t := stmt.(type) {
	case *ast.RenameTableStmt:
		for _, tableInfo := range t.TableToTables {
			n := &node{
				db:    tableInfo.OldTable.Schema.String(),
				table: tableInfo.OldTable.Name.String(),
			}
			ns = append(ns, n)
		}
	case *ast.AlterTableStmt:
		n := &node{
			db:    t.Table.Schema.String(),
			table: t.Table.Name.String(),
		}
		ns = []*node{n}
	case *ast.DropTableStmt:
		for _, table := range t.Tables {
			n := &node{
				db:    table.Schema.String(),
				table: table.Name.String(),
			}
			ns = append(ns, n)
		}
	case *ast.CreateTableStmt:
		n := &node{
			db:    t.Table.Schema.String(),
			table: t.Table.Name.String(),
		}
		ns = []*node{n}
	}
	return ns
}
