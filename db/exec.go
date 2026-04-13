package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/team-ide/framework"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
)

func DoExecBatch(ctx context.Context, sqlConn SqlConn, sqlInfo string, argsList [][]interface{}, showSql bool) (res []sql.Result, err error) {
	if showSql {
		framework.Debug("exec batch sql start", zap.Any("sql", sqlInfo), zap.Any("argsList", argsList))
	}
	var startTime = time.Now()
	st, err := sqlConn.PrepareContext(ctx, sqlInfo)
	if err != nil {
		framework.Error("exec batch sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("argsList", argsList))
		return
	}
	defer func() { _ = st.Close() }()

	var r sql.Result
	for _, args := range argsList {
		r, err = st.Exec(args...)
		if err != nil {
			framework.Error("exec batch sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
			return
		}
		res = append(res, r)
	}
	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("exec batch sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sql", sqlInfo), zap.Any("argsList", argsList))
	}
	return
}
func DoExec(ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool) (result sql.Result, err error) {
	if showSql {
		framework.Debug("exec sql start", zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	var startTime = time.Now()
	result, err = sqlConn.ExecContext(ctx, sqlInfo, args...)
	if err != nil {
		framework.Error("exec sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("exec sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	return
}

func DoExecs(ctx context.Context, sqlConn SqlConn, sqlList []string, argsList [][]interface{}, showSql bool) (resultList []sql.Result, errSql string, errArgs []interface{}, err error) {
	sqlListSize := len(sqlList)
	if sqlListSize == 0 {
		return
	}
	if len(argsList) == 0 {
		argsList = make([][]interface{}, sqlListSize)
	}
	argsListSize := len(argsList)
	if sqlListSize != argsListSize {
		err = errors.New(fmt.Sprintf("sqlList size is [%d] but argsList size is [%d]", sqlListSize, argsListSize))
		framework.Error("exec more sql error:"+err.Error(), zap.Any("sqlList", sqlList), zap.Any("argsList", argsList))
		return
	}
	if showSql {
		framework.Debug("exec more sql start", zap.Any("sqlList", sqlList), zap.Any("argsList", argsList))
	}
	var startTime = time.Now()

	var result sql.Result
	for i := 0; i < sqlListSize; i++ {
		sqlInfo := sqlList[i]
		args := argsList[i]
		if strings.TrimSpace(sqlInfo) == "" {
			continue
		}
		result, err = sqlConn.ExecContext(ctx, sqlInfo, args...)
		if err != nil {
			framework.Error("exec more sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
			errSql = sqlInfo
			errArgs = args
			return
		}
		resultList = append(resultList, result)
	}
	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("exec more sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sqlList", sqlList), zap.Any("argsList", argsList))
	}
	return
}

func DoTxExec(ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool) (result sql.Result, err error) {
	if showSql {
		framework.Debug("exec tx sql start", zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	var startTime = time.Now()
	tx, err := sqlConn.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil && strings.Contains(err.Error(), "Not in transaction") {
				err = nil
			}
		}
	}()
	result, err = tx.ExecContext(ctx, sqlInfo, args...)
	if err != nil {
		framework.Error("exec tx sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("exec tx sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	return
}

func DoTxExecs(ctx context.Context, sqlConn SqlConn, sqlList []string, argsList [][]interface{}, showSql bool) (resultList []sql.Result, errSql string, errArgs []interface{}, err error) {
	sqlListSize := len(sqlList)
	if sqlListSize == 0 {
		return
	}
	if len(argsList) == 0 {
		argsList = make([][]interface{}, sqlListSize)
	}
	argsListSize := len(argsList)
	if sqlListSize != argsListSize {
		err = errors.New(fmt.Sprintf("sqlList size is [%d] but argsList size is [%d]", sqlListSize, argsListSize))
		framework.Error("exec more tx sql error:"+err.Error(), zap.Any("sqlList", sqlList), zap.Any("argsList", argsList))
		return
	}
	if showSql {
		framework.Debug("exec more tx sql start", zap.Any("sqlList", sqlList), zap.Any("argsList", argsList))
	}
	var startTime = time.Now()

	tx, err := sqlConn.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil && strings.Contains(err.Error(), "Not in transaction") {
				err = nil
			}
		}
	}()
	var result sql.Result
	for i := 0; i < sqlListSize; i++ {
		sqlInfo := sqlList[i]
		args := argsList[i]
		if strings.TrimSpace(sqlInfo) == "" {
			continue
		}
		result, err = tx.ExecContext(ctx, sqlInfo, args...)
		if err != nil {
			framework.Error("exec more tx sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
			errSql = sqlInfo
			errArgs = args
			return
		}
		resultList = append(resultList, result)
	}
	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("exec more tx sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sqlList", sqlList), zap.Any("argsList", argsList))
	}

	return
}

func DoQuery(ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool, getSqlValue GetSqlValueType) (list []map[string]interface{}, err error) {
	_, _, list, err = DoQueryWithColumnTypes(ctx, sqlConn, sqlInfo, args, showSql, getSqlValue)
	if err != nil {
		return
	}
	return
}

var (
	ErrorHasMoreRows = errors.New("has more rows by query one")
)

func DoQueryOne(ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool, getSqlValue GetSqlValueType) (data map[string]interface{}, err error) {
	_, _, list, err := DoQueryWithColumnTypes(ctx, sqlConn, sqlInfo, args, showSql, getSqlValue)
	if err != nil {
		return
	}
	if len(list) > 0 {
		data = list[0]
		if len(list) > 1 {
			err = ErrorHasMoreRows
			framework.Error("query one sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
			return
		}
	}
	return
}

func DoQueryWithColumnTypes(ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool, getSqlValue GetSqlValueType) (columns []string, columnTypes []*sql.ColumnType, list []map[string]any, err error) {
	if showSql {
		framework.Debug("query sql start", zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	var startTime = time.Now()

	stmt, err := sqlConn.PrepareContext(ctx, sqlInfo)
	if err != nil {
		framework.Error("query sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.Query(args...)
	if err != nil {
		framework.Error("query sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	defer func() { _ = rows.Close() }()

	columns, err = rows.Columns()
	if err != nil {
		framework.Error("query sql get Columns error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	columnTypes, err = rows.ColumnTypes()
	if err != nil {
		framework.Error("query sql get ColumnTypes error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	var columnCount = len(columns)
	var receivers = make([]any, columnCount)

	for rows.Next() {
		for i := range columnCount {
			var v any
			receivers[i] = &v
		}
		err = rows.Scan(receivers...)
		if err != nil {
			framework.Error("query sql value Scan error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
			return
		}
		item := make(map[string]interface{})
		for index, data := range receivers {
			if getSqlValue != nil {
				item[columns[index]] = getSqlValue(columnTypes[index], data)
			} else {
				item[columns[index]] = GetSqlValue(columnTypes[index], data)
			}
		}
		list = append(list, item)
	}

	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("query sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	return
}

func DoQueryCount(ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool) (count int64, err error) {
	if showSql {
		framework.Debug("query count sql start", zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	var startTime = time.Now()
	stmt, err := sqlConn.PrepareContext(ctx, sqlInfo)
	if err != nil {
		framework.Error("query count sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.Query(args...)
	if err != nil {
		framework.Error("query count sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			framework.Error("query count sql value Scan error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
			return
		}
	}

	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("query count sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	return
}

func DoQueryOneStructWithService[S any](ctx context.Context, service IService, sqlInfo string, args []interface{}) (res S, err error) {
	res, err = DoQueryOneStruct[S](ctx, service.GetSqlConn(), sqlInfo, args, service.ShowQuerySql(), service.GetModelOption())
	return
}
func DoQueryListStructWithService[S any](ctx context.Context, service IService, sqlInfo string, args []interface{}) (res []S, err error) {
	res, err = DoQueryListStruct[S](ctx, service.GetSqlConn(), sqlInfo, args, service.ShowQuerySql(), service.GetModelOption())
	return
}
func DoQueryOneStruct[S any](ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool, modelOption *ModelOption) (res S, err error) {
	list, err := DoQueryListStruct[S](ctx, sqlConn, sqlInfo, args, showSql, modelOption)
	if err != nil {
		return
	}
	if len(list) > 1 {
		err = ErrorHasMoreRows
		return
	}
	res = list[0]
	return
}

func DoQueryListStruct[S any](ctx context.Context, sqlConn SqlConn, sqlInfo string, args []interface{}, showSql bool, modelOption *ModelOption) (list []S, err error) {
	if modelOption == nil {
		modelOption = DefaultModelOption
	}
	var s S
	modelType := reflect.TypeOf(s)

	var elemType reflect.Type
	var isPointer bool

	if modelType.Kind() == reflect.Ptr {
		elemType = modelType.Elem()
		isPointer = true
	} else {
		elemType = modelType
	}

	modelInfo := modelOption.GetModelInfo(elemType)
	if showSql {
		framework.Debug("query sql start", zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	var startTime = time.Now()

	stmt, err := sqlConn.PrepareContext(ctx, sqlInfo)
	if err != nil {
		framework.Error("query sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.Query(args...)
	if err != nil {
		framework.Error("query sql error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		framework.Error("query sql get Columns error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
		return
	}

	// 预构建字段映射和值缓存
	columnCount := len(columns)

	fields := make([]*ModelField, columnCount)

	for i, column := range columns {
		columnLower := strings.ToLower(column)
		field := modelInfo.columnLower[columnLower]
		fields[i] = field
	}
	// 创建基础values切片，包含所有列
	receivers := make([]any, columnCount)

	for rows.Next() {
		for i, field := range fields {
			if field != nil {
				if field.ImplementsSqlScanner {
					// 如果字段实现了Scanner，使用字段类型创建Scanner
					receiver := reflect.New(field.ElemType).Interface()
					receivers[i] = receiver
				} else {
					if field.sqlValueType != nil {
						receivers[i] = reflect.New(field.sqlValueType).Interface()
					} else {
						var val any
						receivers[i] = &val
					}
				}
			} else {
				// 没有匹配字段，使用通用接收器
				var val any
				receivers[i] = &val
			}
		}

		err = rows.Scan(receivers...)
		if err != nil {
			framework.Error("query sql value Scan error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args))
			return
		}

		// 创建新的结构体实例
		onePtr := reflect.New(elemType)
		one := onePtr.Elem()
		// 将扫描到的值赋给结构体字段
		for i, receiver := range receivers {
			field := fields[i]
			if field == nil {
				continue
			}
			var oneField reflect.Value
			if field.ParentFiled == nil {
				oneField = one.Field(field.Index)
			} else {
				if field.ParentFiled.IsPtr {
					parentField := one.Field(field.ParentFiled.Index)
					if parentField.IsNil() {
						parentFieldV := reflect.New(field.ParentFiled.ElemType)
						parentField.Set(parentFieldV)
					}
				}
				oneField = one.FieldByName(field.FieldName)
			}
			var fieldSetter = field.fieldSetter
			if field.ImplementsSqlScanner {
				fieldSetter = fieldSetterScanner
			}
			if fieldSetter == nil {
				err = errors.New("model type [" + elemType.String() + "] field [" + field.FieldName + "] setter is null")
				framework.Error(err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args), zap.Any("field", field))
				return
			} else {
				err = fieldSetter(one, field, oneField, receiver)
				if err != nil {
					framework.Error("model field setter error:"+err.Error(), zap.Any("sql", sqlInfo), zap.Any("args", args), zap.Any("field", field))
					return
				}
			}
		}

		var result S
		if isPointer {
			// 如果S是指针类型，返回结构体的指针
			result = one.Addr().Interface().(S)
		} else {
			// 如果S是非指针类型，返回结构体值
			result = one.Interface().(S)
		}
		list = append(list, result)
	}

	if showSql {
		var endTime = time.Now()
		framework.Debug(fmt.Sprintf("query sql end, use==>%dms", endTime.UnixMilli()-startTime.UnixMilli()), zap.Any("sql", sqlInfo), zap.Any("args", args))
	}
	return
}
