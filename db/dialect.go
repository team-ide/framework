package db

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/team-ide/framework"
	"github.com/team-ide/framework/util"
	"sort"
	"strings"
)

type createDialect func() (res Dialect)

var (
	dialectList []Dialect
)

func GetDialect(name string) Dialect {
	//fmt.Println("GetDialect name:", name)
	for _, one := range dialectList {
		//fmt.Println("find dialect type:", one.Type())
		if strings.EqualFold(one.Type(), name) {
			return one
		}
	}
	for _, one := range dialectList {
		for _, m := range one.Match() {
			//fmt.Println("find dialect match:", m)
			if strings.EqualFold(m, name) {
				return one
			}
		}
	}
	return nil
}

func AddDialect(dia Dialect) Dialect {
	//fmt.Println("add dialect type:", dia.Type(), " match:", dia.Match())
	str := "add dialect type:" + dia.Type() + " match:" + strings.Join(dia.Match(), ",")
	fmt.Println(str)
	framework.Info(str)
	dialectList = append(dialectList, dia)
	return dia
}

func AddByDialectConfigBase64(dialectType string, cfgContent string) (dia Dialect, err error) {
	// 使用StdEncoding解码
	bs, err := base64.StdEncoding.DecodeString(cfgContent)
	if err != nil {
		err = errors.New("decoding base64 error:" + err.Error())
		return
	}
	bs, err = util.UnGzipBytes(bs)
	if err != nil {
		err = errors.New("decoding gzip error:" + err.Error())
		return
	}
	//fmt.Println("AddByDialectConfigBase64:", string(bs))
	cfg, err := ToDialectConfigByBytes(dialectType, bs)
	if err != nil {
		return
	}

	err = cfg.Init()

	if err != nil {
		err = errors.New("dialect config [" + cfg.Type + "] init error:" + err.Error())
		return
	}

	AddDialectConfig(cfg)
	dia = NewDialect(cfg)
	AddDialect(dia)
	return
}

type Info struct {
	DriverName     string
	Dsn            string
	DsnHasDatabase bool
	DsnHasSchema   bool

	// 名称 没有包装 则自动转大写
	NameNoWrapIsUpper bool `json:"nameNoWrapIsUpper"`

	// SQL 切换用户
	SqlChangeUser bool `json:"canChangeUser"`
	// SQL 切换数据库
	SqlChangeDatabase bool `json:"canChangeDatabase"`
	// SQL 切换模式
	SqlChangeSchema bool `json:"canChangeSchema"`

	// 是否有 用户
	HasUser bool `json:"hasUser"`
	// 是否有 库
	HasDatabase bool `json:"hasDatabase"`
	// 是否有 模式
	HasSchema bool `json:"hasSchema"`
	// 是否有 序列
	HasSequence bool `json:"hasSequence"`

	// 创建用户会自动创建 同名 数据库 则 用户登录 将自动登录到 同名 数据库 中
	CreateUserAutoCreateDatabase bool `json:"createUserAutoCreateDatabase"`
	// 创建用户会自动创建 同名 模式 则 用户登录 将自动登录到 同名 模式 中
	CreateUserAutoCreateSchema bool `json:"createUserAutoCreateSchema"`
}

type SqlConn interface {
	PingContext(ctx context.Context) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type Dialect interface {
	Type() string
	Match() []string

	WrapNameChar() string
	WrapStringChar() string
	EscapeStringChar() string
	ArgChar() string

	FormatSqlArgChar(sqlInfo string) (res string)
	FormatPageSql(sqlInfo string, pageSize int64, pageNo int64) (res string)

	GetDriverName() string
	GetDriverDSN(dbCfg *Config, params map[string]string) string

	Open(dbCfg *Config, params map[string]string) (sqlDb *sql.DB, err error)

	Info() (info *Info)

	UserChange(sqlConn SqlConn, userName string, password string) (ok bool, err error)
	UserSelectSql(userName string) (sqlInfo string, sqlArgs []any)
	UserList(sqlConn SqlConn) (userList []*User, err error)
	User(sqlConn SqlConn, userName string) (user *User, err error)
	UserCreateFieldList() (sqlList []*Field)
	UserCreateSql(handler DDLHandler, user *User) (sqlList []string, sqlArgsList [][]any)
	UserCreate(sqlConn SqlConn, handler DDLHandler, user *User) (err error)
	UserDeleteSql(handler DDLHandler, userName string) (sqlList []string, sqlArgsList [][]any)
	UserDelete(sqlConn SqlConn, handler DDLHandler, userName string) (err error)

	DatabaseChange(sqlConn SqlConn, databaseName string) (ok bool, err error)
	DatabaseSelectSql(databaseName string) (sqlInfo string, sqlArgs []any)
	DatabaseList(sqlConn SqlConn) (databaseList []*Database, err error)
	Database(sqlConn SqlConn, databaseName string) (database *Database, err error)
	DatabaseCreateFieldList() (sqlList []*Field)
	DatabaseCreateSql(handler DDLHandler, database *Database) (sqlList []string, sqlArgsList [][]any)
	DatabaseCreate(sqlConn SqlConn, handler DDLHandler, database *Database) (err error)
	DatabaseDeleteSql(handler DDLHandler, databaseName string, cascade bool) (sqlList []string, sqlArgsList [][]any)
	DatabaseDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, cascade bool) (err error)

	SchemaChange(sqlConn SqlConn, schemaName string) (ok bool, err error)
	SchemaSelectSql(databaseName string, schemaName string) (sqlInfo string, sqlArgs []any)
	SchemaList(sqlConn SqlConn, databaseName string) (schemaList []*Schema, err error)
	Schema(sqlConn SqlConn, databaseName string, schemaName string) (schema *Schema, err error)
	SchemaCreateFieldList() (sqlList []*Field)
	SchemaCreateSql(handler DDLHandler, databaseName string, schema *Schema) (sqlList []string, sqlArgsList [][]any)
	SchemaCreate(sqlConn SqlConn, handler DDLHandler, databaseName string, schema *Schema) (err error)
	SchemaDeleteSql(handler DDLHandler, databaseName string, schemaName string, cascade bool) (sqlList []string, sqlArgsList [][]any)
	SchemaDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, cascade bool) (err error)

	SequenceSelectSql(databaseName string, schemaName string, sequenceName string) (sqlInfo string, sqlArgs []any)
	SequenceList(sqlConn SqlConn, databaseName string, schemaName string) (sequenceList []*Sequence, err error)
	Sequence(sqlConn SqlConn, databaseName string, schemaName string, sequenceName string) (sequence *Sequence, err error)
	SequenceCreateFieldList() (sqlList []*Field)
	SequenceCreateSql(handler DDLHandler, databaseName string, schemaName string, sequence *Sequence) (sqlList []string, sqlArgsList [][]any)
	SequenceCreate(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, sequence *Sequence) (err error)
	SequenceDeleteSql(handler DDLHandler, databaseName string, schemaName string, sequenceName string) (sqlList []string, sqlArgsList [][]any)
	SequenceDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, sequenceName string) (err error)

	TableCheckExists(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (exists bool)
	TableSelectSql(databaseName string, schemaName string, tableName string) (sqlInfo string, sqlArgs []any)
	TableList(sqlConn SqlConn, databaseName string, schemaName string) (tableList []*Table, err error)
	Table(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (table *Table, err error)
	TableDetail(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (table *Table, err error)
	TableCreateFieldList() (sqlList []*Field)
	TableCreateSql(handler DDLHandler, databaseName string, schemaName string, table *Table) (sqlList []string, sqlArgsList [][]any)
	TableCreate(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, table *Table) (err error)
	TableDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string) (sqlList []string, sqlArgsList [][]any)
	TableDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string) (err error)

	ColumnCheckExists(sqlConn SqlConn, databaseName string, schemaName string, tableName string, columnName string) (exists bool)
	ColumnSelectSql(databaseName string, schemaName string, tableName string, columnName string) (sqlInfo string, sqlArgs []any)
	ColumnList(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (columnList []*Column, err error)
	Column(sqlConn SqlConn, databaseName string, schemaName string, tableName string, columnName string) (column *Column, err error)
	ColumnAddSql(handler DDLHandler, databaseName string, schemaName string, tableName string, column *Column) (sqlList []string, sqlArgsList [][]any)
	ColumnAdd(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, column *Column) (err error)
	ColumnUpdateSql(handler DDLHandler, databaseName string, schemaName string, tableName string, oldColumn *Column, column *Column) (sqlList []string, sqlArgsList [][]any)
	ColumnUpdate(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, oldColumn *Column, column *Column) (err error)
	ColumnDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string, columnName string) (sqlList []string, sqlArgsList [][]any)
	ColumnDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, columnName string) (err error)

	ConstraintSelectSql(databaseName string, schemaName string, tableName string, constraintName string) (sqlInfo string, sqlArgs []any)
	ConstraintList(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (constraintList []*Constraint, err error)
	Constraint(sqlConn SqlConn, databaseName string, schemaName string, tableName string, constraintName string) (constraint *Constraint, err error)
	ConstraintAddSql(handler DDLHandler, databaseName string, schemaName string, tableName string, constraint *Constraint) (sqlList []string, sqlArgsList [][]any)
	ConstraintAdd(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, constraint *Constraint) (err error)
	ConstraintDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string, constraintName string) (sqlList []string, sqlArgsList [][]any)
	ConstraintDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, constraintName string) (err error)

	IndexCheckExists(sqlConn SqlConn, databaseName string, schemaName string, tableName string, checkIndex *Index) (exists bool)
	IndexSelectSql(databaseName string, schemaName string, tableName string, indexName string) (sqlInfo string, sqlArgs []any)
	IndexList(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (indexList []*Index, err error)
	Index(sqlConn SqlConn, databaseName string, schemaName string, tableName string, indexName string) (index *Index, err error)
	IndexAddSql(handler DDLHandler, databaseName string, schemaName string, tableName string, index *Index) (sqlList []string, sqlArgsList [][]any)
	IndexAdd(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, index *Index) (err error)
	IndexDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string, indexName string) (sqlList []string, sqlArgsList [][]any)
	IndexDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, indexName string) (err error)

	SqlConcat(args ...string) (res string)
}

type DDLHandler interface {
	WrapUserName(userName string) (res string)
	IsAppendDatabaseName() (res bool)
	WrapDatabaseName(databaseName string) (res string)
	IsAppendSchemaName() (res bool)
	WrapSchemaName(schemaName string) (res string)
	WrapTableName(tableName string) (res string)
	WrapColumnName(columnName string) (res string)
	WrapConstraintName(constraintName string) (res string)
	WrapIndexName(indexName string) (res string)
	WrapSequenceName(sequenceName string) (res string)
	WrapName(name string) (res string)
	WrapString(str string) (res string)
}

func NewDDLOption() (res *DDLOption) {
	res = &DDLOption{}
	return
}

type WrapOption struct {
	// 是否 包装 名称
	WrapNameOpen *bool `json:"wrapNameOpen,omitempty"`
	// 包装 名称 的字符
	WrapNameChar *string `json:"wrapNameChar,omitempty"`
	// 是否 包装 表名称 该配置 > NameWrap
	WrapTableNameOpen *bool `json:"wrapTableNameOpen,omitempty"`
	// 是否 包装 字段名称 该配置 > NameWrap
	WrapColumnNameOpen *bool `json:"wrapColumnNameOpen,omitempty"`

	// 包装 字符串的 字符
	WrapStringChar *string `json:"wrapStringChar,omitempty"`
}

func (this_ WrapOption) Set(fn func(o WrapOption)) WrapOption {
	fn(this_)
	return this_
}
func (this_ WrapOption) IsWrapName() bool {
	if this_.WrapNameOpen != nil {
		return *this_.WrapNameOpen
	}
	return false
}
func (this_ WrapOption) IsWrapTableName() bool {
	if this_.WrapTableNameOpen != nil {
		return *this_.WrapTableNameOpen
	}
	return this_.IsWrapName()
}
func (this_ WrapOption) IsWrapColumnName() bool {
	if this_.WrapColumnNameOpen != nil {
		return *this_.WrapColumnNameOpen
	}
	return this_.IsWrapName()
}

func (this_ WrapOption) WrapTableName(name string) string {
	if this_.IsWrapTableName() {
		return this_.WrapName(name)
	}
	return name
}
func (this_ WrapOption) WrapColumnName(name string) string {
	if this_.IsWrapColumnName() {
		return this_.WrapName(name)
	}
	return name
}
func (this_ WrapOption) WrapName(name string) (res string) {
	var wrapChar string
	if this_.WrapNameChar != nil {
		wrapChar = *this_.WrapNameChar
	}
	return wrapChar + name + wrapChar
}
func (this_ WrapOption) Names(names []string) (res string) {
	return strings.Join(names, ",")
}
func (this_ WrapOption) WrapNames(names []string) (res string) {
	var ss []string
	for _, name := range names {
		ss = append(ss, this_.WrapName(name))
	}
	return strings.Join(ss, ",")
}

func (this_ WrapOption) WrapString(str string) (res string) {
	var wrapChar string
	if this_.WrapStringChar != nil {
		wrapChar = *this_.WrapStringChar
	}
	return wrapChar + str + wrapChar
}

type DDLOption struct {
	WrapOption
	// 是否 包装 用户名称 该配置 > NameWrap
	WrapUserNameOpen *bool `json:"wrapUserNameOpen,omitempty"`
	// 是否 包装 数据库名称 该配置 > NameWrap
	WrapDatabaseNameOpen *bool `json:"wrapDatabaseNameOpen,omitempty"`
	// 是否 包装 模式名称 该配置 > NameWrap
	WrapSchemaNameOpen *bool `json:"wrapSchemaNameOpen,omitempty"`
	// 是否 包装 约束名称 该配置 > NameWrap
	WrapConstraintNameOpen *bool `json:"wrapConstraintNameOpen,omitempty"`
	// 是否 包装 索引名称 该配置 > NameWrap
	WrapIndexNameOpen *bool `json:"wrapIndexNameOpen,omitempty"`
	// 是否 包装 序列名称 该配置 > NameWrap
	WrapSequenceNameOpen *bool `json:"wrapSequenceNameOpen,omitempty"`

	// 是否 追加 库名
	AppendDatabaseName *bool `json:"appendDatabaseName,omitempty"`
	// 是否 追加 模式名
	AppendSchemaName *bool `json:"appendSchemaName,omitempty"`
}

func (this_ *DDLOption) IsWrapUserName() bool {
	if this_ != nil && this_.WrapUserNameOpen != nil {
		return *this_.WrapUserNameOpen
	}
	return this_.IsWrapName()
}
func (this_ *DDLOption) IsWrapDatabaseName() bool {
	if this_ != nil && this_.WrapDatabaseNameOpen != nil {
		return *this_.WrapDatabaseNameOpen
	}
	return this_.IsWrapName()
}
func (this_ *DDLOption) IsWrapSchemaName() bool {
	if this_ != nil && this_.WrapSchemaNameOpen != nil {
		return *this_.WrapSchemaNameOpen
	}
	return this_.IsWrapName()
}
func (this_ *DDLOption) IsWrapConstraintName() bool {
	if this_ != nil && this_.WrapConstraintNameOpen != nil {
		return *this_.WrapConstraintNameOpen
	}
	return this_.IsWrapName()
}
func (this_ *DDLOption) IsWrapIndexName() bool {
	if this_ != nil && this_.WrapIndexNameOpen != nil {
		return *this_.WrapIndexNameOpen
	}
	return this_.IsWrapName()
}
func (this_ *DDLOption) IsWrapSequenceName() bool {
	if this_ != nil && this_.WrapSequenceNameOpen != nil {
		return *this_.WrapSequenceNameOpen
	}
	return this_.IsWrapName()
}
func (this_ *DDLOption) IsAppendDatabaseName() bool {
	if this_ != nil && this_.AppendDatabaseName != nil {
		return *this_.AppendDatabaseName
	}
	return false
}
func (this_ *DDLOption) IsAppendSchemaName() bool {
	if this_ != nil && this_.AppendSchemaName != nil {
		return *this_.AppendSchemaName
	}
	return false
}
func (this_ *DDLOption) WrapUserName(name string) string {
	if this_.IsWrapUserName() {
		return this_.WrapName(name)
	}
	return name
}
func (this_ *DDLOption) WrapDatabaseName(name string) string {
	if this_.IsWrapDatabaseName() {
		return this_.WrapName(name)
	}
	return name
}
func (this_ *DDLOption) WrapSchemaName(name string) string {
	if this_.IsWrapSchemaName() {
		return this_.WrapName(name)
	}
	return name
}
func (this_ *DDLOption) WrapConstraintName(name string) string {
	if this_.IsWrapConstraintName() {
		return this_.WrapName(name)
	}
	return name
}
func (this_ *DDLOption) WrapIndexName(name string) string {
	if this_.IsWrapIndexName() {
		return this_.WrapName(name)
	}
	return name
}
func (this_ *DDLOption) WrapSequenceName(name string) string {
	if this_.IsWrapSequenceName() {
		return this_.WrapName(name)
	}
	return name
}

type User struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`

	IsSystem bool `json:"isSystem,omitempty"`

	// 扩展属性
	Extend map[string]any `json:"param,omitempty"`
}

func UserByMap(data map[string]any) (res *User) {
	res = &User{}
	res.Extend = make(map[string]any)
	for key, value := range data {
		keyName := strings.ToLower(key)
		switch keyName {
		case "name", "user_name":
			res.Name = GetStringValue(value)
		default:
			res.Extend[keyName] = value
		}
	}
	return
}

type Database struct {
	Name string `json:"name,omitempty"`

	IsSystem bool `json:"isSystem,omitempty"`

	IfNotExists bool `json:"ifNotExists,omitempty"`
	// 多数用于删除时候 串联删除相关内容
	Cascade bool `json:"cascade,omitempty"`
	// 扩展属性
	Extend map[string]any `json:"param,omitempty"`
}

func DatabaseByMap(data map[string]any) (res *Database) {
	res = &Database{}
	res.Extend = make(map[string]any)
	for key, value := range data {
		keyName := strings.ToLower(key)
		switch keyName {
		case "name", "database_name":
			res.Name = GetStringValue(value)
		default:
			res.Extend[keyName] = value
		}
	}
	return
}

type Schema struct {
	DatabaseName string `json:"databaseName,omitempty"`

	Name string `json:"name,omitempty"`

	IsSystem bool `json:"isSystem,omitempty"`

	IfNotExists bool `json:"ifNotExists,omitempty"`
	// 多数用于删除时候 串联删除相关内容
	Cascade bool `json:"cascade,omitempty"`
	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`
}

func SchemaByMap(data map[string]any) (res *Schema) {
	res = &Schema{}
	res.Extend = make(map[string]any)
	for key, value := range data {

		keyName := strings.ToLower(key)
		switch keyName {
		case "database_name":
			res.DatabaseName = GetStringValue(value)

		case "name", "schema_name":
			res.Name = GetStringValue(value)
		default:
			res.Extend[keyName] = value
		}
	}
	return
}

type Sequence struct {
	DatabaseName string `json:"databaseName,omitempty"`
	SchemaName   string `json:"schemaName,omitempty"`

	Name string `json:"name,omitempty"`

	// 序列 开始值
	Start int64 `json:"start,omitempty"`
	// 序列 每次增长
	Increment int64 `json:"increment,omitempty"`
	Min       int64 `json:"min,omitempty"`
	Max       int64 `json:"max,omitempty"`

	IfNotExists bool `json:"ifNotExists,omitempty"`
	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`
}

func SequenceByMap(data map[string]any) (res *Sequence) {
	res = &Sequence{}
	res.Extend = make(map[string]any)
	for key, value := range data {

		keyName := strings.ToLower(key)
		switch keyName {
		case "database_name":
			res.DatabaseName = GetStringValue(value)
		case "schema_name":
			res.SchemaName = GetStringValue(value)

		case "name", "sequence_name":
			res.Name = GetStringValue(value)
		case "start", "sequence_start":
			res.Start, _ = ToInt64Value(value)
		case "increment", "sequence_increment":
			res.Increment, _ = ToInt64Value(value)
		case "min", "sequence_min":
			res.Min, _ = ToInt64Value(value)
		case "max", "sequence_max":
			res.Max, _ = ToInt64Value(value)
		default:
			res.Extend[keyName] = value
		}
	}
	return
}

func (this_ *Sequence) GenName(handler DDLHandler, databaseName string, schemaName string, tableName string, columnName string) string {
	var tb string
	if schemaName != "" {
		tb += schemaName + "_"
	} else if databaseName != "" {
		tb += databaseName + "_"
	}
	tb += tableName
	var t = "seq"
	return fmt.Sprintf("%s_%s_%s", t, tb, columnName)
}
func (this_ *Sequence) Init(handler DDLHandler, databaseName string, schemaName string, tableName string, columnName string) {
	if this_.Name == "" {
		this_.Name = this_.GenName(handler, databaseName, schemaName, tableName, columnName)
	}
}

type Table struct {
	DatabaseName string `json:"databaseName,omitempty"`
	SchemaName   string `json:"schemaName,omitempty"`

	Name    string `json:"name,omitempty"`
	Comment string `json:"comment,omitempty"`

	AutoIncrement       bool   `json:"autoIncrement,omitempty"`
	AutoIncrementStart  int64  `json:"autoIncrementStart,omitempty"`
	AutoIncrementName   string `json:"autoIncrementName,omitempty"`
	AutoIncrementCreate bool   `json:"autoIncrementCreate,omitempty"`

	ColumnList     []*Column     `json:"columnList,omitempty"`
	ConstraintList []*Constraint `json:"constraintList,omitempty"`
	IndexList      []*Index      `json:"indexList,omitempty"`

	IfNotExists bool `json:"ifNotExists,omitempty"`
	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`
}

func TableByMap(data map[string]any) (res *Table) {
	res = &Table{}
	res.Extend = make(map[string]any)
	for key, value := range data {
		keyName := strings.ToLower(key)
		switch keyName {
		case "database_name":
			res.DatabaseName = GetStringValue(value)
		case "schema_name":
			res.SchemaName = GetStringValue(value)

		case "name", "table_name":
			res.Name = GetStringValue(value)
		case "comment", "table_comment":
			res.Comment = GetStringValue(value)
		case "auto_increment":
			if value != nil {
				res.AutoIncrementStart, _ = ToInt64Value(value)
				if res.AutoIncrementStart > 0 {
					res.AutoIncrement = true
				}
			}
		default:
			res.Extend[keyName] = value
		}
	}
	return
}

func (this_ *Column) Init(handler DDLHandler, databaseName string, schemaName string, tableName string) {

}
func (this_ *Constraint) Init(handler DDLHandler, databaseName string, schemaName string, tableName string) {
	switch strings.ToLower(this_.Type) {
	case "primary", "primary key":
		this_.IsPrimary = true
	case "unique", "unique key":
		this_.IsUnique = true
	case "foreign", "foreign key":
		this_.IsForeign = true
	}
	if this_.IsPrimary {
		this_.Type = "primary"
	} else if this_.IsUnique {
		this_.Type = "unique"
	} else if this_.IsForeign {
		this_.Type = "foreign"
	}
	if this_.Name == "" {
		this_.Name = this_.GenName(databaseName, schemaName, tableName)
	}
}
func (this_ *Constraint) GenName(databaseName string, schemaName string, tableName string) string {
	var tb string
	if schemaName != "" {
		tb += schemaName + "_"
	} else if databaseName != "" {
		tb += databaseName + "_"
	}
	tb += tableName
	var t string
	if this_.IsPrimary {
		t = "p"
	} else if this_.IsUnique {
		t = "u"
	} else if this_.IsForeign {
		t = "f"
	} else {
		t = "c"
	}
	var cs = strings.Join(this_.ColumnNames, "_")
	return fmt.Sprintf("%s_%s_%s", t, tb, cs)
}
func (this_ *Index) GenName(databaseName string, schemaName string, tableName string) string {
	var tb string
	if schemaName != "" {
		tb += schemaName + "_"
	} else if databaseName != "" {
		tb += databaseName + "_"
	}
	tb += tableName
	var t = "idx"
	var cs = strings.Join(this_.ColumnNames, "_")
	return fmt.Sprintf("%s_%s_%s", t, tb, cs)
}
func (this_ *Index) Init(handler DDLHandler, databaseName string, schemaName string, tableName string) {
	if this_.Name == "" {
		this_.Name = this_.GenName(databaseName, schemaName, tableName)
	}
}
func (this_ *Table) Init(databaseName string, schemaName string, handler DDLHandler) {

	for _, one := range this_.ColumnList {
		one.Init(handler, databaseName, schemaName, this_.Name)
	}
	var primaryConstraint *Constraint
	for _, one := range this_.ConstraintList {
		one.Init(handler, databaseName, schemaName, this_.Name)
		if one.IsPrimary {
			primaryConstraint = one
			for _, name := range one.ColumnNames {
				c := this_.FindColumn(name)
				if c != nil {
					c.Key = true
				}
			}
		}
	}
	var keyColumnNames []string
	var keyColumns []*Column
	var autoIncrementColumns []*Column
	for _, one := range this_.ColumnList {
		if one.Key {
			keyColumnNames = append(keyColumnNames, one.Name)
			keyColumns = append(keyColumns, one)
		}
		if one.AutoIncrement {
			autoIncrementColumns = append(autoIncrementColumns, one)
		}
	}
	if primaryConstraint == nil {
		if len(keyColumnNames) > 0 {
			primaryConstraint = &Constraint{}
			primaryConstraint.IsPrimary = true
			primaryConstraint.ColumnNames = keyColumnNames
			this_.ConstraintList = append(this_.ConstraintList, primaryConstraint)
		}
	}

	if len(autoIncrementColumns) > 0 {
		this_.AutoIncrement = true
		this_.AutoIncrementStart = autoIncrementColumns[0].AutoIncrementStart
		this_.AutoIncrementName = autoIncrementColumns[0].AutoIncrementName
		this_.AutoIncrementCreate = autoIncrementColumns[0].AutoIncrementCreate
	} else if this_.AutoIncrement && len(keyColumns) > 0 {
		for _, one := range keyColumns {
			one.AutoIncrement = this_.AutoIncrement
			one.AutoIncrementStart = this_.AutoIncrementStart
			one.AutoIncrementName = this_.AutoIncrementName
			one.AutoIncrementCreate = this_.AutoIncrementCreate
		}
	}
	for _, one := range this_.IndexList {
		one.Init(handler, databaseName, schemaName, this_.Name)
	}

}

func (this_ *Table) FindColumn(name string) (res *Column) {
	for _, one := range this_.ColumnList {
		if strings.EqualFold(one.Name, name) {
			return one
		}
	}
	return
}

type Column struct {
	DatabaseName string `json:"databaseName,omitempty"`
	SchemaName   string `json:"schemaName,omitempty"`
	TableName    string `json:"tableName,omitempty"`

	Name      string `json:"name,omitempty"`
	Comment   string `json:"comment,omitempty"`
	Type      string `json:"type,omitempty"`
	DataType  string `json:"dataType,omitempty"`
	Default   string `json:"default,omitempty"`
	Length    int    `json:"length,omitempty"`
	Precision int    `json:"precision,omitempty"`
	Scale     int    `json:"scale,omitempty"`

	Key bool `json:"key,omitempty"`

	AutoIncrement       bool   `json:"autoIncrement,omitempty"`
	AutoIncrementStart  int64  `json:"autoIncrementStart,omitempty"`
	AutoIncrementName   string `json:"autoIncrementName,omitempty"`
	AutoIncrementCreate bool   `json:"autoIncrementCreate,omitempty"`

	NotNull bool `json:"notNull,omitempty"`

	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`
}

func ColumnByMap(data map[string]any) (res *Column) {
	res = &Column{}
	res.Extend = make(map[string]any)
	for key, value := range data {
		keyName := strings.ToLower(key)
		switch keyName {
		case "database_name":
			res.DatabaseName = GetStringValue(value)
		case "schema_name":
			res.SchemaName = GetStringValue(value)
		case "table_name":
			res.TableName = GetStringValue(value)

		case "name", "column_name":
			res.Name = GetStringValue(value)
		case "comment", "column_comment":
			res.Comment = GetStringValue(value)
		case "type", "column_type":
			res.Type = GetStringValue(value)
		case "data_type":
			res.DataType = GetStringValue(value)
		case "length", "data_length", "column_length", "numeric_length", "character_maximum_length":
			if num, _ := ToIntValue(value); num != 0 {
				res.Length = num
			}
		case "precision", "data_precision", "column_precision", "numeric_precision", "datetime_precision":
			if num, _ := ToIntValue(value); num != 0 {
				res.Precision = num
			}
		case "scale", "data_scale", "column_scale", "numeric_scale":
			if num, _ := ToIntValue(value); num != 0 {
				res.Scale = num
			}
		case "default", "column_default":
			res.Default = GetStringValue(value)
		case "is_nullable":
			nullable := GetStringValue(value)
			if strings.EqualFold(nullable, "no") || strings.EqualFold(nullable, "n") {
				res.NotNull = true
			}
		case "is_not_null":
			res.NotNull = IsTrue(value)
		default:
			if value != nil && value != "" {
				res.Extend[keyName] = value
			}
		}
	}
	if res.Type != "" {
		//dataType, l, p, s := FormatColumnType(res.Type)
		//res.DataType = dataType
		//res.Length = l
		//res.Precision = p
		//res.Scale = s
	}
	if strings.HasPrefix(res.Default, "nextval(") {
		sI := strings.Index(res.Default, "'")
		eI := strings.LastIndex(res.Default, "'")
		if sI > 0 && eI > sI {
			seqName := res.Default[sI+1 : eI]
			res.AutoIncrementName = seqName
			res.Default = ""
		}
	} else if strings.HasSuffix(strings.ToUpper(res.Default), ".NEXTVAL") {
		eI := strings.LastIndex(strings.ToUpper(res.Default), ".NEXTVAL")
		seqName := res.Default[0:eI]
		seqName = strings.TrimPrefix(seqName, "\"")
		seqName = strings.TrimSuffix(seqName, "\"")
		res.AutoIncrementName = seqName
		res.Default = ""
	}
	return
}

type ConstraintColumn struct {
	DatabaseName string `json:"databaseName,omitempty"`
	SchemaName   string `json:"schemaName,omitempty"`
	TableName    string `json:"tableName,omitempty"`

	Name        string `json:"name,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Type        string `json:"type,omitempty"`
	ColumnName  string `json:"columnName,omitempty"`
	ColumnOrder int    `json:"columnOrder,omitempty"`

	// 引用表所在库
	ReferencedDatabaseName string `json:"referencedDatabaseName,omitempty"`
	// 引用表
	ReferencedTableName string `json:"referencedTableName,omitempty"`
	// 引用字段
	ReferencedColumnName string `json:"referencedColumnName,omitempty"`

	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`
}

func ConstraintColumnByMap(data map[string]any) (res *ConstraintColumn) {
	res = &ConstraintColumn{}
	res.Extend = make(map[string]any)
	for key, value := range data {
		keyName := strings.ToLower(key)
		switch keyName {
		case "database_name":
			res.DatabaseName = GetStringValue(value)
		case "schema_name":
			res.SchemaName = GetStringValue(value)
		case "table_name":
			res.TableName = GetStringValue(value)

		case "name", "constraint_name":
			res.Name = GetStringValue(value)
		case "constraint_type":
			res.Type = GetStringValue(value)
		case "comment", "constraint_comment":
			res.Comment = GetStringValue(value)
		case "column_name":
			res.ColumnName = GetStringValue(value)
		case "column_order":
			res.ColumnOrder, _ = ToIntValue(value)

		case "referenced_database_name":
			res.ReferencedDatabaseName = GetStringValue(value)
		case "referenced_table_name":
			res.ReferencedTableName = GetStringValue(value)
		case "referenced_column_name":
			res.ReferencedColumnName = GetStringValue(value)
		default:
			if value != nil && value != "" {
				res.Extend[keyName] = value
			}
		}
	}
	return
}

type Constraint struct {
	DatabaseName string `json:"databaseName,omitempty"`
	SchemaName   string `json:"schemaName,omitempty"`
	TableName    string `json:"tableName,omitempty"`

	Name        string   `json:"name,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Type        string   `json:"type,omitempty"`
	ColumnNames []string `json:"columnNames,omitempty"`

	// 引用表所在库
	ReferencedDatabaseName string `json:"referencedDatabaseName,omitempty"`
	// 引用表
	ReferencedTableName string `json:"referencedTableName,omitempty"`
	// 引用字段
	ReferencedColumnNames []string `json:"referencedColumnNames,omitempty"`

	IsPrimary bool `json:"isPrimary,omitempty"`
	IsUnique  bool `json:"isUnique,omitempty"`
	IsForeign bool `json:"isForeign,omitempty"`

	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`

	columnList []*ConstraintColumn
}

func ConstraintListByColumns(columns []*ConstraintColumn) (res []*Constraint) {
	cache := map[string]*Constraint{}

	for _, one := range columns {
		key := fmt.Sprintf("%s.%s.%s.%s", one.DatabaseName, one.SchemaName, one.TableName, one.Name)
		if one.Name == "" {
			key = fmt.Sprintf("%s.%s.%s.%s.%s.%s", one.DatabaseName, one.SchemaName, one.TableName, one.Type, one.ReferencedTableName, one.ReferencedColumnName)
		}
		constraint := cache[key]
		if constraint == nil {
			constraint = &Constraint{}
			cache[key] = constraint
			res = append(res, constraint)

			constraint.Extend = make(map[string]any)
			constraint.DatabaseName = one.DatabaseName
			constraint.SchemaName = one.SchemaName
			constraint.TableName = one.TableName
			constraint.Name = one.Name
			constraint.Comment = one.Comment
			constraint.Type = one.Type
			constraint.ReferencedDatabaseName = one.ReferencedDatabaseName
			constraint.ReferencedTableName = one.ReferencedTableName

			switch strings.ToLower(one.Type) {
			case "primary", "primary key", "p":
				constraint.IsPrimary = true
				constraint.Type = "primary"
			case "unique", "unique key", "u":
				constraint.IsUnique = true
				constraint.Type = "unique"
			case "foreign", "foreign key", "r":
				constraint.IsForeign = true
				constraint.Type = "foreign"
			}
		}
		constraint.columnList = append(constraint.columnList, one)
		if one.Extend != nil {
			for k, v := range one.Extend {
				constraint.Extend[k] = v
			}
		}
	}
	for _, one := range res {
		sort.Slice(one.columnList, func(i, j int) bool {
			return one.columnList[i].ColumnOrder < one.columnList[j].ColumnOrder
		})
		for _, c := range one.columnList {
			one.ColumnNames = append(one.ColumnNames, c.ColumnName)
			if c.ReferencedColumnName != "" {
				one.ReferencedColumnNames = append(one.ReferencedColumnNames, c.ReferencedColumnName)
			}
		}
	}
	return
}

type IndexColumn struct {
	DatabaseName string `json:"databaseName,omitempty"`
	SchemaName   string `json:"schemaName,omitempty"`
	TableName    string `json:"tableName,omitempty"`

	Name        string `json:"name,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Type        string `json:"type,omitempty"`
	ColumnName  string `json:"columnName,omitempty"`
	ColumnOrder int    `json:"columnOrder,omitempty"`

	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`
}

func IndexColumnByMap(data map[string]any) (res *IndexColumn) {
	//fmt.Println("IndexColumnByMap data:", data)
	res = &IndexColumn{}
	res.Extend = make(map[string]any)
	for key, value := range data {
		keyName := strings.ToLower(key)
		switch keyName {
		case "database_name":
			res.DatabaseName = GetStringValue(value)
		case "schema_name":
			res.SchemaName = GetStringValue(value)
		case "table_name":
			res.TableName = GetStringValue(value)

		case "name", "index_name":
			res.Name = GetStringValue(value)
		case "index_type":
			res.Type = GetStringValue(value)
		case "comment", "index_comment":
			res.Comment = GetStringValue(value)

		case "column_name":
			res.ColumnName = GetStringValue(value)
		case "column_order":
			res.ColumnOrder, _ = ToIntValue(value)
		case "is_primary":
		case "is_unique":
		case "is_exclusion":

		default:
			if value != nil && value != "" {
				res.Extend[keyName] = value
			}
		}
	}
	return
}

type Index struct {
	DatabaseName string `json:"databaseName,omitempty"`
	SchemaName   string `json:"schemaName,omitempty"`
	TableName    string `json:"tableName,omitempty"`

	Name        string   `json:"name,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Type        string   `json:"type,omitempty"`
	ColumnNames []string `json:"columnNames,omitempty"`

	// 扩展属性
	Extend map[string]any `json:"extend,omitempty"`

	columnList []*IndexColumn
}

func IndexListByColumns(columns []*IndexColumn) (res []*Index) {
	cache := map[string]*Index{}

	for _, one := range columns {
		key := fmt.Sprintf("%s.%s.%s.%s", one.DatabaseName, one.SchemaName, one.TableName, one.Name)
		index := cache[key]
		if index == nil {
			index = &Index{}
			cache[key] = index
			res = append(res, index)

			index.Extend = make(map[string]any)
			index.DatabaseName = one.DatabaseName
			index.SchemaName = one.SchemaName
			index.TableName = one.TableName
			index.Name = one.Name
			index.Comment = one.Comment
			index.Type = one.Type
		}
		index.columnList = append(index.columnList, one)
		if one.Extend != nil {
			for k, v := range one.Extend {
				index.Extend[k] = v
			}
		}
	}

	for _, one := range res {
		sort.Slice(one.columnList, func(i, j int) bool {
			return one.columnList[i].ColumnOrder < one.columnList[j].ColumnOrder
		})
		for _, c := range one.columnList {
			one.ColumnNames = append(one.ColumnNames, c.ColumnName)
		}
	}
	return
}
