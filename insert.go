package morm

import (
	"context"
	"database/sql"
	"github.com/soluble1/morm/internal/errs"
	"github.com/soluble1/morm/model"
)

type Inserter[T any] struct {
	builder
	db      *DB
	values  []*T
	columns []string

	onDuplicate *Upsert
}

func (i *Inserter[T]) Exec(ctx context.Context) sql.Result {
	q, err := i.Build()
	if err != nil {
		return Result{
			err: err,
		}
	}
	res, err := i.db.db.ExecContext(ctx, q.SQL, q.Args...)
	return Result{
		res: res,
		err: err,
	}
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
		},
	}
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Upsert() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	i.sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}

	i.model = m
	i.quote(m.TableName)

	// fields 需要插入列的切片，没有设置则表示插入全部的列
	fields := m.Fields
	// 指定了插入的列
	if len(i.columns) != 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, c := range i.columns {
			fd, ok := m.FieldMap[c]
			if !ok {
				return nil, errs.NewErrUnKnowField(c)
			}
			fields = append(fields, fd)
		}
	}
	i.sb.WriteByte('(')

	for idx, c := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(c.ColName)
	}

	i.sb.WriteByte(')')
	i.sb.WriteString(" VALUES")
	i.args = make([]any, 0, len(i.values)*len(m.Fields))

	for j, val := range i.values {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		//refVal := reflect.ValueOf(val).Elem()
		refVal := i.db.valCreator(val, i.model)
		// 遍历需要插入的列
		for idx, c := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')

			//i.args = append(i.args, refVal.FieldByIndex(c.Index).Interface())
			fdVal, err := refVal.Field(c.GoName)
			if err != nil {
				return nil, err
			}
			i.args = append(i.args, fdVal)
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicate != nil {
		// 构造 ON DUPLICATE KEY 部分
		err = i.dialect.buildDuplicateKey(&i.builder, i.onDuplicate)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')

	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}
