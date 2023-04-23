package morm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/soluble1/morm/internal/errs"
)

type Deleter[T any] struct {
	builder
	db *DB

	where []Predicate
}

func (d *Deleter[T]) Exec(ctx context.Context) sql.Result {
	m, err := d.Build()
	if err != nil {
		return nil
	}

	execContext, err := d.db.db.ExecContext(ctx, m.SQL, m.Args)
	return Result{
		res: execContext,
		err: err,
	}
}

func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
		},
	}
}

func (d *Deleter[T]) Where(pr ...Predicate) *Deleter[T] {
	d.where = pr
	return d
}

func (d *Deleter[T]) Build() (*Query, error) {
	d.sb.WriteString("DELETE FROM ")

	t := new(T)
	var err error
	d.model, err = d.db.r.Get(t)
	if err != nil {
		return nil, err
	}

	d.quote(d.model.TableName)

	if len(d.where) > 0 {
		d.sb.WriteByte(' ')
		d.sb.WriteString("WHERE ")
		t := d.where[0]
		for i := 1; i < len(d.where); i++ {
			t = t.And(d.where[i])
		}
		err = d.buildExpression(t)
		if err != nil {
			return nil, err
		}
	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) buildExpression(expression Expression) error {
	switch expr := expression.(type) {
	case nil:
		return nil
	case Value:
		d.sb.WriteByte('?')
		if d.args == nil {
			d.args = make([]any, 0, 8)
		}
		d.args = append(d.args, expr.val)
	case Column:
		d.sb.WriteByte('`')
		fd, ok := d.model.FieldMap[expr.name]
		if !ok {
			return errs.NewErrUnKnowField(expr.name)
		}
		d.sb.WriteString(fd.ColName)
		d.sb.WriteByte('`')
	case Predicate:
		P, ok := expr.left.(Predicate)
		if ok && P.op != opNOT {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.left); err != nil {
			return err
		}
		if ok && P.op != opNOT {
			d.sb.WriteByte(')')
		}

		if expr.op != opNOT && expr.op != "" {
			d.sb.WriteByte(' ')

		}
		d.sb.WriteString(expr.op.String())
		if expr.op != "" {
			d.sb.WriteByte(' ')
		}

		_, ok = expr.right.(Predicate)
		if ok {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(expr.right); err != nil {
			return err
		}
		if ok {
			d.sb.WriteByte(')')
		}
	case RawExpr:
		d.sb.WriteString(expr.raw)
		d.args = append(d.args, expr.args...)
	default:
		return errors.New("orm: 不支持的表达式")
	}
	return nil
}
