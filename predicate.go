package morm

type op string

const (
	opEQ = "="
	opLT = "<"
	opGT = ">"

	opNOT = "NOT"
	opAND = "AND"
	opOR  = "OR"
)

func (o op) String() string {
	return string(o)
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (Predicate) expr() {}

func (p1 Predicate) And(p2 Predicate) Predicate {
	return Predicate{
		left:  p1,
		op:    opAND,
		right: p2,
	}
}

func (p1 Predicate) Or(p2 Predicate) Predicate {
	return Predicate{
		left:  p1,
		op:    opOR,
		right: p2,
	}
}

type Column struct {
	name string
}

func (c Column) selectable() {}

func (c Column) expr() {}

func (c Column) assign() {}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(args any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: valueOf(args),
	}
}

func (c Column) Lt(args any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: valueOf(args),
	}
}

func (c Column) Gt(args any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: valueOf(args),
	}
}

func Not(p Predicate) Predicate {
	return Predicate{
		left:  nil,
		op:    opNOT,
		right: p,
	}
}

type Value struct {
	val any
}

func (Value) expr() {}

func valueOf(val any) Expression {
	return Value{
		val: val,
	}
}
