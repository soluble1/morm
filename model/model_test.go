package model

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"morm/internal/errs"
	"testing"
)

func Test_parseModel(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantModel *Model
		wantErr   error
	}{
		{
			name:  "ptr",
			input: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
					},
					"FirstName": {
						ColName: "first_name",
					},
					"Age": {
						ColName: "age",
					},
					"LastName": {
						ColName: "last_name",
					},
				},
			},
		},
		{
			name:    "struct",
			input:   TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: errs.ErrInputNil,
		},
		{
			name:    "map",
			input:   map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name: "column tag",
			input: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type ColumnTag struct {
					ID uint64 `orm:"column=id"`
				}
				return &ColumnTag{}
			}(),
			wantModel: &Model{
				TableName: "column_tag",
				FieldMap: map[string]*Field{
					// 默认是 i_d
					"ID": {
						ColName: "id",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := registry{}
			m, err := r.Register(tt.input)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantModel, m)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
