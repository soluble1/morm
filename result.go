package morm

import "database/sql"

type Result struct {
	res sql.Result
	err error
}

// LastInsertId 最后插入的行的 id
func (r Result) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.LastInsertId()
}

// RowsAffected 受影响的行数
func (r Result) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.RowsAffected()
}
