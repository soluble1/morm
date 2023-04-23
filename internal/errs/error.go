package errs

import (
	"errors"
	"fmt"
)

var (
	ErrInputNil           = errors.New("orm: 不支持 nil")
	ErrPointerOnly        = errors.New("orm: 只支持指针")
	ErrEmptyTableName     = errors.New("orm: 表名为空")
	ErrTooManyColumns     = errors.New("orm: 过多列")
	ErrNoRows             = errors.New("orm: 未找到数据")
	ErrInsertZeroRow      = errors.New("orm: 插入0行")
	ErrNonSupportOperator = errors.New("orm: set中不支持的操作")
)

func NewErrUnKnowField(name string) error {
	return fmt.Errorf("orm: 未知字段 %s", name)
}

func NewErrUnKnowColumn(name string) error {
	return fmt.Errorf("orm: 未知列 %s", name)
}
