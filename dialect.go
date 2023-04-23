package morm

import (
	"github.com/soluble1/morm/internal/errs"
)

// Dialect 方言构造部分
type Dialect interface {
	// 引号
	quoter() byte
	buildDuplicateKey(b *builder, odk *Upsert) error
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

func (dialect *mysqlDialect) buildDuplicateKey(b *builder, odk *Upsert) error {
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

func (dialect *sqliteDialect) quoter() byte {
	return '`'
}

/*
https://www.sqlite.org/lang_UPSERT.html

		INSERT INTO phonebook(name,phonenumber) VALUES('Alice','704-555-1212')
	  		ON CONFLICT(name) DO UPDATE SET phonenumber=excluded.phonenumber;
*/
func (dialect *sqliteDialect) buildDuplicateKey(b *builder, odk *Upsert) error {
	b.sb.WriteString(" ON CONFLICT")
	if len(odk.conflictColumns) > 0 {
		b.sb.WriteString(" (")
		for i, col := range odk.conflictColumns {
			if i > 0 {
				b.sb.WriteByte(',')
			}
			fd, ok := b.model.FieldMap[col]
			if !ok {
				return errs.NewErrUnKnowField(col)
			}
			b.quote(fd.ColName)
		}
		b.sb.WriteByte(')')
	}
	b.sb.WriteString(" DO UPDATE SET ")
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
			b.sb.WriteString("=excluded.")
			b.quote(fd.ColName)
		}
	}
	return nil
}

// postgreSQL 的 DuplicateKey 和 sqlite 的一样，但是引号不同
type postgreSQL struct {
	standardSQL
}
