package morm

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/soluble1/morm/internal/errs"
	"github.com/soluble1/morm/internal/valuer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)
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

func memoryDB(t *testing.T) *DB {
	orm, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	if err != nil {
		t.Fatal(err)
	}
	return orm
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	testCases := []struct {
		name     string
		query    string
		mockErr  error
		mockRows *sqlmock.Rows
		wantErr  error
		wantVal  *TestModel
	}{
		{
			name:    "single row",
			query:   "SELECT .*",
			mockErr: nil,
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow([]byte("123"), []byte("long"), []byte("18"), []byte("xiao"))
				return rows
			}(),
			wantVal: &TestModel{
				Id:        123,
				FirstName: "long",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "xiao"},
			},
		},
		{
			name:    "less col",
			query:   "SELECT .*",
			mockErr: nil,
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name"})
				rows.AddRow([]byte("123"), []byte("long"))
				return rows
			}(),
			wantVal: &TestModel{
				Id:        123,
				FirstName: "long",
			},
		},
		{
			name:    "invalid col",
			query:   "SELECT .*",
			mockErr: nil,
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "gender"})
				rows.AddRow([]byte("123"), []byte("long"), []byte("18"), []byte("xiao"), []byte("man"))
				return rows
			}(),
			wantErr: errs.ErrTooManyColumns,
		},
		{
			name:    "invalid col",
			query:   "SELECT .*",
			mockErr: nil,
			mockRows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name", "first_name"})
				rows.AddRow([]byte("123"), []byte("long"), []byte("18"), []byte("xiao"), []byte("copy_first_name"))
				return rows
			}(),
			wantErr: errs.ErrTooManyColumns,
		},
	}

	for _, tc := range testCases {
		if tc.mockErr != nil {
			mock.ExpectQuery(tc.query).WillReturnError(tc.mockErr)
		} else {
			mock.ExpectQuery(tc.query).WillReturnRows(tc.mockRows)
		}
	}

	db, err := OpenDB(mockDB, DBUseReflectValuer())
	require.NoError(t, err)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := NewSelector[TestModel](db).Get(context.Background())
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantVal, res)
		})
	}

}

func TestSelector_Select(t *testing.T) {
	db := memoryDB(t)
	tests := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 指定列
			name: "specify columns",
			s:    NewSelector[TestModel](db).Select(C("Id"), C("Age")),
			wantQuery: &Query{
				SQL: "SELECT `id`,`age` FROM `test_model`;",
			},
		},
		{
			// 指定聚合函数
			name: "specify aggregate",
			s:    NewSelector[TestModel](db).Select(Min("Id"), Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT MIN(`id`),AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name:    "count distinct 01",
			s:       NewSelector[TestModel](db).Select(Count("DISTINCT `first_name`")),
			wantErr: errs.NewErrUnKnowField("DISTINCT `first_name`"),
		},
		{
			name: "count distinct 02",
			s:    NewSelector[TestModel](db).Select(Raw("DISTINCT `first_name`")),
			wantQuery: &Query{
				SQL: "SELECT DISTINCT `first_name` FROM `test_model`;",
			},
		},
		{
			name: "raw expression",
			s: NewSelector[TestModel](db).
				Where(Raw("`age` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` < ?;",
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

/*
goos: windows
goarch: amd64
pkg: morm
cpu: Intel(R) Pentium(R) Gold G5420 CPU @ 3.80GHz
BenchmarkQuerier_Get/unsafe-4              10000            326979 ns/op            3412 B/op        110 allocs/op
BenchmarkQuerier_Get/reflect-4             10000           1156586 ns/op            3575 B/op        120 allocs/op
*/
func BenchmarkQuerier_Get(b *testing.B) {
	db, err := Open("sqlite3", fmt.Sprintf("file:benchmark_get.db?cache=shared&mode=memory"))
	if err != nil {
		b.Fatal(err)
	}
	_, err = db.db.Exec(TestModel{}.CreateSQL())
	if err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`)"+
		"VALUES (?,?,?,?)", 12, "Deng", 18, "Ming")

	if err != nil {
		b.Fatal(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.Run("unsafe", func(b *testing.B) {
		db.valCreator = valuer.NewUnsafeValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.valCreator = valuer.NewReflectValue
		for i := 0; i < b.N; i++ {
			_, err = NewSelector[TestModel](db).Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func (TestModel) CreateSQL() string {
	return `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`
}
