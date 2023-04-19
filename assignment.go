package morm

type Assignable interface {
	assign()
}

type Assignment struct {
	column string
	val    any
}

func (a Assignment) assign() {}

func Assign(column string, val any) Assignable {
	return Assignment{
		column: column,
		val:    val,
	}
}
