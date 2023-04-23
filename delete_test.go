package morm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleter_Build(t *testing.T) {
	db := memoryDB(t)
	tests := []struct {
		name string

		d QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name: "single delete",
			d:    NewDeleter[TestModel](db),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name: "single where delete",
			d:    NewDeleter[TestModel](db).Where(C("Id").Eq(23)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{23},
			},
		},
		{
			name: "more where delete",
			d:    NewDeleter[TestModel](db).Where(C("Id").Eq(23), C("FirstName").Eq("xiaolong")),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE (`id` = ?) AND (`first_name` = ?);",
				Args: []any{23, "xiaolong"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := tt.d.Build()
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantQuery, q)
		})
	}
}
