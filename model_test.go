package morm

import (
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
				tableName: "test_model",
				fieldMap: map[string]*field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"Age": {
						colName: "age",
					},
					"LastName": {
						colName: "last_name",
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
				tableName: "column_tag",
				fieldMap: map[string]*field{
					// 默认是 i_d
					"ID": {
						colName: "id",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registry{}
			m, err := r.Register(tt.input)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantModel, m)
		})
	}
}
