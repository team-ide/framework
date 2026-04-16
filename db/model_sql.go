package db

import (
	"errors"
	"github.com/team-ide/framework/util"
	"strings"
)

type SqlSelect struct {
	columns []string
	// 查询条件
	where *Conditions

	// 查询时候 必须 设置条件 除非 设置 SelectAll
	canSelectAll bool
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

// CanSelectAll 查询时候 必须 设置条件 除非 设置 CanSelectAll
func (this_ *SqlSelect) CanSelectAll() *SqlSelect {
	this_.canSelectAll = true
	return this_
}
func (this_ *SqlSelect) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *SqlSelect) SetWhere(where *Conditions) *SqlSelect {
	this_.where = where
	return this_
}
func (this_ *SqlSelect) Select(columns ...string) *SqlSelect {
	this_.columns = append(this_.columns, columns...)
	return this_
}

func (this_ *SqlSelect) GetSql() (sqlInfo string, args []any, err error) {

	b := this_.NewBuilder(nil)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}
	var columns []string
	var includeColumns = this_.selectIncludeColumns
	// 如果 有 包含的字段 则只查询 包含的字段
	if len(includeColumns) > 0 {
		if b.model != nil {
			for _, column := range this_.columns {
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
				for _, column := range this_.columns {
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

	whereSql, whereArgs := this_.GetModelAndWhereSql(b, nil, this_.where)
	if len(whereSql) == 0 {
		if !this_.canSelectAll {
			err = errors.New("select sql 必须设置条件 或者 调下 SelectAll()")
			return
		}
		return
	} else {
		sqlInfo += " WHERE " + whereSql
		args = append(args, whereArgs...)
	}
	appendSql, appendArgs := this_.GetAppendSql()
	sqlInfo += " " + appendSql
	args = append(args, appendArgs...)
	return
}

type SqlCount struct {
	// 查询条件
	where *Conditions

	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

func (this_ *SqlCount) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *SqlCount) SetWhere(where *Conditions) *SqlCount {
	this_.where = where
	return this_
}

func (this_ *SqlCount) GetSql() (sqlInfo string, args []any, err error) {

	b := this_.NewBuilder(nil)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "SELECT COUNT(1) FROM " + b.wrapTableName + ""

	whereSql, whereArgs := this_.GetModelAndWhereSql(b, nil, this_.where)
	if whereSql != "" {
		sqlInfo += " WHERE " + whereSql
		args = append(args, whereArgs...)
	}
	appendSql, appendArgs := this_.GetAppendSql()
	sqlInfo += " " + appendSql
	args = append(args, appendArgs...)
	return
}

type InsertOrUpdateValue struct {
	Column string `json:"column,omitempty"`
	Value  any    `json:"value,omitempty"`
}

type SqlInsert struct {
	values []*InsertOrUpdateValue
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

func (this_ *SqlInsert) Value(column string, value any) *SqlInsert {
	v := &InsertOrUpdateValue{
		Column: column,
		Value:  value,
	}
	this_.values = append(this_.values, v)
	return this_
}

func (this_ *SqlInsert) GetSql() (sqlInfo string, args []any, err error) {
	var vs = this_.values
	if len(vs) == 0 {
		err = errors.New("insert values is null")
		return
	}
	b := this_.NewBuilder(nil)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "INSERT INTO " + b.wrapTableName + " "

	columns, values := GetInsertOrUpdateColumnValues(vs)
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

type SqlUpdate struct {
	values []*InsertOrUpdateValue

	// 更新条件
	where *Conditions

	// 更新时候 必须 设置条件 除非 设置 UpdateAll
	canUpdateAll bool
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

// CanUpdateAll 更新时候 必须 设置条件 除非 设置 CanUpdateAll
func (this_ *SqlUpdate) CanUpdateAll() *SqlUpdate {
	this_.canUpdateAll = true
	return this_
}
func (this_ *SqlUpdate) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *SqlUpdate) Value(column string, value any) *SqlUpdate {
	v := &InsertOrUpdateValue{
		Column: column,
		Value:  value,
	}
	this_.values = append(this_.values, v)
	return this_
}
func (this_ *SqlUpdate) SetWhere(where *Conditions) *SqlUpdate {
	this_.where = where
	return this_
}

func GetInsertOrUpdateColumnValues(list []*InsertOrUpdateValue) (columns []string, values []*util.FieldValue) {
	for _, one := range list {
		columns = append(columns, one.Column)
		fieldValue := util.FieldValueByData(one.Value)
		values = append(values, fieldValue)
	}
	return
}
func (this_ *SqlUpdate) GetSql() (sqlInfo string, args []any, err error) {
	var vs = this_.values
	if len(vs) == 0 {
		err = errors.New("update set values is null")
		return
	}
	b := this_.NewBuilder(nil)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "UPDATE " + b.wrapTableName + " SET"

	columns, values := GetInsertOrUpdateColumnValues(vs)
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

type SqlDelete struct {
	// 删除条件 如果和 model 同时设置 则会拼接条件
	where *Conditions

	// 删除时候 必须 设置条件 除非 设置 DeleteAll
	canDeleteAll bool
	// 模型 设置 如：表名、主键、包含字段、忽略字段、空值设置 等
	*ModelSetting
}

// CanDeleteAll 删除时候 必须 设置条件 除非 设置 CanDeleteAll
func (this_ *SqlDelete) CanDeleteAll() *SqlDelete {
	this_.canDeleteAll = true
	return this_
}
func (this_ *SqlDelete) Where() *Conditions {
	if this_.where == nil {
		this_.where = NewConditions()
	}
	return this_.where
}
func (this_ *SqlDelete) SetWhere(where *Conditions) *SqlDelete {
	this_.where = where
	return this_
}

func (this_ *SqlDelete) GetSql() (sqlInfo string, args []any, err error) {

	b := this_.NewBuilder(nil)
	if b.wrapTableName == "" {
		err = errors.New("table name is empty")
		return
	}

	sqlInfo += "DELETE FROM " + b.wrapTableName + ""

	whereSql, whereArgs := this_.GetModelAndWhereSql(b, nil, this_.where)
	if len(whereSql) == 0 {
		if !this_.canDeleteAll {
			err = errors.New("delete sql 必须设置条件 或者 调下 DeleteAll()")
			return
		}
		return
	} else {
		sqlInfo += " WHERE " + whereSql
		args = append(args, whereArgs...)
	}
	appendSql, appendArgs := this_.GetAppendSql()
	sqlInfo += " " + appendSql
	args = append(args, appendArgs...)
	return
}
