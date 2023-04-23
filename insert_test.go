package morm

import (
	"database/sql"
	"github.com/soluble1/morm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	tests := []struct {
		name      string
		insert    QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		// 不传入 value
		{
			name:    "no value",
			insert:  NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},

		// 不指定列且只有一条 value
		{
			name: "single inset",
			insert: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "xiao",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "long"},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?);",
				Args: []any{int64(12), "xiao", int8(18), &sql.NullString{Valid: true, String: "long"}},
			},
		},

		// 不指定列有多条 value
		{
			name: "multi inset",
			insert: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "xiao",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "long"},
			}, &TestModel{
				Id:        13,
				FirstName: "ma",
				Age:       22,
				LastName:  &sql.NullString{Valid: true, String: "jun"},
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?),(?,?,?,?);",
				Args: []any{
					int64(12), "xiao", int8(18), &sql.NullString{Valid: true, String: "long"},
					int64(13), "ma", int8(22), &sql.NullString{Valid: true, String: "jun"},
				},
			},
		},

		// 指定列有多条 value
		{
			name: "multi inset",
			insert: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "xiao",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "long"},
			}).Columns("Age", "FirstName"),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`age`,`first_name`) VALUES(?,?);",
				Args: []any{int8(18), "xiao"},
			},
		},

		{
			// 直接指定值
			name: "upsert",
			insert: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "xiao",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "long"},
			}).Upsert().Update(Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?)" +
					" ON DUPLICATE KEY UPDATE `age`=?;",
				Args: []any{int64(12), "xiao", int8(18), &sql.NullString{Valid: true, String: "long"}, 19},
			},
		},

		{
			name: "upsert VALUES(`age`)",
			insert: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "xiao",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "long"},
			}).Upsert().Update(C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?)" +
					" ON DUPLICATE KEY UPDATE `age`=VALUES(`age`);",
				Args: []any{int64(12), "xiao", int8(18), &sql.NullString{Valid: true, String: "long"}},
			},
		},

		{
			name: "upsert multiple",
			insert: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "xiao",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "long"},
			}).Upsert().Update(Assign("Age", 19), C("FirstName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES(?,?,?,?)" +
					" ON DUPLICATE KEY UPDATE `age`=?,`first_name`=VALUES(`first_name`);",
				Args: []any{int64(12), "xiao", int8(18), &sql.NullString{Valid: true, String: "long"}, 19},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := tt.insert.Build()
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantQuery, q)
		})
	}
}
