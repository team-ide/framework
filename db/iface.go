package db

import (
	"context"
	"database/sql"
)

type IService interface {
	Dialect
	// Close 关闭 db 客户端
	Close()
	// GetConfig 获取 数据库 配置
	GetConfig() Config
	// GetDialect 获取 方言
	GetDialect() Dialect
	// GetSqlConn 获取 sql.DB
	GetSqlConn() SqlConn

	SetDynamicService(dynamicService IDynamicService)

	GetDDLHandler() (handler DDLHandler)
	GetModelOption() (modelOption *ModelOption)

	ShowQuerySql() bool

	ShowExecSql() bool

	SqlHandler
	// IDynamicService 动态数据源服务相关
	IDynamicService

	// Exec 执行 SQL
	Exec(ctx context.Context, sql string, args []any) (result sql.Result, err error)
	// Execs 批量执行 SQL
	Execs(ctx context.Context, sqlList []string, argsList [][]any) (results []sql.Result, err error)
	// QueryCount 统计查询 SQL
	QueryCount(ctx context.Context, sql string, args []any) (count int64, err error)
	// QueryMapOne 查询 单个 map SQL  返回 < map >
	QueryMapOne(ctx context.Context, sql string, args []any) (one map[string]any, err error)
	// QueryMapList 查询 列表 map SQL  返回 list < map >
	QueryMapList(ctx context.Context, sql string, args []any) (list []map[string]any, err error)
	// QueryMapPage 分页查询 列表 map 列表 SQL  返回 list < map >
	QueryMapPage(ctx context.Context, sql string, args []any, pageSize int64, pageNo int64) (list []map[string]any, err error)

	ModelSelect(model any, sets ...ModelSelectSet) (res *ModelSelect)
	SqlSelect(table string, columns ...string) (res *SqlSelect)

	Count(ctx context.Context, model IModel, sets ...ModelCountSet) (res int64, err error)
	ModelCount(model any, sets ...ModelCountSet) (res *ModelCount)
	SqlCount(table string, sets ...SqlCountSet) (res *SqlCount)

	Insert(ctx context.Context, model IModel, sets ...ModelInsertSet) (res sql.Result, err error)
	ModelInsert(model any, sets ...ModelInsertSet) (res *ModelInsert)
	SqlInsert(table string, sets ...SqlInsertSet) (res *SqlInsert)

	Update(ctx context.Context, model IModel, sets ...ModelUpdateSet) (res sql.Result, err error)
	ModelUpdate(model any, sets ...ModelUpdateSet) (res *ModelUpdate)
	SqlUpdate(table string, sets ...SqlUpdateSet) (res *SqlUpdate)

	Delete(ctx context.Context, model IModel, sets ...ModelDeleteSet) (res sql.Result, err error)
	ModelDelete(model any, sets ...ModelDeleteSet) (res *ModelDelete)
	SqlDelete(table string, sets ...SqlDeleteSet) (res *SqlDelete)
}

type IDynamicService interface {
	// RealService 真实服务 根据 表名 和 参数 获取正则
	RealService(param *SqlParam, tableName string) IService
}
