package morm

type Aggregate struct {
	arg string
	fn  string
}

func (Aggregate) selectable() {}

func Avg(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  "AVG",
	}
}

func Min(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  "MIN",
	}
}

func Max(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  "MAX",
	}
}

func Count(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  "COUNT",
	}
}

func Sum(col string) Aggregate {
	return Aggregate{
		arg: col,
		fn:  "SUM",
	}
}
