package morm

// Expression 标记接口
type Expression interface {
	expr()
}

type Selectable interface {
	selectable()
}

type RawExpr struct {
	raw  string
	args []any
}

func (r RawExpr) selectable() {}

func (r RawExpr) expr() {}

func Raw(raw string, args ...any) RawExpr {
	return RawExpr{
		raw:  raw,
		args: args,
	}
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
