package db

import (
	"errors"
	"github.com/team-ide/framework/util"
	"reflect"
	"strings"
)

func NewModelSelect(model any) (res *ModelSelect) {
	res = &ModelSelect{}
	res.model = model
	res.ModelSetting = &ModelSetting{}
	return
}

type ModelSelect struct {
	// 如果设置了 model 根据属性值查询
	model any
	// 查询条件
	where *Conditions

	// 查询时候 必须 设置条件 除非 设置 SelectAll
	canSelectAll bool
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

// CanSelectAll 查询时候 必须 设置条件 除非 设置 CanSelectAll
func (this_ *ModelSelect) CanSelectAll() *ModelSelect {
	this_.canSelectAll = true
	return this_
}
func (this_ *ModelSelect) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *ModelSelect) SetWhere(where *Conditions) *ModelSelect {
	this_.where = where
	return this_
}

func (this_ *ModelSelect) GetSql() (sqlInfo string, args []any, err error) {

	b := this_.NewBuilder(this_.model)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}
	var columns []string
	var includeColumns = this_.selectIncludeColumns
	// 如果 有 包含的字段 则只查询 包含的字段
	if len(includeColumns) > 0 {
		if b.model != nil {
			modelColumns := b.GetColumns(b.modelValue)
			for _, column := range modelColumns {
				if this_.IsSelectInclude(column) {
					columns = append(columns, column)
				}
			}
		} else {
			columns = append(columns, includeColumns...)
		}
	} else {
		var excludeColumns = this_.selectExcludeColumns
		// 如果 有 排除的字段 则只查询 未排除的字段
		if len(excludeColumns) > 0 {
			if b.model != nil {
				modelColumns := b.GetColumns(b.modelValue)
				for _, column := range modelColumns {
					if !this_.IsSelectExclude(column) {
						columns = append(columns, column)
					}
				}
			}
		}
	}
	var selectColumnStr string
	if len(columns) > 0 {
		var addColumns int
		for _, column := range columns {
			wrapColumn := b.WrapColumnName(b.sqlParam, column)
			if wrapColumn == "" {
				continue
			}
			if addColumns > 0 {
				selectColumnStr += ", "
			}
			addColumns++
			selectColumnStr += wrapColumn
		}
	} else {
		selectColumnStr = "*"
	}

	sqlInfo += "SELECT " + selectColumnStr + " FROM " + b.wrapTableName + ""

	whereSql, whereArgs := this_.GetModelAndWhereSql(b, this_.model, this_.where)
	if len(whereSql) == 0 {
		if !this_.canSelectAll {
			err = errors.New("select sql 必须设置条件 或者 调下 CanSelectAll()")
			return
		}
	} else {
		sqlInfo += " WHERE " + whereSql
		args = append(args, whereArgs...)
	}
	appendSql, appendArgs := this_.GetAppendSql()
	sqlInfo += " " + appendSql
	args = append(args, appendArgs...)
	return
}
func NewModelCount(model any) (res *ModelCount) {
	res = &ModelCount{}
	res.model = model
	res.ModelSetting = &ModelSetting{}
	return
}

type ModelCount struct {
	// 如果设置了 model 根据属性值查询
	model any
	// 查询条件
	where *Conditions

	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

func (this_ *ModelCount) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *ModelCount) SetWhere(where *Conditions) *ModelCount {
	this_.where = where
	return this_
}

func (this_ *ModelCount) GetSql() (sqlInfo string, args []any, err error) {

	b := this_.NewBuilder(this_.model)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "SELECT COUNT(1) FROM " + b.wrapTableName + ""

	whereSql, whereArgs := this_.GetModelAndWhereSql(b, this_.model, this_.where)
	if whereSql != "" {
		sqlInfo += " WHERE " + whereSql
		args = append(args, whereArgs...)
	}
	appendSql, appendArgs := this_.GetAppendSql()
	sqlInfo += " " + appendSql
	args = append(args, appendArgs...)
	return
}

func NewModelInsert(model any) (res *ModelInsert) {
	res = &ModelInsert{}
	res.model = model
	res.ModelSetting = &ModelSetting{}
	return
}

type ModelInsert struct {
	// 需要保存的 model
	model any
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

func (this_ *ModelInsert) GetSql() (sqlInfo string, args []any, err error) {
	var model = this_.model
	if model == nil {
		err = errors.New("insert model is null")
		return
	}
	b := this_.NewBuilder(model)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "INSERT INTO " + b.wrapTableName + " "

	columns, values := b.GetColumnValues(b.modelValue)
	var wrapColumns []string
	var wrapValues []string
	for i, column := range columns {
		isPrimaryKey := b.IsPrimaryKey(column)
		if !isPrimaryKey {
			if !this_.Included(IncludedPlaceValue, column, values[i]) {
				continue
			}
		}
		wrapColumn := b.WrapColumnName(b.sqlParam, column)
		if wrapColumn == "" {
			continue
		}
		wrapColumns = append(wrapColumns, wrapColumn)
		fieldValue := values[i]
		// 如果是 主键且无值 则设置为 null
		if isPrimaryKey && (fieldValue.IsNull() || fieldValue.IsZero() || fieldValue.IsEmpty()) {
			wrapValues = append(wrapValues, "NULL")
		} else {
			isNull, v := this_.GetValue(values[i])
			if isNull {
				wrapValues = append(wrapValues, "NULL")
			} else {
				wrapValues = append(wrapValues, "?")
				args = append(args, v)
			}
		}
	}
	if len(wrapColumns) == 0 {
		err = errors.New("insert sql columns is empty")
		return
	}
	sqlInfo += "(" + strings.Join(wrapColumns, ", ") + ") "
	sqlInfo += "VALUES (" + strings.Join(wrapValues, ", ") + ")"

	return
}

func NewModelUpdate(model any) (res *ModelUpdate) {
	res = &ModelUpdate{}
	res.model = model
	res.ModelSetting = &ModelSetting{}
	return
}

type ModelUpdate struct {
	// 需要更新的 model 只会根据 属性值设置更新 不根据属性值查询
	model any
	// 更新条件
	where *Conditions

	// 更新时候 必须 设置条件 除非 设置 UpdateAll
	canUpdateAll bool
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

// CanUpdateAll 更新时候 必须 设置条件 除非 设置 CanUpdateAll
func (this_ *ModelUpdate) CanUpdateAll() *ModelUpdate {
	this_.canUpdateAll = true
	return this_
}
func (this_ *ModelUpdate) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *ModelUpdate) SetWhere(where *Conditions) *ModelUpdate {
	this_.where = where
	return this_
}

func (this_ *ModelUpdate) GetSql() (sqlInfo string, args []any, err error) {
	var model = this_.model
	if model == nil {
		err = errors.New("update model is null")
		return
	}
	b := this_.NewBuilder(model)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "UPDATE " + b.wrapTableName + " SET"

	columns, values := b.GetColumnValues(b.modelValue)
	var wrapColumns []string
	var wrapValues []*util.FieldValue
	for i, column := range columns {
		if !this_.Included(IncludedPlaceValue, column, values[i]) {
			continue
		}
		wrapColumn := b.WrapColumnName(b.sqlParam, column)
		if wrapColumn == "" {
			continue
		}
		wrapColumns = append(wrapColumns, wrapColumn)
		wrapValues = append(wrapValues, values[i])
	}
	if len(wrapColumns) == 0 {
		err = errors.New("update sql columns is empty")
		return
	}
	for i, wrapColumn := range wrapColumns {
		if i > 0 {
			sqlInfo += ","
		}
		isNull, v := this_.GetValue(wrapValues[i])
		if isNull {
			sqlInfo += " " + wrapColumn + " = NULL"
		} else {
			sqlInfo += " " + wrapColumn + " = ?"
			args = append(args, v)
		}
	}
	var whereSql string
	var whereArgs []any
	where := this_.where
	if where != nil {
		whereSql, whereArgs = where.Build(b, this_.service)
	}
	if len(whereSql) == 0 {
		if !this_.canUpdateAll {
			err = errors.New("update sql 必须设置条件 或者 调下 UpdateAll()")
			return
		}
	} else {
		sqlInfo += " WHERE " + whereSql
		args = append(args, whereArgs...)
	}
	appendSql, appendArgs := this_.GetAppendSql()
	sqlInfo += " " + appendSql
	args = append(args, appendArgs...)
	return
}

func NewModelDelete(model any) (res *ModelDelete) {
	res = &ModelDelete{}
	res.model = model
	res.ModelSetting = &ModelSetting{}
	return
}

type ModelDelete struct {
	// 如果设置了 model 根据属性值查询
	model any
	// 删除条件 如果和 model 同时设置 则会拼接条件
	where *Conditions

	// 删除时候 必须 设置条件 除非 设置 DeleteAll
	canDeleteAll bool
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

// CanDeleteAll 删除时候 必须 设置条件 除非 设置 CanDeleteAll
func (this_ *ModelDelete) CanDeleteAll() *ModelDelete {
	this_.canDeleteAll = true
	return this_
}
func (this_ *ModelDelete) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *ModelDelete) SetWhere(where *Conditions) *ModelDelete {
	this_.where = where
	return this_
}

func (this_ *ModelDelete) GetSql() (sqlInfo string, args []any, err error) {

	b := this_.NewBuilder(this_.model)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "DELETE FROM " + b.wrapTableName + ""

	whereSql, whereArgs := this_.GetModelAndWhereSql(b, this_.model, this_.where)
	if len(whereSql) == 0 {
		if !this_.canDeleteAll {
			err = errors.New("delete sql 必须设置条件 或者 调下 DeleteAll()")
			return
		}
	} else {
		sqlInfo += " WHERE " + whereSql
		args = append(args, whereArgs...)
	}

	appendSql, appendArgs := this_.GetAppendSql()
	sqlInfo += " " + appendSql
	args = append(args, appendArgs...)
	return
}

func (this_ *ModelSetting) GetModelAndWhereSql(b *OrmSqlBuilder, model any, where *Conditions) (whereSql string, whereArgs []any) {

	var newWhere = NewConditions()
	if model != nil {
		columns, values := b.GetColumnValues(b.modelValue)
		for i, column := range columns {
			// 作为 条件 不取根据字段名称过滤  只根据值过滤
			if !this_.Included(IncludedPlaceWhere, "", values[i]) {
				continue
			}
			wrapColumn := b.WrapColumnName(b.sqlParam, column)
			if wrapColumn == "" {
				continue
			}
			o := b.whereOperators[strings.ToLower(column)]
			isNull, v := this_.GetValue(values[i])
			if o != nil {
				switch strings.ToLower(o.Operator) {
				case "%like%":
					newWhere.Like(wrapColumn, &SqlConcatValue{
						Values: []string{"%", "?", "%"},
						Value:  v,
					})
				case "%like":
					newWhere.Like(wrapColumn, &SqlConcatValue{
						Values: []string{"%", "?"},
						Value:  v,
					})
				case "like":
					newWhere.Like(wrapColumn, v)
				default:
					if isNull {
						newWhere.IsNull(wrapColumn)
					} else {
						newWhere.Eq(wrapColumn, v)
					}
				}
			} else {
				if isNull {
					newWhere.IsNull(wrapColumn)
				} else {
					newWhere.Eq(wrapColumn, v)
				}
			}
		}
	}
	if where != nil {
		newWhere.AndGroup(where)
	}
	whereSql, whereArgs = newWhere.Build(b, this_.service)
	if len(whereSql) == 0 {
		return
	}
	return
}

type ModelSetting struct {
	// 配置表名 如果没有配置 需要 model 实现 xx 接口
	tableName string
	// 配置主键 如果没有配置 需要 model
	primaryKey []string

	// 是否 包含 0，默认 忽略 0 值
	includeZero bool
	// 是否 包含 空字符串，默认 忽略 空字符串 值
	includeEmpty bool
	// 是否 包含 空，默认 忽略 null 值
	includeNull bool

	// 查询 包含的字段 指的是 查询 字段
	selectIncludeColumns   []string
	selectIncludeColumnStr string
	// 查询 排除的字段 指的是 查询 字段
	selectExcludeColumns   []string
	selectExcludeColumnStr string

	// 值 包含的字段 指的是 新增、更新 字段
	valueIncludeColumns   []string
	valueIncludeColumnStr string
	// 值 排除的字段 指的是 新增、更新 字段
	valueExcludeColumns   []string
	valueExcludeColumnStr string

	// 条件 包含的字段 指的是 条件 字段
	whereIncludeColumns   []string
	whereIncludeColumnStr string
	// 条件 排除的字段 指的是 条件 字段
	whereExcludeColumns   []string
	whereExcludeColumnStr string

	whereOperators []*WhereOperator

	// 是否将 0 值 设置为 null，需要 IncludeZero = true
	zeroUseNull bool
	// 是否将 空字符串 值 设置为 null，需要 IncludeEmpty = true
	emptyUseNull bool

	service *Service

	sqlHandler SqlHandler

	modelOption *ModelOption

	appendSql  []string
	appendArgs []any
}

type IModel interface {
	IGetTableName
	IGetPrimaryKey
}

type IGetTableName interface {
	GetTableName() string
}
type IGetPrimaryKey interface {
	GetPrimaryKey() []string
}

func (this_ *ModelSetting) GetAppendSql() (appendSql string, appendArgs []any) {
	appendSql = strings.Join(this_.appendSql, " ")
	appendArgs = this_.appendArgs
	return
}
func (this_ *ModelSetting) AppendSql(appendSql string, appendArgs ...any) *ModelSetting {
	if appendSql != "" {
		this_.appendSql = append(this_.appendSql, appendSql)
	}
	this_.appendArgs = append(this_.appendArgs, appendArgs...)
	return this_
}
func (this_ *ModelSetting) SetTableName(tableName string) *ModelSetting {
	this_.tableName = tableName
	return this_
}

func (this_ *ModelSetting) GetTableName(model any) string {
	if this_.tableName != "" {
		return this_.tableName
	}
	if model != nil {
		g, ok := model.(IGetTableName)
		if ok {
			return g.GetTableName()
		}
	}
	return ""
}

func (this_ *ModelSetting) SetPrimaryKey(columns ...string) *ModelSetting {
	this_.primaryKey = columns
	return this_
}

func (this_ *ModelSetting) GetPrimaryKey(model any) []string {
	if len(this_.primaryKey) > 0 {
		return this_.primaryKey
	}
	if model != nil {
		g, ok := model.(IGetPrimaryKey)
		if ok {
			return g.GetPrimaryKey()
		}
	}
	return []string{}
}

// IncludeNull 设置 包含 null，默认 忽略 null 值
func (this_ *ModelSetting) IncludeNull() *ModelSetting {
	this_.includeNull = true
	return this_
}

// IncludeZero 设置 包含 0，默认 忽略 0 值
func (this_ *ModelSetting) IncludeZero() *ModelSetting {
	this_.includeZero = true
	return this_
}

// IncludeEmpty 设置 包含 空字符串，默认 忽略 空字符串 值
func (this_ *ModelSetting) IncludeEmpty() *ModelSetting {
	this_.includeEmpty = true
	return this_
}

// ZeroUseNull 将 0 值 设置为 null，需要 先设置 IncludeZero
func (this_ *ModelSetting) ZeroUseNull() *ModelSetting {
	this_.zeroUseNull = true
	return this_
}

// EmptyUseNull 将 空字符串 值 设置为 null，需要 先设置 IncludeEmpty
func (this_ *ModelSetting) EmptyUseNull() *ModelSetting {
	this_.emptyUseNull = true
	return this_
}

func (this_ *ModelSetting) columnsEmpty(columns *[]string, columnStr *string) *ModelSetting {
	*columns = []string{}
	*columnStr = ""
	return this_
}
func (this_ *ModelSetting) columnsAdd(columns *[]string, columnStr *string, adds ...string) *ModelSetting {
	*columns = append(*columns, adds...)
	*columnStr = "," + strings.ToLower(strings.Join(*columns, ",")) + ","
	return this_
}
func (this_ *ModelSetting) columnsFind(columnStr string, find string) bool {
	return strings.Contains(columnStr, ","+strings.ToLower(find)+",")
}

func (this_ *ModelSetting) EmptySelectInclude() *ModelSetting {
	return this_.columnsEmpty(&this_.selectIncludeColumns, &this_.selectIncludeColumnStr)
}
func (this_ *ModelSetting) SelectInclude(columns ...string) *ModelSetting {
	return this_.columnsAdd(&this_.selectIncludeColumns, &this_.selectIncludeColumnStr, columns...)
}
func (this_ *ModelSetting) IsSelectInclude(column string) bool {
	return this_.columnsFind(this_.selectIncludeColumnStr, column)
}
func (this_ *ModelSetting) EmptySelectExclude() *ModelSetting {
	return this_.columnsEmpty(&this_.selectExcludeColumns, &this_.selectExcludeColumnStr)
}
func (this_ *ModelSetting) SelectExclude(columns ...string) *ModelSetting {
	return this_.columnsAdd(&this_.selectExcludeColumns, &this_.selectExcludeColumnStr, columns...)
}
func (this_ *ModelSetting) IsSelectExclude(column string) bool {
	return this_.columnsFind(this_.selectExcludeColumnStr, column)
}

func (this_ *ModelSetting) EmptyValueInclude() *ModelSetting {
	return this_.columnsEmpty(&this_.valueIncludeColumns, &this_.valueIncludeColumnStr)
}
func (this_ *ModelSetting) ValueInclude(columns ...string) *ModelSetting {
	return this_.columnsAdd(&this_.valueIncludeColumns, &this_.valueIncludeColumnStr, columns...)
}
func (this_ *ModelSetting) IsValueInclude(column string) bool {
	return this_.columnsFind(this_.valueIncludeColumnStr, column)
}
func (this_ *ModelSetting) EmptyValueExclude() *ModelSetting {
	return this_.columnsEmpty(&this_.valueExcludeColumns, &this_.valueExcludeColumnStr)
}
func (this_ *ModelSetting) ValueExclude(columns ...string) *ModelSetting {
	return this_.columnsAdd(&this_.valueExcludeColumns, &this_.valueExcludeColumnStr, columns...)
}
func (this_ *ModelSetting) IsValueExclude(column string) bool {
	return this_.columnsFind(this_.valueExcludeColumnStr, column)
}

func (this_ *ModelSetting) EmptyWhereInclude() *ModelSetting {
	return this_.columnsEmpty(&this_.whereIncludeColumns, &this_.whereIncludeColumnStr)
}
func (this_ *ModelSetting) WhereInclude(columns ...string) *ModelSetting {
	return this_.columnsAdd(&this_.whereIncludeColumns, &this_.whereIncludeColumnStr, columns...)
}
func (this_ *ModelSetting) IsWhereInclude(column string) bool {
	return this_.columnsFind(this_.whereIncludeColumnStr, column)
}
func (this_ *ModelSetting) EmptyWhereExclude() *ModelSetting {
	return this_.columnsEmpty(&this_.whereExcludeColumns, &this_.whereExcludeColumnStr)
}
func (this_ *ModelSetting) WhereExclude(columns ...string) *ModelSetting {
	return this_.columnsAdd(&this_.whereExcludeColumns, &this_.whereExcludeColumnStr, columns...)
}
func (this_ *ModelSetting) IsWhereExclude(column string) bool {
	return this_.columnsFind(this_.whereExcludeColumnStr, column)
}

type IncludedPlace int

var IncludedPlaceValue = IncludedPlace(1)
var IncludedPlaceWhere = IncludedPlace(2)

func (this_ *ModelSetting) Included(place IncludedPlace, columnName string, columnValue *util.FieldValue) bool {
	if columnName != "" {
		if place == IncludedPlaceValue {
			if len(this_.valueExcludeColumns) > 0 {
				// 如果是 排除字段 直接返回 不包含
				if this_.IsValueExclude(columnName) {
					return false
				}
			}
			if len(this_.valueIncludeColumns) > 0 {
				// 如果是 包含字段 直接返回 包含
				if this_.IsValueInclude(columnName) {
					return true
				}
				// 如果 配置了包含 则 不在里边的字段 直接 忽略
				return false
			}
		} else if place == IncludedPlaceWhere {
			if len(this_.whereExcludeColumns) > 0 {
				// 如果是 排除字段 直接返回 不包含
				if this_.IsWhereExclude(columnName) {
					return false
				}
			}
			if len(this_.whereIncludeColumns) > 0 {
				// 如果是 包含字段 直接返回 包含
				if this_.IsWhereInclude(columnName) {
					return true
				}
				// 如果 配置了包含 则 不在里边的字段 直接 忽略
				return false
			}
		}
	}

	// 忽略 null 值
	if columnValue.IsNull() && !this_.includeNull {
		return false
	}
	// 忽略 0 值
	if columnValue.IsZero() && !this_.includeZero {
		return false
	}
	// 忽略 "" 值
	if columnValue.IsEmpty() && !this_.includeEmpty {
		return false
	}

	return true
}

func (this_ *ModelSetting) GetValue(columnValue *util.FieldValue) (isNull bool, res any) {
	if this_.zeroUseNull && columnValue.IsZero() {
		isNull = true
		return
	}
	if this_.emptyUseNull && columnValue.IsEmpty() {
		isNull = true
		return
	}
	isNull = columnValue.IsNull()
	res = columnValue.GetData()
	return
}
func (this_ *ModelSetting) SetSqlHandler(sqlHandler SqlHandler) *ModelSetting {
	this_.sqlHandler = sqlHandler
	return this_
}
func (this_ *ModelSetting) SetModelOption(modelOption *ModelOption) *ModelSetting {
	this_.modelOption = modelOption
	return this_
}
func (this_ *ModelSetting) SetService(service *Service) *ModelSetting {
	this_.service = service
	return this_
}
func (this_ *ModelSetting) GetService() IService {
	return this_.service
}

func (this_ *ModelSetting) NewBuilder(model any) (res *OrmSqlBuilder) {
	res = new(OrmSqlBuilder)
	res.primaryKey = this_.GetPrimaryKey(model)
	res.primaryKeyStr = "," + strings.ToLower(strings.Join(res.primaryKey, ",")) + ","
	res.ModelOption = this_.modelOption
	res.SqlHandler = this_.sqlHandler
	if this_.service != nil {
		if res.ModelOption == nil {
			res.ModelOption = this_.service.GetModelOption()
		}
		if res.SqlHandler == nil {
			res.SqlHandler = this_.service
		}
	}
	if model != nil {
		res.model = model
		res.modelValue = reflect.ValueOf(model)
	}
	if res.ModelOption == nil {
		res.ModelOption = DefaultModelOption
	}
	if res.SqlHandler == nil {
		res.SqlHandler = DefaultSqlOption
	}

	res.sqlParam = SqlParamByValue(res.model, res.modelValue)
	res.tableName = this_.GetTableName(model)
	if res.tableName != "" {
		res.realTableName = res.tableName
		res.wrapTableName = res.realTableName
		res.realTableName = res.SqlHandler.RealTableName(res.sqlParam, res.tableName)
		res.wrapTableName = res.SqlHandler.WrapTableName(res.sqlParam, res.tableName)
	}
	res.whereOperators = make(map[string]*WhereOperator)
	for _, one := range this_.whereOperators {
		res.whereOperators[strings.ToLower(one.Column)] = one
	}
	return
}

type WhereOperator struct {
	Column   string
	Operator string
}

func (this_ *ModelSetting) WhereOperator(column string, op string) *ModelSetting {
	w := &WhereOperator{}
	w.Column = column
	w.Operator = op
	this_.whereOperators = append(this_.whereOperators, w)
	return this_
}

type OrmSqlBuilder struct {
	model      any
	modelValue reflect.Value

	*ModelOption
	SqlHandler

	sqlParam *SqlParam

	tableName     string
	realTableName string
	wrapTableName string

	primaryKey    []string
	primaryKeyStr string

	whereOperators map[string]*WhereOperator
}

func (this_ *OrmSqlBuilder) IsPrimaryKey(column string) bool {
	return strings.Contains(this_.primaryKeyStr, ","+strings.ToLower(column)+",")
}
