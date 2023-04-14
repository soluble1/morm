package morm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	tests := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "From",
			s:    NewSelector[TestModel]().From("`test_sql_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_sql_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "not From",
			s:    &Selector[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "null From",
			s:    NewSelector[TestModel]().From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "with db",
			s:    NewSelector[TestModel]().From("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "single predicate",
			s:    NewSelector[TestModel]().Where(C("id").Eq(12)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `id` = ?;",
				Args: []any{12},
			},
		},
		{
			name: "multi predicate",
			s:    NewSelector[TestModel]().Where(C("Age").Gt(18), C("Age").Lt(35)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` > ?) AND (`Age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			name: "not predicate",
			s:    NewSelector[TestModel]().Where(Not(C("Age").Gt(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`Age` > ?);",
				Args: []any{18},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q, err := test.s.Build()
			assert.Equal(t, test.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, test.wantQuery, q)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
