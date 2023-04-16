package valuer

import (
	"database/sql"
	"morm/internal/errs"
	"morm/internal/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	t     any
	model *model.Model
	addr  unsafe.Pointer
}

func NewUnsafeValue(t any, model *model.Model) Value {
	addr := unsafe.Pointer(reflect.ValueOf(t).Pointer())
	return &unsafeValue{
		t:     t,
		model: model,
		addr:  addr,
	}
}

func (u *unsafeValue) SetColumns(rows *sql.Rows) error {
	// 没有数据
	if !rows.Next() {
		return errs.ErrNoRows
	}
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	// 校验列数
	if len(cols) > len(u.model.ColumnMap) {
		return errs.ErrTooManyColumns
	}

	vals := make([]any, 0, len(cols))
	for _, col := range cols {
		fd, ok := u.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnKnowColumn(col)
		}

		// 计算字段的真实地址：对象起始地址 + 字段偏移量
		fdVal := reflect.NewAt(fd.Typ, unsafe.Pointer(uintptr(u.addr)+fd.Offset))
		// Scan 需要指针不需要调用 Elem
		vals = append(vals, fdVal.Interface())
	}

	return rows.Scan(vals...)
}
