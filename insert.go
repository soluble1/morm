package morm

import (
	"morm/internal/errs"
	"morm/model"
	"reflect"
)

type Inserter[T any] struct {
	builder
	db      *DB
	values  []*T
	columns []string

	onDuplicate *OnDuplicateKey
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

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
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
		refVal := reflect.ValueOf(val).Elem()
		for idx, c := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')

			i.args = append(i.args, refVal.FieldByIndex(c.Index).Interface())
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

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &OnDuplicateKey{
		assigns: assigns,
	}
	return o.i
}

type OnDuplicateKey struct {
	assigns []Assignable
}
