package morm

import (
	"morm/internal/errs"
)

// Dialect 方言构造部分
type Dialect interface {
	// 引号
	quoter() byte
	buildDuplicateKey(b *builder, odk *OnDuplicateKey) error
}

// SQL 标准的方言实现
type standardSQL struct {
}

type mysqlDialect struct {
	standardSQL
}

func (dialect *mysqlDialect) quoter() byte {
	return '`'
}

func (dialect *mysqlDialect) buildDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch expr := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[expr.column]
			if !ok {
				return errs.NewErrUnKnowField(expr.column)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=?")
			b.args = append(b.args, expr.val)
		case Column:
			fd, ok := b.model.FieldMap[expr.name]
			if !ok {
				return errs.NewErrUnKnowField(expr.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		}
	}
	return nil
}

type sqliteDialect struct {
	standardSQL
}
