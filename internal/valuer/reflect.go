package valuer

import (
	"database/sql"
	"github.com/soluble1/morm/internal/errs"
	"github.com/soluble1/morm/model"
	"reflect"
)

type reflectValue struct {
	val   reflect.Value
	model *model.Model
}

func NewReflectValue(t any, model *model.Model) Value {
	return &reflectValue{
		val:   reflect.ValueOf(t).Elem(),
		model: model,
	}
}

func (r *reflectValue) Field(name string) (any, error) {
	val := r.val

	// 判断一下这个字段存不存在
	typ := val.Type()
	_, ok := typ.FieldByName(name)
	if !ok {
		return nil, errs.NewErrUnKnowField(name)
	}

	return val.FieldByName(name).Interface(), nil
}

func (r *reflectValue) SetColumns(rows *sql.Rows) error {
	// 没有数据
	if !rows.Next() {
		return errs.ErrNoRows
	}
	// Columns 返回查询结果中的所有列名
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	// 校验列数
	if len(cols) > len(r.model.ColumnMap) {
		return errs.ErrTooManyColumns
	}

	vals := make([]any, 0, len(cols))
	eleVals := make([]reflect.Value, 0, len(cols))
	for _, col := range cols {
		fd, ok := r.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnKnowColumn(col)
		}
		// 如果 fd.typ 是 int 那么 reflect.New(fd.typ) 是 *int
		//vals = append(vals, reflect.New(fd.typ).Elem().Interface())
		fdVal := reflect.New(fd.Typ)
		eleVals = append(eleVals, fdVal.Elem())
		// Scan 需要指针不需要调用 Elem
		vals = append(vals, fdVal.Interface())
	}
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	// vals = [123, "long", 18, "xiao"] 将他放到 T 中返回
	tVal := r.val
	for i, col := range cols {
		fd := r.model.ColumnMap[col]
		//tVal.FieldByName(fd.goName).Set(reflect.ValueOf(vals[i]))
		tVal.FieldByIndex(fd.Index).Set(eleVals[i])
	}
	return nil
}
