package morm

type DBOption func(db *DB)

type DB struct {
	r *registry
}

func DBWithRegistry(r *registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

// NewDB 如果用户没有指定 registry 就使用默认的 registry
func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{
		r: &registry{},
	}

	for _, opt := range opts {
		opt(db)
	}

	return db, nil
}
