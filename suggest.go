package suggest

import (
	"database/sql"
	"strings"

	"github.com/lib/pq"
)

type Suggest struct {
	DataSource string
}

// Prepare a buck insertion for lots of data. If check returns false, then bulk insertion won't be executed.
func (suggest Suggest) BulkInsert(check func(_ *sql.DB) bool, bulk func(_ *sql.Stmt) error, table string, columns ...string) (err error) {
	var db *sql.DB
	db, err = sql.Open("postgres", suggest.DataSource)
	if err != nil {
		return
	}
	defer db.Close()

	defer func() {
		if err != nil && strings.Contains(err.Error(), "duplicate key value") {
			err = nil
		}
	}()

	if check != nil && !check(db) {
		return
	}

	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return
	}

	var stmt *sql.Stmt
	stmt, err = tx.Prepare(pq.CopyIn(table, columns...))
	if err != nil {
		return
	}

	if err = bulk(stmt); err != nil {
		return
	}

	if _, err = stmt.Exec(); err != nil {
		return
	}

	if err = stmt.Close(); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	return
}

// Execute an SQL query statement.
func (suggest Suggest) Query(sqlQuery string, args ...interface{}) (rets []interface{}, err error) {
	var db *sql.DB
	db, err = sql.Open("postgres", suggest.DataSource)
	if err != nil {
		return
	}
	defer db.Close()
	var rows *sql.Rows
	rows, err = db.Query(sqlQuery, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cols []string
		cols, err = rows.Columns()
		if err != nil {
			return
		}
		var ret []interface{}
		for range cols {
			var data interface{}
			ret = append(ret, &data)
		}
		if err = rows.Scan(ret...); err != nil {
			return
		}
		rets = append(rets, ret)
	}
	err = rows.Err()
	return
}
