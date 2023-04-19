package valuer

import (
	"database/sql"
	"morm/model"
)

// Value 是对结构体实例的内部抽象
type Value interface {
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}

// Creator 简单的factory
type Creator func(t any, model *model.Model) Value
