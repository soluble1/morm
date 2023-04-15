package morm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
	tests := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "From",
			s:    NewSelector[TestModel](db).From("`test_sql_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_sql_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "not From",
			s:    NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "null From",
			s:    NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "with db",
			s:    NewSelector[TestModel](db).From("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_db`.`test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name: "single predicate",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq(12)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{12},
			},
		},
		{
			name: "multi predicate",
			s:    NewSelector[TestModel](db).Where(C("Age").Gt(18), C("Age").Lt(35)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`age` < ?);",
				Args: []any{18, 35},
			},
		},
		{
			name: "not predicate",
			s:    NewSelector[TestModel](db).Where(Not(C("Age").Gt(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE NOT (`age` > ?);",
				Args: []any{18},
			},
		},
		{
			name: "more predicate",
			s: NewSelector[TestModel](db).
				Where(Not(C("Age").Gt(18)), C("Id").Eq(8).Or(C("FirstName").Eq("xiaolong"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE NOT (`age` > ?) AND ((`id` = ?) OR (`first_name` = ?));",
				Args: []any{18, 8, "xiaolong"},
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
