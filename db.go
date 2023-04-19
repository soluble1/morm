package morm

import (
	"database/sql"
	"github.com/soluble1/morm/internal/valuer"
	"github.com/soluble1/morm/model"
)

type DBOption func(db *DB)

type DB struct {
	db *sql.DB
	r  model.Registry

	valCreator valuer.Creator

	dialect Dialect
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

// Open 如果用户没有指定 registry 就使用默认的 registry
func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:          model.NewRegistry(),
		db:         db,
		valCreator: valuer.NewUnsafeValue,
		dialect:    &mysqlDialect{},
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func DBUseReflectValuer() DBOption {
	return func(db *DB) {
		db.valCreator = valuer.NewReflectValue
	}
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}
