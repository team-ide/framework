package util

import (
	"reflect"
	"strings"
	"sync"
)

var (
	structCache = &sync.Map{}
)

type StructInfo struct {
	StructType reflect.Type

	IsMap bool

	Fields []*StructField

	FieldMap   map[string]*StructField
	FieldLower map[string]*StructField
}

func (this_ *StructInfo) AddField(in *StructField) *StructInfo {
	if in == nil {
		return this_
	}
	if this_.FieldMap[in.FieldName] != nil {
		return this_
	}
	this_.Fields = append(this_.Fields, in)
	this_.FieldMap[in.FieldName] = in
	this_.FieldLower[strings.ToLower(in.FieldName)] = in

	return this_
}

type StructField struct {
	Field reflect.StructField
	Index int

	// 字段 是匿名 对象
	IsAnonymous bool
	// 匿名 对象 信息
	AnonymousModel *StructInfo

	// 是 匿名 对象 字段 这里放置 上层的 匿名  对象字段
	ParentFiled *StructField

	IsPtr bool
	Kind  reflect.Kind

	ElemType reflect.Type

	FieldName string

	IsString bool
	IsNumber bool
	IsBool   bool
}

func GetStructInfo(inType reflect.Type) (info *StructInfo) {
	var structType = inType
	for structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	if structType.Kind() == reflect.Map {
		info = &StructInfo{
			IsMap: true,
		}
		return
	}
	if cached, ok := structCache.Load(structType); ok {
		return cached.(*StructInfo)

	}
	info = LoadStructInfo(structType)

	structCache.Store(structType, info)
	return
}
func LoadStructInfo(structType reflect.Type) (info *StructInfo) {
	loadingCache := map[reflect.Type]*StructInfo{}
	info = _loadStructInfo(structType, loadingCache)
	return
}
func _loadStructInfo(structType reflect.Type, loadingCache map[reflect.Type]*StructInfo) (info *StructInfo) {
	info = loadingCache[structType]
	if info != nil {
		return
	}
	info = &StructInfo{}
	loadingCache[structType] = info
	info.StructType = structType

	info.FieldMap = map[string]*StructField{}
	info.FieldLower = map[string]*StructField{}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		structField := &StructField{
			Field: field,
			Index: i,
		}
		structField.FieldName = field.Name

		info.AddField(structField)

		structField.IsAnonymous = field.Anonymous
		structField.ElemType = field.Type
		structField.Kind = structField.ElemType.Kind()
		structField.IsPtr = structField.Kind == reflect.Ptr
		if structField.IsPtr {
			structField.ElemType = structField.ElemType.Elem()
			structField.Kind = structField.ElemType.Kind()
		}
		if structField.IsAnonymous &&
			structField.Kind == reflect.Struct &&
			structField.ElemType != structType {
			structField.AnonymousModel = _loadStructInfo(structField.ElemType, loadingCache)
			for _, subField := range structField.AnonymousModel.Fields {
				subField.ParentFiled = structField
				info.AddField(subField)
			}
		}
	}
	return
}

type FieldValue struct {
	value     reflect.Value
	valueType reflect.Type
	data      any
	isNull    bool
	isZero    bool
	isEmpty   bool
}

func (this_ *FieldValue) GetValueType() reflect.Type {
	return this_.valueType
}
func (this_ *FieldValue) GetValue() reflect.Value {
	return this_.value
}
func (this_ *FieldValue) GetData() any {
	return this_.data
}
func (this_ *FieldValue) IsNull() bool {
	return this_.isNull
}
func (this_ *FieldValue) IsZero() bool {
	return this_.isZero
}
func (this_ *FieldValue) IsEmpty() bool {
	return this_.isEmpty
}

func FieldValueByValue(v reflect.Value) (res *FieldValue) {
	res = &FieldValue{}
	res.value = v
	res.valueType = v.Type()
	if v.IsValid() {
		vKind := v.Kind()
		switch vKind {
		case reflect.String:
			res.isEmpty = v.String() == ""
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			res.isZero = v.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			res.isZero = v.Uint() == 0
		case reflect.Float32, reflect.Float64:
			res.isZero = v.Float() == 0
		case reflect.Complex64, reflect.Complex128:
			rc := real(v.Complex())
			ic := imag(v.Complex())
			res.isZero = rc == 0 && ic == 0
		case reflect.Invalid:
			res.isNull = true
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			res.isNull = v.IsNil()

		default:
		}
		res.data = v.Interface()
	}
	return
}

func FieldValueByData(data any) (res *FieldValue) {
	return FieldValueByValue(reflect.ValueOf(data))
}
