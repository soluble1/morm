package morm

import (
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	tbl string
}

func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.tbl = tbl
	return s
}

func NewSelector[T any]() *Selector[T] {
	return &Selector[T]{}
}

func (s *Selector[T]) Build() (*Query, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	if s.tbl == "" {
		var t T
		typ := reflect.TypeOf(t)
		name := typ.Name()
		sb.WriteByte('`')
		sb.WriteString(name)
		sb.WriteByte('`')
	} else {
		sb.WriteString(s.tbl)
	}
	sb.WriteByte(';')
	return &Query{
		SQL: sb.String(),
	}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}
