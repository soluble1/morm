package valuer

import (
	"database/sql"
	"fmt"
	"github.com/soluble1/morm/internal/errs"
	"github.com/soluble1/morm/model"
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

func (u *unsafeValue) Field(name string) (any, error) {
	fdMeta, ok := u.model.FieldMap[name]
	if !ok {
		return 0, fmt.Errorf("invalid field %s", name)
	}
	ptr := unsafe.Pointer(uintptr(u.addr) + fdMeta.Offset)
	if ptr == nil {
		return 0, fmt.Errorf("invalid address of the field: %s", name)
	}
	// 创建一个新的指向该字段的指针
	res := reflect.NewAt(fdMeta.Typ, ptr).Elem().Interface()
	return res, nil
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

func (u *unsafeValue) GetStructs(rows *sql.Rows) error {
	return nil
}
