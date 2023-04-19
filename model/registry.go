package model

import (
	"github.com/soluble1/morm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type Registry interface {
	Get(val any) (*Model, error)
	// Register 注册一个模型
	Register(val any, opts ...ModelOpt) (*Model, error)
}

type registry struct {
	models sync.Map
}

func NewRegistry() Registry {
	return &registry{}
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	m, err := r.Register(val)
	if err != nil {
		return nil, err
	}
	r.models.Store(typ, m)
	return m.(*Model), nil
}

func (r *registry) Register(val any, opts ...ModelOpt) (*Model, error) {
	if val == nil {
		return nil, errs.ErrInputNil
	}
	typ := reflect.TypeOf(val)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numField := typ.NumField()

	fieldMap := make(map[string]*Field, numField)
	colMap := make(map[string]*Field, numField)
	columns := make([]*Field, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		ormTagStrs := r.parseTag(fd.Tag)
		var colName string
		colName, ok := ormTagStrs["column"]
		if !ok || colName == "" {
			colName = underscoreName(fd.Name)
		}
		fdData := &Field{
			ColName: colName,
			Typ:     fd.Type,
			GoName:  fd.Name,
			Offset:  fd.Offset,
			Index:   fd.Index,
		}
		fieldMap[fd.Name] = fdData
		colMap[colName] = fdData
		columns[i] = fdData
	}

	var tableName string
	if tn, ok := val.(TableName); ok {
		tableName = tn.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	res := &Model{
		TableName: underscoreName(typ.Name()),
		FieldMap:  fieldMap,
		ColumnMap: colMap,
		Fields:    columns,
	}

	for _, opt := range opts {
		if err := opt(res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (r *registry) parseTag(tag reflect.StructTag) map[string]string {
	ormTag := tag.Get("orm")
	strs := strings.Split(ormTag, ",")
	res := make(map[string]string, len(strs))
	for _, str := range strs {
		segs := strings.Split(str, "=")
		key := segs[0]
		var val = ""
		if len(segs) > 1 {
			val = segs[1]
		}
		res[key] = val
	}
	return res
}

func underscoreName(name string) string {
	var buf []byte
	for i, v := range name {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
