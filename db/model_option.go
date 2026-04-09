package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	DefaultModelOption = NewModelOption()
)

func NewModelOption() (res *ModelOption) {
	res = &ModelOption{}
	return
}

type ModelOption struct {
	// 字段 tag 注解，默认 column 如：`column:"xx"`
	getColumn FiledGetColumn
	// 字段 tag 注解，默认 column 如：`column:"xx"`
	ColumnTag string `json:"columnTag,omitempty"`
	// 是否 使用json主键，如果 字段 tag 注解 未找到 使用 json 注解
	ColumnUseJsonTag bool `json:"columnUseJsonTag,omitempty"`
	// 如果 以上都未配置 使用 字段名称
	ColumnUseFieldName bool `json:"columnUseFieldName,omitempty"`

	// 字段名称 严格判断 默认 false `userId 将 匹配 Userid UserId USERID`，
	ColumnStrictCase bool `json:"columnStrictCase,omitempty"`

	getValue FiledGetValue

	modelCache sync.Map
}

type FiledGetColumn func() string
type FiledGetValue func(columnName string, columnValue *FieldValue) any

func (this_ *ModelOption) SetGetColumn(getColumn FiledGetColumn) *ModelOption {
	this_.getColumn = getColumn
	return this_
}
func (this_ *ModelOption) SetGetValue(getValue FiledGetValue) *ModelOption {
	this_.getValue = getValue
	return this_
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

func (this_ *ModelOption) GetColumnValues(modelV reflect.Value) (columns []string, values []*FieldValue) {
	modelType := modelV.Type()
	modelInfo := this_.GetModelInfo(modelType)

	for modelV.Kind() == reflect.Ptr {
		modelV = modelV.Elem()
	}

	if modelInfo.isMap {
		// 直接 解析 key value

		for _, kV := range modelV.MapKeys() {
			// key 必须是 string
			if kV.Type().Kind() != reflect.String {
				continue
			}
			columnName := kV.String()
			vV := modelV.MapIndex(kV)

			columns = append(columns, columnName)
			fieldValue := FieldValueByValue(vV)
			values = append(values, fieldValue)
		}
		return
	}
	for _, column := range modelInfo.columns {
		var filedV reflect.Value
		if column.parentFiled == nil {
			filedV = modelV.Field(column.Index)
		} else {
			//if column.parentFiled.IsPtr {
			//	parentField := modelV.Field(column.parentFiled.Index)
			//	if parentField.IsNil() {
			//		continue
			//	}
			//}
			filedV = modelV.FieldByName(column.FieldName)
		}
		columns = append(columns, column.ColumnName)

		fieldValue := FieldValueByValue(filedV)
		values = append(values, fieldValue)
	}
	return
}

func (this_ *ModelOption) GetColumns(modelV reflect.Value) (columns []string) {
	modelType := modelV.Type()
	modelInfo := this_.GetModelInfo(modelType)

	for modelV.Kind() == reflect.Ptr {
		modelV = modelV.Elem()
	}

	if modelInfo.isMap {
		// 直接 解析 key value

		for _, kV := range modelV.MapKeys() {
			// key 必须是 string
			if kV.Type().Kind() != reflect.String {
				continue
			}
			columnName := kV.String()
			columns = append(columns, columnName)
		}
		return
	}
	for _, column := range modelInfo.columns {
		columns = append(columns, column.ColumnName)
	}
	return
}

func (this_ *ModelOption) GetColumnValue(modelV reflect.Value, columnName string) (fieldValue *FieldValue) {
	modelType := modelV.Type()
	modelInfo := this_.GetModelInfo(modelType)

	for modelV.Kind() == reflect.Ptr {
		modelV = modelV.Elem()
	}

	if modelInfo.isMap {
		// 直接 解析 key value
		vV := modelV.MapIndex(reflect.ValueOf(columnName))

		fieldValue = FieldValueByValue(vV)
		return
	}
	for _, column := range modelInfo.columns {
		if column.ColumnName == columnName {

			var filedV reflect.Value
			if column.parentFiled == nil {
				filedV = modelV.Field(column.Index)
			} else {
				//if column.parentFiled.IsPtr {
				//	parentField := modelV.Field(column.parentFiled.Index)
				//	if parentField.IsNil() {
				//		continue
				//	}
				//}
				filedV = modelV.FieldByName(column.FieldName)
			}
			fieldValue = FieldValueByValue(filedV)
			return
		}
	}
	return
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

type ModelInfo struct {
	modelType reflect.Type

	isMap bool

	fields []*ModelField

	columns     []*ModelField
	columnMap   map[string]*ModelField
	columnLower map[string]*ModelField
	fieldMap    map[string]*ModelField
	fieldLower  map[string]*ModelField
	jsonMap     map[string]*ModelField
	jsonLower   map[string]*ModelField
}

type ModelField struct {
	Field reflect.StructField
	Index int

	// 字段 是匿名 对象
	IsAnonymous bool
	// 匿名 对象 信息
	AnonymousModel *ModelInfo

	// 是 匿名 对象 字段 这里放置 上层的 匿名  对象字段
	parentFiled *ModelField

	IsPtr bool
	Kind  reflect.Kind

	elemType     reflect.Type
	sqlValueType reflect.Type

	// 是否 实现了 sql.Scanner 接口
	ImplementsSqlScanner bool

	FieldName  string
	ColumnName string
	JsonName   string

	IsString bool
	IsNumber bool
	IsBool   bool

	fieldSetter FieldSetterType
}

type FieldSetterType func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error)

var (
	timeType         = reflect.TypeOf(time.Time{})
	sqlScanner       sql.Scanner
	sqlScannerType   = reflect.TypeOf(&sqlScanner).Elem()
	driverValuer     driver.Valuer
	driverValuerType = reflect.TypeOf(&driverValuer).Elem()
)

func (this_ *ModelOption) GetModelInfo(inType reflect.Type) (info *ModelInfo) {
	var modelType = inType
	for modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() == reflect.Map {
		info = &ModelInfo{
			isMap: true,
		}
		return
	}
	if cached, ok := this_.modelCache.Load(modelType); ok {
		return cached.(*ModelInfo)

	}
	loadingCache := map[reflect.Type]*ModelInfo{}
	info = this_.loadModelInfo(modelType, loadingCache)

	this_.modelCache.Store(modelType, info)
	return
}

func (this_ *ModelOption) loadModelInfo(modelType reflect.Type, loadingCache map[reflect.Type]*ModelInfo) (info *ModelInfo) {
	info = loadingCache[modelType]
	if info != nil {
		return
	}
	info = &ModelInfo{}
	info.modelType = modelType
	info.columnMap = map[string]*ModelField{}
	info.columnLower = map[string]*ModelField{}

	info.fieldMap = map[string]*ModelField{}
	info.fieldLower = map[string]*ModelField{}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		modelField := &ModelField{
			Field: field,
			Index: i,
		}
		modelField.FieldName = field.Name

		info.fields = append(info.fields, modelField)
		info.fieldMap[modelField.FieldName] = modelField
		info.fieldLower[strings.ToLower(modelField.FieldName)] = modelField

		var str string
		var columnName string
		var tag = this_.ColumnTag
		if tag == "" {
			tag = "column"
		}
		str = field.Tag.Get(tag)
		if str == "" && this_.ColumnUseJsonTag {
			str = field.Tag.Get("json")
		}
		if str == "" && this_.ColumnUseFieldName {
			str = field.Name
		}
		if str != "" && str != "-" {
			ss := strings.Split(str, ",")
			columnName = ss[0]
		}

		modelField.IsAnonymous = field.Anonymous
		modelField.elemType = field.Type
		modelField.Kind = modelField.elemType.Kind()
		modelField.IsPtr = modelField.Kind == reflect.Ptr
		if modelField.IsPtr {
			modelField.elemType = modelField.elemType.Elem()
			modelField.Kind = modelField.elemType.Kind()
		}

		modelField.ColumnName = columnName
		if modelField.ColumnName != "" {

			info.columns = append(info.columns, modelField)
			info.columnMap[columnName] = modelField
			info.columnLower[strings.ToLower(columnName)] = modelField

			// 检查是否实现了sql.Scanner
			modelField.ImplementsSqlScanner = field.Type.Implements(sqlScannerType)
			//fmt.Println("Field:", modelField)
			switch modelField.Kind {
			case reflect.String:
				var val sql.NullString
				modelField.sqlValueType = reflect.TypeOf(val)
				modelField.IsString = true
				modelField.fieldSetter = fieldSetterString
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				var val sql.NullInt64
				modelField.sqlValueType = reflect.TypeOf(val)
				modelField.IsNumber = true
				modelField.fieldSetter = fieldSetterInt64
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				var val sql.NullInt64
				modelField.sqlValueType = reflect.TypeOf(val)
				modelField.IsNumber = true
				modelField.fieldSetter = fieldSetterInt64
			case reflect.Float32, reflect.Float64:
				var val sql.NullFloat64
				modelField.sqlValueType = reflect.TypeOf(val)
				modelField.IsNumber = true
				modelField.fieldSetter = fieldSetterFloat64
			case reflect.Bool:
				var val sql.NullBool
				modelField.sqlValueType = reflect.TypeOf(val)
				modelField.IsBool = true
				modelField.fieldSetter = fieldSetterBool
			case reflect.Struct:
				// 检查是否是time.Time
				if modelField.elemType == timeType {
					var val sql.NullTime
					modelField.sqlValueType = reflect.TypeOf(val)
					modelField.fieldSetter = fieldSetterTime
				} else {
					// 其他结构体类型
					modelField.fieldSetter = fieldSetterOther
				}
			default:
				// 其他类型
				modelField.fieldSetter = fieldSetterOther
			}

		}
	}
	loadingCache[modelType] = info
	for _, field := range info.fields {
		if field.IsAnonymous && field.Kind == reflect.Struct {
			field.AnonymousModel = this_.loadModelInfo(field.elemType, loadingCache)
			for _, subField := range field.AnonymousModel.fields {
				subField.parentFiled = field
				if info.fieldMap[subField.FieldName] != nil {
					continue
				}
				if subField.ColumnName != "" {
					if info.columnLower[strings.ToLower(subField.ColumnName)] != nil {
						continue
					}

					info.columns = append(info.columns, subField)
					info.columnMap[subField.ColumnName] = subField
					info.columnLower[strings.ToLower(subField.ColumnName)] = subField
				}
			}
		}
	}
	return
}

var fieldSetterScanner FieldSetterType = func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error) {
	if receiver == nil {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}
	//fmt.Println("fieldSetterScanner:", receiver)
	value := reflect.ValueOf(receiver)
	if modelField.IsPtr {
		fieldValue.Set(value)
	} else {
		fieldValue.Set(value.Elem())
	}
	return
}
var fieldSetterString FieldSetterType = func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error) {
	if receiver == nil {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}
	nullStr := receiver.(*sql.NullString)
	if !nullStr.Valid {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}
	return setFieldValue(fieldValue, nullStr.String, modelField.IsPtr)
}
var fieldSetterInt64 FieldSetterType = func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error) {
	if receiver == nil {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	nullInt := receiver.(*sql.NullInt64)
	if !nullInt.Valid {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	return setFieldValue(fieldValue, nullInt.Int64, modelField.IsPtr)
}
var fieldSetterFloat64 FieldSetterType = func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error) {
	if receiver == nil {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	nullFloat := receiver.(*sql.NullFloat64)
	if !nullFloat.Valid {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	return setFieldValue(fieldValue, nullFloat.Float64, modelField.IsPtr)
}

var fieldSetterBool FieldSetterType = func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error) {
	if receiver == nil {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	nullBool := receiver.(*sql.NullBool)
	if !nullBool.Valid {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	return setFieldValue(fieldValue, nullBool.Bool, modelField.IsPtr)
}

var fieldSetterTime FieldSetterType = func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error) {
	if receiver == nil {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	nullTime := receiver.(*sql.NullTime)
	if !nullTime.Valid {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return
	}

	return setFieldValue(fieldValue, nullTime.Time, modelField.IsPtr)
}

var fieldSetterOther FieldSetterType = func(model reflect.Value, modelField *ModelField, fieldValue reflect.Value, receiver any) (err error) {
	if receiver == nil {
		if modelField.IsPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return nil
	}

	val := reflect.ValueOf(receiver).Elem().Interface()
	return setFieldValue(fieldValue, val, modelField.IsPtr)
}

// 通用的字段赋值函数
func setFieldValue(fieldValue reflect.Value, val interface{}, isPtr bool) error {
	if val == nil {
		if isPtr {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return nil
	}

	// 获取实际值
	actualVal := reflect.ValueOf(val)

	// 处理指针字段
	if isPtr {
		ptrType := fieldValue.Type()
		elemType := ptrType.Elem()

		// 创建新指针
		newPtr := reflect.New(elemType)

		// 赋值给指针指向的值
		if err := setValueToField(newPtr.Elem(), actualVal); err != nil {
			return err
		}

		fieldValue.Set(newPtr)
		return nil
	}

	// 非指针字段直接赋值
	return setValueToField(fieldValue, actualVal)
}

// 将值设置到字段（处理类型转换）
func setValueToField(field reflect.Value, val reflect.Value) error {
	fieldType := field.Type()

	// 如果类型相同或可赋值，直接赋值
	if val.Type().AssignableTo(fieldType) {
		field.Set(val)
		return nil
	}

	// 类型转换
	if val.Type().ConvertibleTo(fieldType) {
		field.Set(val.Convert(fieldType))
		return nil
	}

	// 特殊类型转换处理
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprint(val.Interface()))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := toInt64(val); err == nil {
			field.SetInt(intVal)
		} else {
			return fmt.Errorf("cannot convert %v to %v", val.Type(), fieldType)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if intVal, err := toInt64(val); err == nil && intVal >= 0 {
			field.SetUint(uint64(intVal))
		} else {
			return fmt.Errorf("cannot convert %v to %v", val.Type(), fieldType)
		}

	case reflect.Float32, reflect.Float64:
		if floatVal, err := toFloat64(val); err == nil {
			field.SetFloat(floatVal)
		} else {
			return fmt.Errorf("cannot convert %v to %v", val.Type(), fieldType)
		}

	case reflect.Bool:
		if boolVal, err := toBool(val); err == nil {
			field.SetBool(boolVal)
		} else {
			return fmt.Errorf("cannot convert %v to bool", val.Type())
		}

	default:
		return fmt.Errorf("unsupported type conversion: %v to %v", val.Type(), fieldType)
	}

	return nil
}

// 辅助函数：转换为int64
func toInt64(val reflect.Value) (int64, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(val.Float()), nil
	case reflect.String:
		return strconv.ParseInt(val.String(), 10, 64)
	case reflect.Bool:
		if val.Bool() {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %v to int64", val.Type())
	}
}

// 辅助函数：转换为float64
func toFloat64(val reflect.Value) (float64, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	case reflect.String:
		return strconv.ParseFloat(val.String(), 64)
	default:
		return 0, fmt.Errorf("cannot convert %v to float64", val.Type())
	}
}

// 辅助函数：转换为bool
func toBool(val reflect.Value) (bool, error) {
	switch val.Kind() {
	case reflect.Bool:
		return val.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() != 0, nil
	case reflect.String:
		return strconv.ParseBool(val.String())
	default:
		return false, fmt.Errorf("cannot convert %v to bool", val.Type())
	}
}
