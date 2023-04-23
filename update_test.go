package morm

import (
	"github.com/soluble1/morm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdater_Build(t *testing.T) {
	db := memoryDB(t)
	tests := []struct {
		name string

		u QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name: "single update",

			u: NewUpdater[TestModel](db).Set(C("Id").Eq(24)),

			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `id` = ?;",
				Args: []any{24},
			},
		},
		{
			name: "more update",

			u: NewUpdater[TestModel](db).Set(C("Id").Eq(24), C("Age").Eq(19)),

			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `id` = ?, `age` = ?;",
				Args: []any{24, 19},
			},
		},
		{
			name: "Err Field update",

			u: NewUpdater[TestModel](db).Set(C("Id").Eq(24), C("age").Eq(19)),

			wantErr: errs.NewErrUnKnowField("age"),
		},
		{
			name: "Err update",

			u: NewUpdater[TestModel](db).Set(C("Id").Eq(24), C("Age").Lt(19)),

			wantErr: errs.ErrNonSupportOperator,
		},
		{
			name: "update where",

			u: NewUpdater[TestModel](db).
				Set(C("Id").Eq(24), C("Age").Eq(19)).
				Where(C("Id").Lt(10).Or(C("Age").Gt(18))),

			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `id` = ?, `age` = ? WHERE (`id` < ?) OR (`age` > ?);",
				Args: []any{24, 19, 10, 18},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := tt.u.Build()
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantQuery, q)
		})
	}
}
