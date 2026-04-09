package db

import (
	"reflect"
	"strings"
)

type SqlConfigurationSet func(c *SqlConfiguration)

func NewSqlConfiguration(sets ...SqlConfigurationSet) (c *SqlConfiguration) {
	c = &SqlConfiguration{}
	for _, set := range sets {
		set(c)
	}
	return
}

type SqlConfiguration struct {
	sqlHandler SqlHandler
}

func (this_ *SqlConfiguration) SetSqlHandler(sqlHandler SqlHandler) *SqlConfiguration {
	this_.sqlHandler = sqlHandler
	return this_
}
func (this_ *SqlConfiguration) GetSqlHandler() SqlHandler {
	return this_.sqlHandler
}

type SqlHandler interface {
	// RealTableName 获取真实 表名 ，SQL 中将使用返回的 表名 拼接 返回空 则不替换
	RealTableName(param *SqlParam, tableName string) (realTableName string)
	// WrapTableName 包装 表名 ，SQL 中将使用返回的 表名 拼接 返回空 则不替换
	WrapTableName(param *SqlParam, tableName string) (wrapTableName string)
	WrapColumnName(param *SqlParam, columnName string) (wrapColumnName string)
}

func NewMapSqlParam() (res *SqlParam) {
	var data = map[string]any{}
	res = SqlParamByMap(data)
	return
}
func SqlParamByMap(data map[string]any) (res *SqlParam) {
	res = &SqlParam{}
	res.param = data
	res.data = data
	return
}

func SqlParamByData(data any) (res *SqlParam) {
	value := reflect.ValueOf(data)
	res = SqlParamByValue(data, value)
	return
}

func SqlParamByValue(data any, value reflect.Value) (res *SqlParam) {
	res = &SqlParam{}

	res.data = data
	res.value = value
	return
}

type SqlParam struct {

	// 是 读取
	isQuery bool
	// 是 插入
	isInsert bool
	// 是 插入
	isUpdate bool
	// 是 插入
	isDelete bool
	// 是 插入
	isDDL bool

	param map[string]any

	data  any
	value reflect.Value
}

func (this_ *SqlParam) GetData() any {
	return this_.data
}

func (this_ *SqlParam) Set(name string, value any) *SqlParam {
	param := this_.param
	if param == nil {
		param = map[string]any{}
		this_.param = param
	}
	param[name] = value
	return this_
}

func (this_ *SqlParam) GetStringParam(name string) (res string) {
	param := this_.param
	if param == nil {
		return
	}
	v, ok := param[name]
	if !ok {
		return
	}
	res = GetStringValue(v)
	return
}

func (this_ *SqlParam) GetParam(name string) (res any) {
	param := this_.param
	if param == nil {
		res = this_.data
		return
	}
	res, _ = param[name]
	return
}
func (this_ *SqlParam) GetObjectField(obj any, name string) (res any) {
	if obj == nil {
		return
	}
	objV := reflect.ValueOf(obj)
	if objV.Kind() == reflect.Ptr {
		objV = objV.Elem()
	}
	if objV.Kind() == reflect.Map {
		fieldV := objV.MapIndex(reflect.ValueOf(name))
		if fieldV.IsValid() {
			res = fieldV.Interface()
		}
	} else if objV.Kind() == reflect.Struct {
		fieldV := objV.FieldByName(name)
		if fieldV.IsValid() {
			res = fieldV.Interface()
		}
	} else if objV.Kind() == reflect.String {
		if name == "length" || name == "size" {
			res = len(objV.String())
		}
	} else if objV.Kind() == reflect.Array || objV.Kind() == reflect.Slice {
		if name == "length" || name == "size" {
			res = objV.Len()
		}
	}
	return
}

func (this_ *SqlParam) FindParam(name string) (v any, find bool) {
	param := this_.param
	if param == nil {
		return
	}
	v, find = param[name]
	return
}

var (
	DefaultSqlOption = NewSqlOption()
)

func NewSqlOption() (res *SqlOption) {
	res = &SqlOption{}
	return res
}

type SqlOption struct {
	WrapOption
}

func (this_ *SqlOption) Set(fn func(o *SqlOption)) *SqlOption {
	fn(this_)
	return this_
}
func (this_ *SqlOption) RealTableName(param *SqlParam, tableName string) (realTableName string) {
	return tableName
}
func (this_ *SqlOption) WrapTableName(param *SqlParam, tableName string) (wrapTableName string) {
	return this_.WrapOption.WrapTableName(tableName)
}
func (this_ *SqlOption) WrapColumnName(param *SqlParam, columnName string) (wrapColumnName string) {
	return this_.WrapOption.WrapColumnName(columnName)
}

func NewSqlTemplate(sqlList ...string) (t *SqlTemplate) {
	t = &SqlTemplate{}
	t.sqlList = append(t.sqlList, sqlList...)
	return
}
func SqlTemplateHasContent(t *SqlTemplate) bool {
	if t == nil {
		return false
	}
	s := strings.TrimSpace(t.GetTemplateSql())
	if s == "" {
		return false
	}
	return true
}

func (this_ *SqlTemplate) Append(sqlList ...string) *SqlTemplate {
	if len(sqlList) == 0 {
		return this_
	}
	this_.sqlList = append(this_.sqlList, sqlList...)
	this_.SqlList = nil
	return this_
}

func (this_ *SqlTemplate) GetSql(handler SqlHandler, param map[string]any) (sqlInfo string, sqlArgs []any) {
	for _, one := range this_.SqlList {
		sqlInfo, sqlArgs = one.GetSqlByMap(handler, param)
		if len(sqlInfo) > 0 {
			return
		}
	}
	return
}
func (this_ *SqlTemplate) GetSqlList(handler SqlHandler, param map[string]any) (sqlList []string, sqlArgsList [][]any) {
	if this_ == nil {
		return
	}
	var sqlInfo string
	var sqlArgs []any
	for _, one := range this_.SqlList {
		sqlInfo, sqlArgs = one.GetSqlByMap(handler, param)
		if len(sqlInfo) > 0 {
			sqlList = append(sqlList, sqlInfo)
			sqlArgsList = append(sqlArgsList, sqlArgs)
		}
	}
	return
}
func (this_ *SqlNodeSql) GetSqlByMap(handler SqlHandler, data map[string]any) (sqlInfo string, args []any) {
	b := &SqlBuilder{}
	b.SqlParam = SqlParamByMap(data)
	b.handler = handler
	this_.Append(b)

	sqlInfo = strings.TrimSpace(b.sqlInfo)
	args = b.sqlArgs
	return
}

// SqlTemplate SQL 模板语句
// 将 SQL 解析为 普通语句、参数替换语句、参数占位语句、if 语句 等
// ${xx}: 参数占位语句，拼接SQL 为 ? 然后拼接参数
// #{xx}和{xx}: 参数替换语句，直接将 参数 拼接SQL
// {xx}: 参数替换语句，直接将 参数 替换
// if(xx){xxx}: 条件
// [ss {x}]: BracketOptional = true，非必须，有参数情况，如果有 x 这个参数 则拼接
// [ss ss dfd]: BracketOptional = true，非必须，无参数情况，如果参数有 ssSsDfd 则拼接
type SqlTemplate struct {
	sqlList    []string
	SqlList    []*SqlNodeSql
	sqlHandler SqlHandler

	// 中括号 可选 内容 是否
	BracketOptional bool

	// 设置 Debug 表示输出日志
	Debug bool
}

func (this_ *SqlTemplate) SetSqlHandler(sqlHandler SqlHandler) *SqlTemplate {
	this_.sqlHandler = sqlHandler
	return this_
}
func (this_ *SqlTemplate) GetSqlHandler() SqlHandler {
	return this_.sqlHandler
}

type SqlBuilder struct {
	sqlInfo string
	sqlArgs []any

	handler SqlHandler
	*SqlParam
}

func (this_ *SqlBuilder) Append(s string, args ...any) {
	this_.sqlInfo += s
	this_.sqlArgs = append(this_.sqlArgs, args...)
}

func (this_ *SqlBuilder) Test(test SqlExpression) (res bool) {
	if test == nil {
		return
	}
	v := test.GetValue(this_)
	res = IsTrue(v)
	//fmt.Println("run test type:", reflect.TypeOf(test))
	//fmt.Println("run test json:", GetStringValue(test))
	//fmt.Println("run test value:", v)
	return
}

func (this_ *SqlBuilder) GetValue(e SqlExpression) (res any) {
	if e == nil {
		return
	}
	res = e.GetValue(this_)
	return
}
