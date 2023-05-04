package morm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/soluble1/morm/internal/errs"
)

type Updater[T any] struct {
	builder
	db *DB

	sets  []Predicate
	where []Predicate
}

func (u *Updater[T]) Exec(ctx context.Context) sql.Result {
	q, err := u.Build()
	if err != nil {
		return Result{
			err: err,
		}
	}
	res, err := u.db.db.ExecContext(ctx, q.SQL, q.Args...)
	return Result{
		res: res,
		err: err,
	}
}

func NewUpdater[T any](db *DB) *Updater[T] {
	return &Updater[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
		},
	}
}

func (u *Updater[T]) Set(sets ...Predicate) *Updater[T] {
	u.sets = sets
	return u
}

func (u *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	u.where = ps
	return u
}

func (u *Updater[T]) Build() (*Query, error) {
	t := new(T)
	var err error
	u.model, err = u.db.r.Get(t)
	if err != nil {
		return nil, err
	}

	u.sb.WriteString("UPDATE ")
	u.quote(u.model.TableName)
	u.sb.WriteByte(' ')
	u.sb.WriteString("SET ")

	if len(u.sets) == 0 {
		return nil, errs.ErrNonSupportOperator
	}

	for i := 0; i < len(u.sets); i++ {
		if i > 0 {
			u.sb.WriteByte(',')
			u.sb.WriteByte(' ')
		}
		p := u.sets[i]
		if p.op != opEQ {
			return nil, errs.ErrNonSupportOperator
		}
		l, ok := p.left.(Column)
		if !ok {
			return nil, errs.ErrNonSupportOperator
		}
		r, ok := p.right.(Value)
		if !ok {
			return nil, errs.ErrNonSupportOperator
		}
		lname, ok := u.model.FieldMap[l.name]
		if !ok {
			return nil, errs.NewErrUnKnowField(l.name)
		}
		u.quote(lname.ColName)
		u.sb.WriteByte(' ')
		u.sb.WriteString(p.op.String())
		u.sb.WriteByte(' ')
		u.sb.WriteByte('?')
		u.args = append(u.args, r.val)
	}

	if len(u.where) > 0 {
		u.sb.WriteByte(' ')
		u.sb.WriteString("WHERE ")
		t := u.where[0]
		for i := 1; i < len(u.where); i++ {
			t = t.And(u.where[i])
		}
		err = u.buildExpression(t)
		if err != nil {
			return nil, err
		}
	}

	u.sb.WriteByte(';')

	return &Query{
		SQL:  u.sb.String(),
		Args: u.args,
	}, nil
}

func (u *Updater[T]) buildExpression(expression Expression) error {
	switch expr := expression.(type) {
	case nil:
		return nil
	case Value:
		u.sb.WriteByte('?')
		if u.args == nil {
			u.args = make([]any, 0, 8)
		}
		u.args = append(u.args, expr.val)
	case Column:
		u.sb.WriteByte('`')
		fd, ok := u.model.FieldMap[expr.name]
		if !ok {
			return errs.NewErrUnKnowField(expr.name)
		}
		u.sb.WriteString(fd.ColName)
		u.sb.WriteByte('`')
	case Predicate:
		P, ok := expr.left.(Predicate)
		if ok && P.op != opNOT {
			u.sb.WriteByte('(')
		}
		if err := u.buildExpression(expr.left); err != nil {
			return err
		}
		if ok && P.op != opNOT {
			u.sb.WriteByte(')')
		}

		if expr.op != opNOT && expr.op != "" {
			u.sb.WriteByte(' ')

		}
		u.sb.WriteString(expr.op.String())
		if expr.op != "" {
			u.sb.WriteByte(' ')
		}

		_, ok = expr.right.(Predicate)
		if ok {
			u.sb.WriteByte('(')
		}
		if err := u.buildExpression(expr.right); err != nil {
			return err
		}
		if ok {
			u.sb.WriteByte(')')
		}
	case RawExpr:
		u.sb.WriteString(expr.raw)
		u.args = append(u.args, expr.args...)
	default:
		return errors.New("orm: 不支持的表达式")
	}
	return nil
}
