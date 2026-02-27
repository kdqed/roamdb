package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Request struct {
	Database string `json:"database"`
	Query    string `json:"query"`
	Params   []any  `json:"params"`
}

type Response struct {
	Columns []string         `json:"columns,omitempty"`
	Rows    []map[string]any `json:"rows,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type job struct {
	query  string
	params []any
	result chan Response
}

type dbEntry struct {
	queue chan job
}

var (
	dbsMu sync.Mutex
	dbs   = map[string]*dbEntry{}
)

func getEntry(name string) (*dbEntry, error) {
	dbsMu.Lock()
	defer dbsMu.Unlock()
	if e, ok := dbs[name]; ok {
		return e, nil
	}
	os.MkdirAll("data", 0755)
	db, err := sql.Open("sqlite3", filepath.Join("data", name+".db"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	e := &dbEntry{queue: make(chan job, 256)}
	go func() {
		for j := range e.queue {
			rows, err := db.Query(j.query, j.params...)
			var r Response
			if err != nil {
				r.Error = err.Error()
				j.result <- r
				continue
			}
			cols, _ := rows.Columns()
			r.Columns = cols
			for rows.Next() {
				vals := make([]any, len(cols))
				ptrs := make([]any, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				rows.Scan(ptrs...)
				row := map[string]any{}
				for i, c := range cols {
					row[c] = vals[i]
				}
				r.Rows = append(r.Rows, row)
			}
			rows.Close()
			j.result <- r
		}
	}()
	dbs[name] = e
	return e, nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			return
		}
		e, err := getEntry(req.Database)
		if err != nil {
			enc.Encode(Response{Error: err.Error()})
			return
		}
		ch := make(chan Response, 1)
		e.queue <- job{query: req.Query, params: req.Params, result: ch}
		enc.Encode(<-ch)
	}
}

func main() {
    port := ":26227"
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on " + port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}
