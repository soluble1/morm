package morm

import "morm/internal/errs"

type ModelOpt func(m *Model) error

type Model struct {
	// 结构体对应的表名
	tableName string
	// 字段名对应的列名
	fieldMap map[string]*field
}

func ModelWithTableName(name string) ModelOpt {
	return func(m *Model) error {
		if name == "" {
			return errs.ErrEmptyTableName
		}
		m.tableName = name
		return nil
	}
}

func ModeWithColumnName(field string, colName string) ModelOpt {
	return func(m *Model) error {
		fd, ok := m.fieldMap[field]
		if !ok {
			return errs.NewErrUnKnowField(field)
		}
		fd.colName = colName
		return nil
	}
}

func ModeWithColumn(field string, col *field) ModelOpt {
	return func(m *Model) error {
		m.fieldMap[field] = col
		return nil
	}
}

type field struct {
	// 字段对应的列名
	colName string
}

type TableName interface {
	TableName() string
}
