package morm

import (
	"context"
	"errors"
	"morm/internal/errs"
	model2 "morm/internal/model"
	"strings"
)

type Selector[T any] struct {
	tbl   string
	where []Predicate
	sb    strings.Builder

	args  []any
	model *model2.Model
	db    *DB

	columns []Selectable
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.tbl = tbl
	return s
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	t := new(T)
	var err error
	s.model, err = s.db.r.Get(t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	if len(s.columns) == 0 {
		s.sb.WriteString("*")
	} else {
		for i, c := range s.columns {
			switch col := c.(type) {
			case Column:
				fd, ok := s.model.FieldMap[col.name]
				if !ok {
					return nil, errs.NewErrUnKnowField(col.name)
				}
				if i > 0 {
					s.sb.WriteByte(',')
				}
				s.sb.WriteByte('`')
				s.sb.WriteString(fd.ColName)
				s.sb.WriteByte('`')
			case Aggregate:
				fd, ok := s.model.FieldMap[col.arg]
				if !ok {
					return nil, errs.NewErrUnKnowField(col.arg)
				}
				if i > 0 {
					s.sb.WriteByte(',')
				}
				s.sb.WriteString(col.fn)
				s.sb.WriteByte('(')
				s.sb.WriteByte('`')
				s.sb.WriteString(fd.ColName)
				s.sb.WriteByte('`')
				s.sb.WriteByte(')')
			case RawExpr:
				s.sb.WriteString(col.raw)
				if len(col.args) > 0 {
					s.args = append(s.args, col.args...)
				}
			}
		}
	}
	s.sb.WriteString(" FROM ")

	if s.tbl == "" {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.tbl)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		pred := s.where[0]
		for i := 1; i < len(s.where); i++ {
			pred = pred.And(s.where[i])
		}
		err := s.buildExpression(pred)
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(expression Expression) error {
	switch expr := expression.(type) {
	case nil:
		return nil
	case Value:
		s.sb.WriteByte('?')
		if s.args == nil {
			s.args = make([]any, 0, 8)
		}
		s.args = append(s.args, expr.val)
	case Column:
		s.sb.WriteByte('`')
		fd, ok := s.model.FieldMap[expr.name]
		if !ok {
			return errs.NewErrUnKnowField(expr.name)
		}
		s.sb.WriteString(fd.ColName)
		s.sb.WriteByte('`')
	case Predicate:
		P, ok := expr.left.(Predicate)
		if ok && P.op != opNOT {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.left); err != nil {
			return err
		}
		if ok && P.op != opNOT {
			s.sb.WriteByte(')')
		}

		if expr.op != opNOT && expr.op != "" {
			s.sb.WriteByte(' ')

		}
		s.sb.WriteString(expr.op.String())
		if expr.op != "" {
			s.sb.WriteByte(' ')
		}

		_, ok = expr.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
	case RawExpr:
		s.sb.WriteString(expr.raw)
		s.args = append(s.args, expr.args...)
	default:
		return errors.New("orm: 不支持的表达式")
	}
	return nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	t := new(T)

	val := s.db.valCreator(t, s.model)

	return t, val.SetColumns(rows)
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// TODO implement me
	panic("implement me")
}
