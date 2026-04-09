package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/team-ide/framework"
	"io"
	"net"
	"reflect"
)

type Config struct {
	// Disabled 禁用 上层 初始化服务时候 可以判断该属性 如果为 配置 true 则不去初始化服务
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	DialectType string `json:"dialectType,omitempty" yaml:"dialectType,omitempty"`

	Host     string `json:"host,omitempty" yaml:"host,omitempty"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Address  string `json:"address,omitempty" yaml:"address,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	Database string `json:"database,omitempty" yaml:"database,omitempty"`
	Schema   string `json:"schema,omitempty" yaml:"schema,omitempty"`

	DatabasePath string `json:"databasePath,omitempty" yaml:"databasePath,omitempty"`

	ServerName string `json:"serverName,omitempty" yaml:"serverName,omitempty"`

	Dsn       string `json:"dsn,omitempty" yaml:"dsn,omitempty"`
	DsnAppend string `json:"dsnAppend,omitempty" yaml:"dsnAppend,omitempty"`

	MaxIdleConn int `json:"maxIdleConn,omitempty" yaml:"maxIdleConn,omitempty"`
	MaxOpenConn int `json:"maxOpenConn,omitempty" yaml:"maxOpenConn,omitempty"`

	TlsConfig     string `json:"tlsConfig,omitempty" yaml:"tlsConfig,omitempty"`
	TlsRootCert   string `json:"tlsRootCert,omitempty" yaml:"tlsRootCert,omitempty"`
	TlsClientCert string `json:"tlsClientCert,omitempty" yaml:"tlsClientCert,omitempty"`
	TlsClientKey  string `json:"tlsClientKey,omitempty" yaml:"tlsClientKey,omitempty"`

	getDatabasePath func(cfg *Config) string

	connProxy ConnProxy
}

type ConnProxy interface {
	Dial(n string, addr string) (net.Conn, error)
}

func (this_ *Config) SetConnProxy(connProxy ConnProxy) {
	this_.connProxy = connProxy
}
func (this_ *Config) GetConnProxy() ConnProxy {
	return this_.connProxy
}

func (this_ *Config) SetGetDatabasePath(getDatabasePath func(cfg *Config) string) {
	this_.getDatabasePath = getDatabasePath
}

func (this_ *Config) Clone() (res *Config) {
	res = &Config{}
	bs, _ := json.Marshal(this_)
	_ = json.Unmarshal(bs, res)
	res.connProxy = this_.connProxy
	res.getDatabasePath = this_.getDatabasePath
	return res
}

func NewSqlDb(cfg *Config) (sqlDb *sql.DB, err error) {
	dbType := GetDialect(cfg.Type)
	if dbType == nil {
		err = errors.New("db type [" + cfg.Type + "] not support")
		return
	}
	driverDsn := dbType.GetDriverDSN(cfg, nil)
	driverName := dbType.GetDriverName()
	sqlDb, err = sql.Open(driverName, driverDsn)
	if err != nil {
		return
	}
	err = sqlDb.Ping()
	if err != nil {
		return
	}
	return
}
func New(cfg *Config, sqlConn SqlConn) (res *Service, err error) {
	var dialectType = cfg.DialectType
	if dialectType == "" {
		dialectType = cfg.Type
	}
	dia := GetDialect(dialectType)
	if dia == nil {
		err = errors.New("dialect type [" + dialectType + "] not found dialect")
		return
	}
	if sqlConn == nil {
		sqlConn, err = NewSqlDb(cfg)
		if err != nil {
			return
		}
	}
	ser := &Service{}
	ser.cfg = cfg
	ser.sqlConn = sqlConn
	ser.Dialect = dia

	return ser, nil
}

type Service struct {
	cfg         *Config
	sqlConn     SqlConn
	ddlHandler  DDLHandler
	sqlHandler  SqlHandler
	modelOption *ModelOption
	Dialect
	dynamicService IDynamicService

	openShowQuerySql bool
	openShowExecSql  bool

	getSqlValue_   GetSqlValueType
	setFieldValue_ SetFieldValueType
}

func (this_ *Service) OpenShowQuerySql() *Service {
	this_.openShowQuerySql = true
	return this_
}
func (this_ *Service) CloseShowQuerySql() *Service {
	this_.openShowQuerySql = false
	return this_
}
func (this_ *Service) ShowQuerySql() bool {
	return this_.openShowQuerySql
}
func (this_ *Service) OpenShowExecSql() *Service {
	this_.openShowExecSql = true
	return this_
}
func (this_ *Service) CloseShowExecSql() *Service {
	this_.openShowExecSql = false
	return this_
}
func (this_ *Service) ShowExecSql() bool {
	return this_.openShowExecSql
}

type GetSqlValueType func(columnType *sql.ColumnType, data any) (res any)
type SetFieldValueType func(columnType *sql.ColumnType, field reflect.StructField, fieldValue reflect.Value, value any) (err error)

func (this_ *Service) SetGetSqlValue(getSqlValue GetSqlValueType) {
	this_.getSqlValue_ = getSqlValue
}
func (this_ *Service) GetSqlValue() GetSqlValueType {
	return this_.getSqlValue_
}

func (this_ *Service) SetSetFieldValue(setFieldValue SetFieldValueType) {
	this_.setFieldValue_ = setFieldValue
}
func (this_ *Service) SetFieldValue() SetFieldValueType {
	return this_.setFieldValue_
}

func (this_ *Service) SetDynamicService(dynamicService IDynamicService) {
	this_.dynamicService = dynamicService
}
func (this_ *Service) SetModelOption(modelOption *ModelOption) {

	this_.modelOption = modelOption
}
func (this_ *Service) GetModelOption() (modelOption *ModelOption) {
	if this_.modelOption == nil {
		return DefaultModelOption
	}
	return this_.modelOption
}
func (this_ *Service) SetDDLHandler(handler DDLHandler) {
	this_.ddlHandler = handler
}
func (this_ *Service) GetDDLHandler() (handler DDLHandler) {
	return this_.ddlHandler
}
func (this_ *Service) SetSqlHandler(handler SqlHandler) {
	this_.sqlHandler = handler
}
func (this_ *Service) GetSqlHandler() (handler SqlHandler) {
	return this_.sqlHandler
}

func (this_ *Service) Close() {
	if this_ == nil {
		return
	}
	var c = this_.sqlConn
	this_.sqlConn = nil
	if c == nil {
		return
	}
	closer, ok := c.(io.Closer)
	if ok {
		_ = closer.Close()
	}
}
func (this_ *Service) GetConfig() Config {
	return *this_.cfg
}

func (this_ *Service) GetDialect() Dialect {
	return this_.Dialect
}
func (this_ *Service) GetSqlConn() SqlConn {
	return this_.sqlConn
}

func (this_ *Service) GetDatabaseName() string {
	return this_.cfg.Database
}
func (this_ *Service) GetSchemaName() string {
	return this_.cfg.Schema
}
func (this_ *Service) RealService(param *SqlParam, tableName string) IService {
	ds := this_.dynamicService
	if ds != nil && ds != this_ {
		return ds.RealService(param, tableName)
	}
	return this_
}
func (this_ *Service) RealTableName(param *SqlParam, tableName string) (realTableName string) {
	sqlHandler := this_.sqlHandler
	if sqlHandler != nil && sqlHandler != this_ {
		return sqlHandler.RealTableName(param, tableName)
	}
	return tableName
}
func (this_ *Service) WrapTableName(param *SqlParam, tableName string) (wrapTableName string) {
	sqlHandler := this_.sqlHandler
	if sqlHandler != nil && sqlHandler != this_ {
		return sqlHandler.WrapTableName(param, tableName)
	}
	return tableName
}
func (this_ *Service) WrapColumnName(param *SqlParam, columnName string) (wrapColumnName string) {
	sqlHandler := this_.sqlHandler
	if sqlHandler != nil && sqlHandler != this_ {
		return sqlHandler.WrapColumnName(param, columnName)
	}
	return columnName
}

func (this_ *Service) CheckTableExist(tableName string) (res bool) {
	return this_.GetDialect().TableCheckExists(this_.GetSqlConn(), this_.GetDatabaseName(), this_.GetSchemaName(), tableName)
}

func (this_ *Service) CreateTable(table *Table) (err error) {
	return this_.GetDialect().TableCreate(this_.GetSqlConn(), this_.GetDDLHandler(), this_.GetDatabaseName(), this_.GetSchemaName(), table)
}

func (this_ *Service) CheckColumnExist(tableName string, columnName string) (res bool) {
	return this_.GetDialect().ColumnCheckExists(this_.GetSqlConn(), this_.GetDatabaseName(), this_.GetSchemaName(), tableName, columnName)
}

func (this_ *Service) AddColumn(tableName string, column *Column) (err error) {
	return this_.GetDialect().ColumnAdd(this_.GetSqlConn(), this_.GetDDLHandler(), this_.GetDatabaseName(), this_.GetSchemaName(), tableName, column)
}

func (this_ *Service) CheckIndexExist(tableName string, checkIndex *Index) (res bool) {
	return this_.GetDialect().IndexCheckExists(this_.GetSqlConn(), this_.GetDatabaseName(), this_.GetSchemaName(), tableName, checkIndex)
}

func (this_ *Service) CreateIndex(tableName string, index *Index) (err error) {
	return this_.GetDialect().IndexAdd(this_.GetSqlConn(), this_.GetDDLHandler(), this_.GetDatabaseName(), this_.GetSchemaName(), tableName, index)
}

func (this_ *Service) Exec(ctx context.Context, sqlInfo string, args []any) (res sql.Result, err error) {

	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	res, err = DoExec(ctx, this_.GetSqlConn(), sqlInfo, args, this_.ShowExecSql())
	if err != nil {
		return
	}
	return
}

func (this_ *Service) Execs(ctx context.Context, sqlList []string, argsList [][]any) (results []sql.Result, err error) {
	for i, sqlInfo := range sqlList {
		sqlList[i] = this_.FormatSqlArgChar(sqlInfo)
	}
	results, _, _, err = DoExecs(ctx, this_.GetSqlConn(), sqlList, argsList, this_.ShowExecSql())
	if err != nil {
		return
	}
	return
}

func (this_ *Service) QueryCount(ctx context.Context, sqlInfo string, args []any) (count int64, err error) {
	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	count, err = DoQueryCount(ctx, this_.GetSqlConn(), sqlInfo, args, this_.ShowQuerySql())
	if err != nil {
		return
	}
	return
}

func (this_ *Service) QueryMapOne(ctx context.Context, sqlInfo string, args []any) (one map[string]any, err error) {
	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	one, err = DoQueryOne(ctx, this_.GetSqlConn(), sqlInfo, args, this_.ShowQuerySql(), this_.GetSqlValue())
	if err != nil {
		return
	}
	return
}

func (this_ *Service) QueryMapList(ctx context.Context, sqlInfo string, args []any) (list []map[string]any, err error) {
	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	list, err = DoQuery(ctx, this_.GetSqlConn(), sqlInfo, args, this_.ShowQuerySql(), this_.GetSqlValue())
	if err != nil {
		return
	}
	return
}

func (this_ *Service) QueryMapPage(ctx context.Context, sqlInfo string, args []any, pageSize int64, pageNo int64) (list []map[string]any, err error) {
	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	sqlInfo = this_.FormatPageSql(sqlInfo, pageSize, pageNo)
	list, err = DoQuery(ctx, this_.GetSqlConn(), sqlInfo, args, this_.ShowQuerySql(), this_.GetSqlValue())
	if err != nil {
		return
	}
	return
}

func (this_ *Service) ModelSelect(model any) (res *ModelSelect) {
	res = &ModelSelect{}
	res.ModelSetting = &ModelSetting{}
	res.model = model
	res.service = this_
	return
}

func (this_ *Service) SqlSelect(table string, columns ...string) (res *SqlSelect) {
	res = &SqlSelect{}
	res.ModelSetting = &ModelSetting{}
	res.SetTableName(table)
	res.Select(columns...)
	res.service = this_
	return
}

type IModelSql interface {
	GetSql() (sqlInfo string, args []any, err error)
	GetService() (service IService)
}

func DoQueryOneWithModel[S any](ctx context.Context, s IModelSql) (res S, err error) {
	list, err := DoQueryListWithModel[S](ctx, s)
	if err != nil {
		return
	}
	size := len(list)
	if size > 1 {
		err = ErrorHasMoreRows
		return
	} else if size == 1 {
		res = list[0]
	}
	return
}
func DoQueryListWithModel[S any](ctx context.Context, s IModelSql) (res []S, err error) {
	sqlInfo, args, err := s.GetSql()
	if err != nil {
		framework.Error("sql select get sql error:" + err.Error())
		return
	}
	service := s.GetService()
	if service == nil {
		err = errors.New("sql select service is null")
		framework.Error(err.Error())
		return
	}
	res, err = DoQueryListWithSql[S](ctx, service, sqlInfo, args)
	return
}
func DoQueryPageWithModel[S any](ctx context.Context, s IModelSql, pageSize int64, pageNo int64) (res []S, err error) {
	sqlInfo, args, err := s.GetSql()
	if err != nil {
		framework.Error("sql select get sql error:" + err.Error())
		return
	}
	service := s.GetService()
	if service == nil {
		err = errors.New("sql select service is null")
		framework.Error(err.Error())
		return
	}
	res, err = DoQueryPageWithSql[S](ctx, service, sqlInfo, args, pageSize, pageNo)
	return
}
func DoQueryOneWithSql[S any](ctx context.Context, service IService, sqlInfo string, sqlArgs []any) (res S, err error) {
	list, err := DoQueryListWithSql[S](ctx, service, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	size := len(list)
	if size > 1 {
		err = ErrorHasMoreRows
		return
	} else if size == 1 {
		res = list[0]
	}
	return
}
func DoQueryListWithSql[S any](ctx context.Context, service IService, sqlInfo string, sqlArgs []any) (res []S, err error) {
	sqlInfo = service.FormatSqlArgChar(sqlInfo)
	res, err = DoQueryListStruct[S](ctx, service.GetSqlConn(), sqlInfo, sqlArgs, service.ShowQuerySql(), service.GetModelOption())
	return
}
func DoQueryPageWithSql[S any](ctx context.Context, service IService, sqlInfo string, sqlArgs []any, pageSize int64, pageNo int64) (res []S, err error) {
	sqlInfo = service.FormatSqlArgChar(sqlInfo)
	sqlInfo = service.FormatPageSql(sqlInfo, pageSize, pageNo)
	res, err = DoQueryListStruct[S](ctx, service.GetSqlConn(), sqlInfo, sqlArgs, service.ShowQuerySql(), service.GetModelOption())
	return
}

func (this_ *Service) Count(ctx context.Context, model IModel) (res int64, err error) {
	m := this_.ModelCount(model)
	res, err = m.Count(ctx)
	return
}
func (this_ *Service) ModelCount(model any) (res *ModelCount) {
	res = &ModelCount{}
	res.ModelSetting = &ModelSetting{}
	res.model = model
	res.service = this_
	return
}
func (this_ *ModelCount) Count(ctx context.Context) (res int64, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("model count get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("model count service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.QueryCount(ctx, sqlInfo, args)
	return
}
func (this_ *Service) SqlCount(table string) (res *SqlCount) {
	res = &SqlCount{}
	res.ModelSetting = &ModelSetting{}
	res.SetTableName(table)
	res.service = this_
	return
}
func (this_ *SqlCount) Count(ctx context.Context) (res int64, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("sql count get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("sql count service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.QueryCount(ctx, sqlInfo, args)
	return
}

func (this_ *Service) Insert(ctx context.Context, model IModel) (res sql.Result, err error) {
	m := this_.ModelInsert(model)
	res, err = m.Exec(ctx)
	return
}
func (this_ *Service) ModelInsert(model any) (res *ModelInsert) {
	res = &ModelInsert{}
	res.ModelSetting = &ModelSetting{}
	res.model = model
	res.service = this_
	return
}
func (this_ *ModelInsert) Exec(ctx context.Context) (res sql.Result, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("model insert get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("model insert service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.Exec(ctx, sqlInfo, args)
	return
}
func (this_ *Service) SqlInsert(table string) (res *SqlInsert) {
	res = &SqlInsert{}
	res.ModelSetting = &ModelSetting{}
	res.SetTableName(table)
	res.service = this_
	return
}
func (this_ *SqlInsert) Exec(ctx context.Context) (res sql.Result, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("sql insert get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("sql insert service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.Exec(ctx, sqlInfo, args)
	return
}

func (this_ *Service) Update(ctx context.Context, model IModel) (res sql.Result, err error) {
	keys := model.GetPrimaryKey()
	m := this_.ModelUpdate(model)
	m.ExcludeColumn(keys...)
	where := m.Where()
	b := m.NewBuilder(model)
	for _, key := range keys {
		fieldValue := b.GetColumnValue(b.modelValue, key)
		isNull, value := m.GetValue(fieldValue)
		if isNull {
			where.IsNull(key)
		} else {
			where.Eq(key, value)
		}
	}
	res, err = m.Exec(ctx)
	return
}
func (this_ *Service) ModelUpdate(model any) (res *ModelUpdate) {
	res = &ModelUpdate{}
	res.ModelSetting = &ModelSetting{}
	res.model = model
	res.service = this_
	return
}
func (this_ *ModelUpdate) Exec(ctx context.Context) (res sql.Result, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("model update get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("model update service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.Exec(ctx, sqlInfo, args)
	return
}
func (this_ *Service) SqlUpdate(table string) (res *SqlUpdate) {
	res = &SqlUpdate{}
	res.ModelSetting = &ModelSetting{}
	res.SetTableName(table)
	res.service = this_
	return
}
func (this_ *SqlUpdate) Exec(ctx context.Context) (res sql.Result, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("sql update get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("sql update service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.Exec(ctx, sqlInfo, args)
	return
}

func (this_ *Service) Delete(ctx context.Context, model IModel) (res sql.Result, err error) {
	m := this_.ModelDelete(model)
	res, err = m.Exec(ctx)
	return
}

func (this_ *Service) ModelDelete(model any) (res *ModelDelete) {
	res = &ModelDelete{}
	res.ModelSetting = &ModelSetting{}
	res.model = model
	res.service = this_
	return
}
func (this_ *ModelDelete) Exec(ctx context.Context) (res sql.Result, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("model delete get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("model delete service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.Exec(ctx, sqlInfo, args)
	return
}

func (this_ *Service) SqlDelete(table string) (res *SqlDelete) {
	res = &SqlDelete{}
	res.ModelSetting = &ModelSetting{}
	res.SetTableName(table)
	res.service = this_
	return
}
func (this_ *SqlDelete) Exec(ctx context.Context) (res sql.Result, err error) {
	sqlInfo, args, err := this_.GetSql()
	if err != nil {
		framework.Error("sql delete get sql error:" + err.Error())
		return
	}
	if this_.service == nil {
		err = errors.New("sql delete service is null")
		framework.Error(err.Error())
		return
	}
	sqlInfo = this_.service.FormatSqlArgChar(sqlInfo)
	res, err = this_.service.Exec(ctx, sqlInfo, args)
	return
}
