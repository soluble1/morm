package model

import (
	"morm/internal/errs"
	"reflect"
)

type ModelOpt func(m *Model) error

type Model struct {
	// 结构体对应的表名
	TableName string
	// 字段名对应的列名
	FieldMap map[string]*field

	// 列名到字段的映射
	ColumnMap map[string]*field
}

func ModelWithTableName(name string) ModelOpt {
	return func(m *Model) error {
		if name == "" {
			return errs.ErrEmptyTableName
		}
		m.TableName = name
		return nil
	}
}

func ModeWithColumnName(field string, colName string) ModelOpt {
	return func(m *Model) error {
		fd, ok := m.FieldMap[field]
		if !ok {
			return errs.NewErrUnKnowField(field)
		}
		fd.ColName = colName
		return nil
	}
}

func ModeWithColumn(field string, col *field) ModelOpt {
	return func(m *Model) error {
		m.FieldMap[field] = col
		return nil
	}
}

type field struct {
	// 字段名
	GoName string
	// 字段对应的列名
	ColName string

	// 字段偏移量
	Offset uintptr

	Typ reflect.Type
}

type TableName interface {
	TableName() string
}
