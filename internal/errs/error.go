package errs

import (
	"errors"
	"fmt"
)

var (
	ErrInputNil       = errors.New("orm: 不支持 nil")
	ErrPointerOnly    = errors.New("orm: 只支持指针")
	ErrEmptyTableName = errors.New("orm: 表名为空")
)

func NewErrUnKnowField(name string) error {
	return fmt.Errorf("orm: 未知字段 %s", name)
}
