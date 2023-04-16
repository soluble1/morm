package valuer

import (
	"database/sql"
	"morm/internal/errs"
	"morm/internal/model"
	"reflect"
)

type reflectValue struct {
	t     any
	model *model.Model
}

func NewReflectValue(t any, model *model.Model) Value {
	return &reflectValue{
		t:     t,
		model: model,
	}
}

func (r *reflectValue) SetColumns(rows *sql.Rows) error {
	// 没有数据
	if !rows.Next() {
		return errs.ErrNoRows
	}
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
	t := r.t
	tVal := reflect.ValueOf(t).Elem()
	for i, col := range cols {
		fd := r.model.ColumnMap[col]
		//tVal.FieldByName(fd.goName).Set(reflect.ValueOf(vals[i]))
		tVal.FieldByName(fd.GoName).Set(eleVals[i])
	}
	return nil
}
