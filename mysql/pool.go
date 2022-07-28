package mysql

import (
	"code.hzmantu.com/dts/structs"
	"sync"
)

type Pool struct {
	Config   *structs.SourceConfig
	ch       chan *Conn
	max      int
	num      int
	closed   bool
	readLock sync.Mutex
	putLock  sync.Mutex
	initSql  []string
}

func (p *Pool) AppendInitSql(sql string) {
	p.initSql = append(p.initSql, sql)
	for i := 0; i < p.num; i++ {
		p.Get().Exec(sql)
	}
}

func NewPool(config *structs.SourceConfig) *Pool {
	if config.MaxConnNums == 0 {
		config.MaxConnNums = 3
	}
	return &Pool{
		Config: config,
		ch:     make(chan *Conn, config.MaxConnNums),
		max:    config.MaxConnNums,
	}
}

func (p *Pool) newConn() *Conn {
	conn := NewConn(p)
	for _, sql := range p.initSql {
		conn.Run(sql)
	}
	return conn
}

func (p *Pool) Get() *Conn {
	p.readLock.Lock()
	defer p.readLock.Unlock()
	if len(p.ch) == 0 && p.num < p.max {
		p.num++
		return p.newConn()
	} else {
		return <-p.ch
	}
}

func (p *Pool) Put(conn *Conn) {
	p.putLock.Lock()
	defer p.putLock.Unlock()

	if len(p.ch) < p.max {
		p.ch <- conn
	} else {
		_ = conn.Close()
	}
}

func (p *Pool) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true
	if p.ch != nil {
		close(p.ch)
	}

	for c := range p.ch {
		_ = c.Close()
	}
	return nil
}
