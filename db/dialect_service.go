package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/team-ide/framework"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

func NewDialect(cfg *DialectConfig) Dialect {
	res := &DialectService{}
	res.cfg = cfg

	return res
}

type DialectService struct {
	cfg *DialectConfig
}

func (this_ *DialectService) Type() string {
	return this_.cfg.Type
}

func (this_ *DialectService) Match() []string {
	return this_.cfg.Match
}

func (this_ *DialectService) WrapNameChar() string {
	return this_.cfg.GetNameWrapChar()
}
func (this_ *DialectService) WrapStringChar() string {
	return this_.cfg.GetStringWrapChar()
}
func (this_ *DialectService) EscapeStringChar() string {
	return this_.cfg.GetStringEscapeChar()
}
func (this_ *DialectService) ArgChar() string {
	return this_.cfg.GetArgChar()
}

func (this_ *DialectService) GetDriverName() string {
	driver := this_.cfg.GetDriver()
	if driver == nil {
		return ""
	}
	return driver.Name
}

func (this_ *DialectService) GetDriverDSN(dbCfg *Config, params map[string]string) string {
	driver := this_.cfg.GetDriver()
	if driver != nil {
		return FormatDriverDSN(driver.Dsn, dbCfg, driver.IsUrl, driver.ParamSpaceChar, driver.CanPathEscape, driver.Params, params)
	}
	return FormatDriverDSN("", dbCfg, false, "", false, nil, params)
}

func (this_ *DialectService) Open(dbCfg *Config, params map[string]string) (sqlDb *sql.DB, err error) {
	driverName := this_.GetDriverName()
	driverDsn := this_.GetDriverDSN(dbCfg, params)
	//fmt.Println("open driver:", driverName)
	//fmt.Println("open driver dsn:", driverDsn)
	return sql.Open(driverName, driverDsn)
}

func (this_ *DialectService) Info() (info *Info) {
	info = &Info{}
	driver := this_.cfg.GetDriver()
	if driver != nil {
		info.DriverName = driver.Name
		info.Dsn = driver.Dsn
		info.DsnHasDatabase = driver.HasDatabase
		info.DsnHasSchema = driver.HasSchema
	}
	info.NameNoWrapIsUpper = this_.cfg.GetNameNoWrapIsUpper()

	info.CreateUserAutoCreateDatabase = this_.cfg.CreateUserAutoCreateDatabase()
	info.CreateUserAutoCreateSchema = this_.cfg.CreateUserAutoCreateSchema()

	info.SqlChangeUser = SqlTemplateHasContent(this_.cfg.GetUserChange())
	info.SqlChangeDatabase = SqlTemplateHasContent(this_.cfg.GetDatabaseChange())
	info.SqlChangeSchema = SqlTemplateHasContent(this_.cfg.GetSchemaChange())

	info.HasUser = SqlTemplateHasContent(this_.cfg.GetUserSelect())
	info.HasDatabase = SqlTemplateHasContent(this_.cfg.GetDatabaseSelect())
	info.HasSchema = SqlTemplateHasContent(this_.cfg.GetSchemaSelect())
	info.HasSequence = SqlTemplateHasContent(this_.cfg.GetSequenceSelect())
	return
}

func (this_ *DialectService) UserChange(sqlConn SqlConn, userName string, password string) (ok bool, err error) {
	params := make(map[string]any)
	params["user"] = userName
	params["userWrap"] = this_.GetDDLHandler(nil).WrapUserName(userName)
	params["userName"] = userName
	params["userNameWrap"] = this_.GetDDLHandler(nil).WrapUserName(userName)
	params["password"] = password
	sqlList, sqlArgsList := this_.TemplateSqlList(this_.cfg.GetUserChange(), params)
	if len(sqlList) == 0 {
		return
	}
	for i, sqlInfo := range sqlList {
		_, err = sqlConn.ExecContext(context.Background(), sqlInfo, sqlArgsList[i]...)
		if err != nil {
			return
		}
	}
	ok = true
	return
}

func (this_ *DialectService) DatabaseChange(sqlConn SqlConn, databaseName string) (ok bool, err error) {
	params := make(map[string]any)
	params["database"] = databaseName
	params["databaseWrap"] = this_.GetDDLHandler(nil).WrapDatabaseName(databaseName)
	params["databaseName"] = databaseName
	params["databaseNameWrap"] = this_.GetDDLHandler(nil).WrapDatabaseName(databaseName)
	sqlList, sqlArgsList := this_.TemplateSqlList(this_.cfg.GetDatabaseChange(), params)
	if len(sqlList) == 0 {
		return
	}
	err = this_.Execs(sqlConn, sqlList, sqlArgsList)
	if err != nil {
		return
	}
	ok = true
	return
}

func (this_ *DialectService) SchemaChange(sqlConn SqlConn, schemaName string) (ok bool, err error) {
	params := make(map[string]any)
	params["schema"] = schemaName
	params["schemaWrap"] = this_.GetDDLHandler(nil).WrapSchemaName(schemaName)
	params["schemaName"] = schemaName
	params["schemaNameWrap"] = this_.GetDDLHandler(nil).WrapSchemaName(schemaName)
	sqlList, sqlArgsList := this_.TemplateSqlList(this_.cfg.GetSchemaChange(), params)
	if len(sqlList) == 0 {
		return
	}
	err = this_.Execs(sqlConn, sqlList, sqlArgsList)
	if err != nil {
		return
	}
	ok = true
	return
}

func (this_ *DialectService) UserSelectSql(userName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["userName"] = userName
	return this_.TemplateSql(this_.cfg.GetUserSelect(), param)
}
func (this_ *DialectService) UserList(sqlConn SqlConn) (userList []*User, err error) {
	sqlInfo, sqlArgs := this_.UserSelectSql("")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	for _, row := range list {
		user := UserByMap(row)
		user.IsSystem = this_.cfg.IsSystemUser(user.Name)
		userList = append(userList, user)
	}
	return
}
func (this_ *DialectService) User(sqlConn SqlConn, userName string) (user *User, err error) {
	sqlInfo, sqlArgs := this_.UserSelectSql(userName)
	if len(sqlInfo) == 0 {
		return
	}
	row, err := this_.QueryOne(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	if row != nil {
		user = UserByMap(row)
		user.IsSystem = this_.cfg.IsSystemUser(user.Name)
	}
	return
}
func (this_ *DialectService) UserParam(handler DDLHandler, user *User) (param map[string]any) {
	param = this_.GetParam(handler, user)
	return
}
func (this_ *DialectService) UserCreateFieldList() (sqlList []*Field) {
	return this_.cfg.GetUserCreateFieldList()
}
func (this_ *DialectService) UserCreateSql(handler DDLHandler, user *User) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetUserCreate(), this_.UserParam(handler, user))
}
func (this_ *DialectService) UserCreate(sqlConn SqlConn, handler DDLHandler, user *User) (err error) {
	sqlList, sqlArgsList := this_.UserCreateSql(handler, user)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) UserDeleteSql(handler DDLHandler, userName string) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetUserDelete(), this_.UserParam(handler, &User{Name: userName}))
}
func (this_ *DialectService) UserDelete(sqlConn SqlConn, handler DDLHandler, userName string) (err error) {
	sqlList, sqlArgsList := this_.UserDeleteSql(handler, userName)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) DatabaseSelectSql(databaseName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["databaseName"] = databaseName
	return this_.TemplateSql(this_.cfg.GetDatabaseSelect(), param)
}
func (this_ *DialectService) DatabaseList(sqlConn SqlConn) (databaseList []*Database, err error) {
	sqlInfo, sqlArgs := this_.DatabaseSelectSql("")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	for _, row := range list {
		database := DatabaseByMap(row)
		database.IsSystem = this_.cfg.IsSystemDatabase(database.Name)
		databaseList = append(databaseList, database)
	}
	return
}
func (this_ *DialectService) Database(sqlConn SqlConn, databaseName string) (database *Database, err error) {
	sqlInfo, sqlArgs := this_.DatabaseSelectSql(databaseName)
	if len(sqlInfo) == 0 {
		return
	}
	row, err := this_.QueryOne(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	if row != nil {
		database = DatabaseByMap(row)
		database.IsSystem = this_.cfg.IsSystemDatabase(database.Name)
	}
	return
}
func (this_ *DialectService) DatabaseParam(handler DDLHandler, database *Database) (param map[string]any) {
	param = this_.GetParam(handler, database)
	return
}
func (this_ *DialectService) DatabaseCreateFieldList() (sqlList []*Field) {
	return this_.cfg.GetDatabaseCreateFieldList()
}
func (this_ *DialectService) DatabaseCreateSql(handler DDLHandler, database *Database) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetDatabaseCreate(), this_.DatabaseParam(handler, database))
}
func (this_ *DialectService) DatabaseCreate(sqlConn SqlConn, handler DDLHandler, database *Database) (err error) {
	sqlList, sqlArgsList := this_.DatabaseCreateSql(handler, database)
	err = this_.Execs(sqlConn, sqlList, sqlArgsList)
	if err != nil {
		if database.IfNotExists && strings.Contains(strings.ToLower(err.Error()), "already exists") {
			err = nil
		}
	}
	return
}
func (this_ *DialectService) DatabaseDeleteSql(handler DDLHandler, databaseName string, cascade bool) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetDatabaseDelete(), this_.DatabaseParam(handler, &Database{Name: databaseName, Cascade: cascade}))
}
func (this_ *DialectService) DatabaseDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, cascade bool) (err error) {
	sqlList, sqlArgsList := this_.DatabaseDeleteSql(handler, databaseName, cascade)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) SchemaSelectSql(databaseName string, schemaName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["databaseName"] = databaseName
	param["schemaName"] = schemaName
	return this_.TemplateSql(this_.cfg.GetSchemaSelect(), param)
}
func (this_ *DialectService) SchemaList(sqlConn SqlConn, databaseName string) (schemaList []*Schema, err error) {
	sqlInfo, sqlArgs := this_.SchemaSelectSql(databaseName, "")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	for _, row := range list {
		schema := SchemaByMap(row)
		schema.IsSystem = this_.cfg.IsSystemSchema(schema.Name)
		schemaList = append(schemaList, schema)
	}
	return
}
func (this_ *DialectService) Schema(sqlConn SqlConn, databaseName string, schemaName string) (schema *Schema, err error) {
	sqlInfo, sqlArgs := this_.SchemaSelectSql(databaseName, schemaName)
	if len(sqlInfo) == 0 {
		return
	}
	row, err := this_.QueryOne(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	if row != nil {
		schema = SchemaByMap(row)
		schema.IsSystem = this_.cfg.IsSystemSchema(schema.Name)
	}
	return
}

func (this_ *DialectService) SchemaParam(handler DDLHandler, databaseName string, schema *Schema) (param map[string]any) {
	param = this_.GetParam(handler, schema)
	delete(param, "databaseName")
	delete(param, "databaseNameWrap")
	if databaseName != "" && this_.GetDDLHandler(handler).IsAppendDatabaseName() {
		param["databaseName"] = databaseName
		param["databaseNameWrap"] = this_.GetDDLHandler(handler).WrapDatabaseName(databaseName)
	}
	return
}
func (this_ *DialectService) SchemaCreateFieldList() (sqlList []*Field) {
	return this_.cfg.GetSchemaCreateFieldList()
}
func (this_ *DialectService) SchemaCreateSql(handler DDLHandler, databaseName string, schema *Schema) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetSchemaCreate(), this_.SchemaParam(handler, databaseName, schema))
}
func (this_ *DialectService) SchemaCreate(sqlConn SqlConn, handler DDLHandler, databaseName string, schema *Schema) (err error) {
	sqlList, sqlArgsList := this_.SchemaCreateSql(handler, databaseName, schema)
	err = this_.Execs(sqlConn, sqlList, sqlArgsList)
	if err != nil {
		if schema.IfNotExists && strings.Contains(strings.ToLower(err.Error()), "already exists") {
			err = nil
		}
	}
	return
}
func (this_ *DialectService) SchemaDeleteSql(handler DDLHandler, databaseName string, schemaName string, cascade bool) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetSchemaDelete(), this_.SchemaParam(handler, databaseName, &Schema{Name: schemaName, Cascade: cascade}))
}
func (this_ *DialectService) SchemaDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, cascade bool) (err error) {
	sqlList, sqlArgsList := this_.SchemaDeleteSql(handler, databaseName, schemaName, cascade)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) SequenceSelectSql(databaseName string, schemaName string, sequenceName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["databaseName"] = databaseName
	param["schemaName"] = schemaName
	param["sequenceName"] = sequenceName
	return this_.TemplateSql(this_.cfg.GetSequenceSelect(), param)
}
func (this_ *DialectService) SequenceList(sqlConn SqlConn, databaseName string, schemaName string) (sequenceList []*Sequence, err error) {
	sqlInfo, sqlArgs := this_.SequenceSelectSql(databaseName, schemaName, "")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	for _, row := range list {
		sequence := SequenceByMap(row)
		sequenceList = append(sequenceList, sequence)
	}
	return
}
func (this_ *DialectService) Sequence(sqlConn SqlConn, databaseName string, schemaName string, sequenceName string) (sequence *Sequence, err error) {
	sqlInfo, sqlArgs := this_.SequenceSelectSql(databaseName, schemaName, sequenceName)
	if len(sqlInfo) == 0 {
		return
	}
	row, err := this_.QueryOne(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	if row != nil {
		sequence = SequenceByMap(row)
	}
	return
}
func (this_ *DialectService) SequenceParam(handler DDLHandler, databaseName string, schemaName string, sequence *Sequence) (param map[string]any) {
	param = this_.GetParam(handler, sequence)

	delete(param, "databaseName")
	delete(param, "databaseNameWrap")
	if databaseName != "" && this_.GetDDLHandler(handler).IsAppendDatabaseName() {
		param["databaseName"] = databaseName
		param["databaseNameWrap"] = this_.GetDDLHandler(handler).WrapDatabaseName(databaseName)
	}
	delete(param, "schemaName")
	delete(param, "schemaNameWrap")
	if schemaName != "" && this_.GetDDLHandler(handler).IsAppendSchemaName() {
		param["schemaName"] = schemaName
		param["schemaNameWrap"] = this_.GetDDLHandler(handler).WrapDatabaseName(schemaName)
	}
	return
}
func (this_ *DialectService) SequenceCreateFieldList() (sqlList []*Field) {
	return this_.cfg.GetSequenceCreateFieldList()
}
func (this_ *DialectService) SequenceCreateSql(handler DDLHandler, databaseName string, schemaName string, sequence *Sequence) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetSequenceCreate(), this_.SequenceParam(handler, databaseName, schemaName, sequence))
}
func (this_ *DialectService) SequenceCreate(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, sequence *Sequence) (err error) {
	sqlList, sqlArgsList := this_.SequenceCreateSql(handler, databaseName, schemaName, sequence)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) SequenceDeleteSql(handler DDLHandler, databaseName string, schemaName string, sequenceName string) (sqlList []string, sqlArgsList [][]any) {
	return this_.TemplateSqlList(this_.cfg.GetSequenceDelete(), this_.SequenceParam(handler, databaseName, schemaName, &Sequence{Name: sequenceName}))
}
func (this_ *DialectService) SequenceDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, sequenceName string) (err error) {
	sqlList, sqlArgsList := this_.SequenceDeleteSql(handler, databaseName, schemaName, sequenceName)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) TableSelectSql(databaseName string, schemaName string, tableName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["databaseName"] = databaseName
	param["schemaName"] = schemaName
	param["tableName"] = tableName
	return this_.TemplateSql(this_.cfg.GetTableSelect(), param)
}

func (this_ *DialectService) TableCheckExists(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (exists bool) {
	framework.Info("检测 表 是否存在", zap.Any("tableName", tableName))
	_, err := DoQueryCount(context.Background(), sqlConn, "SELECT COUNT(1) FROM "+tableName+" WHERE 1=2", nil, false)
	if err != nil {
		framework.Info("检测 表 不存在", zap.Any("tableName", tableName))
		return
	}
	framework.Info("检测 表 已存在", zap.Any("tableName", tableName))
	exists = true
	return
}
func (this_ *DialectService) ColumnCheckExists(sqlConn SqlConn, databaseName string, schemaName string, tableName string, columnName string) (exists bool) {
	framework.Info("检测 表 字段 是否存在", zap.Any("tableName", tableName), zap.Any("columnName", columnName))
	_, err := DoQueryCount(context.Background(), sqlConn, "SELECT COUNT("+columnName+") FROM "+tableName+" WHERE 1=2", nil, false)
	if err != nil {
		framework.Info("检测 表 字段 不存在", zap.Any("tableName", tableName), zap.Any("columnName", columnName))
		return
	}
	framework.Info("检测 表 字段 已存在", zap.Any("tableName", tableName), zap.Any("columnName", columnName))
	exists = true
	return
}
func (this_ *DialectService) IndexCheckExists(sqlConn SqlConn, databaseName string, schemaName string, tableName string, checkIndex *Index) (exists bool) {
	framework.Info("检测 索引 是否存在", zap.Any("tableName", tableName), zap.Any("checkIndex", checkIndex))
	indexList, err := this_.IndexList(sqlConn, databaseName, schemaName, tableName)
	if err != nil {
		return
	}
	columnNamesStr := strings.Join(checkIndex.ColumnNames, ",")

	for _, index := range indexList {
		indexColumnNamesStr := strings.Join(index.ColumnNames, ",")
		if strings.EqualFold(indexColumnNamesStr, columnNamesStr) {
			exists = true
			break
		}
		if index.Name != "" && strings.EqualFold(index.Name, checkIndex.Name) {
			exists = true
			break
		}
	}
	if exists {
		framework.Info("检测 索引 已存在", zap.Any("tableName", tableName), zap.Any("checkIndex", checkIndex))
	} else {
		framework.Info("检测 索引 不存在", zap.Any("tableName", tableName), zap.Any("checkIndex", checkIndex))
	}
	return
}
func (this_ *DialectService) TableList(sqlConn SqlConn, databaseName string, schemaName string) (tableList []*Table, err error) {
	sqlInfo, sqlArgs := this_.TableSelectSql(databaseName, schemaName, "")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	for _, row := range list {
		tableList = append(tableList, TableByMap(row))
	}
	return
}
func (this_ *DialectService) Table(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (table *Table, err error) {
	sqlInfo, sqlArgs := this_.TableSelectSql(databaseName, schemaName, tableName)
	if len(sqlInfo) == 0 {
		return
	}
	row, err := this_.QueryOne(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	if row != nil {
		table = TableByMap(row)
	}
	return
}
func (this_ *DialectService) TableDetailList(sqlConn SqlConn, databaseName string, schemaName string) (tableList []*Table, err error) {

	tableList, err = this_.TableList(sqlConn, databaseName, schemaName)
	if err != nil {
		return
	}
	if len(tableList) == 0 {
		return
	}
	tableCache := map[string]*Table{}
	for _, one := range tableList {
		key := fmt.Sprintf("%s.%s.%s", one.DatabaseName, one.SchemaName, one.Name)
		tableCache[key] = one
	}
	columnList, err := this_.ColumnList(sqlConn, databaseName, schemaName, "")
	if err != nil {
		return
	}
	for _, one := range columnList {
		key := fmt.Sprintf("%s.%s.%s", one.DatabaseName, one.SchemaName, one.TableName)
		if tableCache[key] == nil {
			continue
		}
		tableCache[key].ColumnList = append(tableCache[key].ColumnList, one)
	}
	constraintList, err := this_.ConstraintList(sqlConn, databaseName, schemaName, "")
	if err != nil {
		return
	}
	for _, one := range constraintList {
		key := fmt.Sprintf("%s.%s.%s", one.DatabaseName, one.SchemaName, one.TableName)
		if tableCache[key] == nil {
			continue
		}
		tableCache[key].ConstraintList = append(tableCache[key].ConstraintList, one)
	}
	indexList, err := this_.IndexList(sqlConn, databaseName, schemaName, "")
	if err != nil {
		return
	}
	for _, one := range indexList {
		key := fmt.Sprintf("%s.%s.%s", one.DatabaseName, one.SchemaName, one.TableName)
		if tableCache[key] == nil {
			continue
		}
		tableCache[key].IndexList = append(tableCache[key].IndexList, one)
	}
	return
}
func (this_ *DialectService) TableDetail(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (table *Table, err error) {
	table, err = this_.Table(sqlConn, databaseName, schemaName, tableName)
	if err != nil {
		return
	}
	if table == nil {
		return
	}
	table.ColumnList, err = this_.ColumnList(sqlConn, databaseName, schemaName, tableName)
	if err != nil {
		return
	}
	table.ConstraintList, err = this_.ConstraintList(sqlConn, databaseName, schemaName, tableName)
	if err != nil {
		return
	}
	table.IndexList, err = this_.IndexList(sqlConn, databaseName, schemaName, tableName)
	if err != nil {
		return
	}
	return
}
func (this_ *DialectService) appendSql(sqlList *[]string, sqlArgsList *[][]any, t *SqlTemplate, param map[string]any) {
	list, argsList := this_.TemplateSqlList(t, param)
	for i, s := range list {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		*sqlList = append(*sqlList, s)
		*sqlArgsList = append(*sqlArgsList, argsList[i])
	}
}
func (this_ *DialectService) TableParam(handler DDLHandler, databaseName string, schemaName string, table *Table) (param map[string]any) {
	table.Init(databaseName, schemaName, handler)

	param = map[string]any{}
	this_.AppendParam(handler, param, table)

	delete(param, "databaseName")
	delete(param, "databaseNameWrap")
	if databaseName != "" && this_.GetDDLHandler(handler).IsAppendDatabaseName() {
		param["databaseName"] = databaseName
		param["databaseNameWrap"] = this_.GetDDLHandler(handler).WrapDatabaseName(databaseName)
	}
	delete(param, "schemaName")
	delete(param, "schemaNameWrap")
	if schemaName != "" && this_.GetDDLHandler(handler).IsAppendSchemaName() {
		param["schemaName"] = schemaName
		param["schemaNameWrap"] = this_.GetDDLHandler(handler).WrapDatabaseName(schemaName)
	}

	var contentSqlList []string
	var contentSqlArgsList [][]any
	for _, one := range table.ColumnList {
		oneParam := this_.ColumnParam(handler, databaseName, schemaName, table.Name, one)
		this_.appendSql(&contentSqlList, &contentSqlArgsList, this_.cfg.GetTableCreateColumn(), oneParam)
	}
	for _, one := range table.ConstraintList {
		oneParam := this_.ConstraintParam(handler, databaseName, schemaName, table.Name, one)
		this_.AppendParam(handler, oneParam, table, "name", "databaseName", "schemaName")
		this_.appendSql(&contentSqlList, &contentSqlArgsList, this_.cfg.GetTableCreateConstraint(), oneParam)
	}
	for _, one := range table.IndexList {
		oneParam := this_.IndexParam(handler, databaseName, schemaName, table.Name, one)
		this_.AppendParam(handler, oneParam, table, "name", "databaseName", "schemaName")
		this_.appendSql(&contentSqlList, &contentSqlArgsList, this_.cfg.GetTableCreateIndex(), oneParam)
	}
	var content = strings.Join(contentSqlList, ",\n  ")
	if content != "" {
		content = "  " + content
	}
	param["content"] = content
	return
}
func (this_ *DialectService) GetColumnType(handler DDLHandler, column *Column) (columnType string) {
	columnType = column.Type
	if strings.Contains(columnType, "(") {
		return
	}
	if columnType == "" && column.DataType != "" {
		columnType = column.DataType
	}
	t := this_.cfg.MatchType(columnType, column.Length, column.Precision, column.Scale)
	if t != nil {
		columnType = t.FormatColumnType(column.Length, column.Precision, column.Scale)
	}
	return
}
func (this_ *DialectService) GetColumnDefault(handler DDLHandler, column *Column) (res string) {
	if column.Default == "" {
		return
	}
	var dataType = column.Type
	if dataType == "" {
		dataType = column.DataType
	}
	if strings.Index(dataType, "(") > 0 {
		dataType = dataType[0:strings.Index(dataType, "(")]
	}
	var isInteger bool
	dataType = strings.ToLower(strings.TrimSpace(dataType))
	if dataType == "" {
		isInteger = isAllPositiveNegativeDecimal(column.Default)
	} else {
		if strings.HasPrefix(dataType, "int") ||
			strings.HasSuffix(dataType, "int") ||
			strings.HasPrefix(dataType, "float") ||
			strings.HasPrefix(dataType, "serial") ||
			strings.HasSuffix(dataType, "serial") ||
			strings.EqualFold(dataType, "number") ||
			strings.EqualFold(dataType, "double") ||
			strings.EqualFold(dataType, "decimal") ||
			strings.EqualFold(dataType, "dec") ||
			strings.EqualFold(dataType, "bit") {
			isInteger = true
		}
	}
	if isInteger {
		res = column.Default
	} else {
		res = this_.GetDDLHandler(handler).WrapString(column.Default)
	}

	return
}

// isAllPositiveNegativeDecimal 判断字符串是否全部由正负小数构成
func isAllPositiveNegativeDecimal(s string) bool {
	// 正则表达式匹配一个或多个数字，可选的正负号和一个小数点
	// ^ 表示字符串开始，$ 表示字符串结束
	// [-+]? 表示可选的正负号
	// \d+ 表示一个或多个数字
	// \. 表示小数点
	// \d* 表示零个或多个数字
	regex := `^[-+]?(\d+(\.\d*)?|\.\d+)$`

	// Compile 正则表达式
	re, err := regexp.Compile(regex)
	if err != nil {
		return false
	}
	// MatchString 检查字符串是否匹配正则表达式
	return re.MatchString(s)
}
func (this_ *DialectService) ColumnParam(handler DDLHandler, databaseName string, schemaName string, tableName string, column *Column) (param map[string]any) {
	column.Init(handler, databaseName, schemaName, tableName)
	param = this_.GetParam(handler, column)
	param["columnType"] = this_.GetColumnType(handler, column)
	param["default"] = this_.GetColumnDefault(handler, column)
	this_.AppendParamTableName(handler, param, databaseName, schemaName, tableName)
	return
}
func (this_ *DialectService) ConstraintParam(handler DDLHandler, databaseName string, schemaName string, tableName string, constraint *Constraint) (param map[string]any) {
	constraint.Init(handler, databaseName, schemaName, tableName)
	param = this_.GetParam(handler, constraint)
	this_.AppendParamTableName(handler, param, databaseName, schemaName, tableName)
	return
}
func (this_ *DialectService) IndexParam(handler DDLHandler, databaseName string, schemaName string, tableName string, index *Index) (param map[string]any) {
	index.Init(handler, databaseName, schemaName, tableName)
	param = this_.GetParam(handler, index)
	this_.AppendParamTableName(handler, param, databaseName, schemaName, tableName)
	return
}
func (this_ *DialectService) TableCreateFieldList() (sqlList []*Field) {
	return this_.cfg.GetTableCreateFieldList()
}
func (this_ *DialectService) TableCreateSql(handler DDLHandler, databaseName string, schemaName string, table *Table) (sqlList []string, sqlArgsList [][]any) {
	canCreateSeq := SqlTemplateHasContent(this_.cfg.GetSequenceCreate())
	var sequenceList []*Sequence
	for _, column := range table.ColumnList {
		if column.AutoIncrement && canCreateSeq {
			sequence := &Sequence{}
			sequence.IfNotExists = true
			sequence.Name = column.AutoIncrementName
			sequence.Start = column.AutoIncrementStart
			sequence.Increment = 1
			if sequence.Increment == 0 {
				sequence.Increment = 1
			}
			sequence.Init(handler, databaseName, schemaName, table.Name, column.Name)
			column.AutoIncrementName = sequence.Name
			sequenceList = append(sequenceList, sequence)
		}
	}
	for _, sequence := range sequenceList {
		var sequenceCreateSqlList []string
		var sequenceCreateSqlArgsList [][]any
		sequenceCreateSqlList, sequenceCreateSqlArgsList = this_.SequenceCreateSql(handler, databaseName, schemaName, sequence)
		sqlList = append(sqlList, sequenceCreateSqlList...)
		sqlArgsList = append(sqlArgsList, sequenceCreateSqlArgsList...)
	}
	var param = this_.TableParam(handler, databaseName, schemaName, table)
	this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetTableCreate(), param)

	for _, column := range table.ColumnList {
		if column.AutoIncrement && canCreateSeq {
			var p = map[string]any{}
			var triggerName = table.Name + "_" + "_before_insert"
			p["name"] = triggerName
			p["nameWrap"] = this_.GetDDLHandler(handler).WrapSequenceName(triggerName)
			p["sequenceName"] = column.AutoIncrementName
			p["sequenceNameWrap"] = this_.GetDDLHandler(handler).WrapSequenceName(column.AutoIncrementName)
			p["columnName"] = column.Name
			p["columnNameWrap"] = this_.GetDDLHandler(handler).WrapColumnName(column.Name)
			sList, sArgsList := this_.TemplateSqlList(this_.cfg.GetTableCreateSequenceTrigger(), p)
			sqlList = append(sqlList, sList...)
			sqlArgsList = append(sqlArgsList, sArgsList...)
		}
	}

	if table.Comment != "" {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetTableCreateComment(), param)
	}
	for _, one := range table.ColumnList {
		oneParam := this_.ColumnParam(handler, databaseName, schemaName, table.Name, one)
		if one.Comment != "" {
			this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetTableCreateColumnComment(), oneParam)
		}
	}
	for _, one := range table.ConstraintList {
		if SqlTemplateHasContent(this_.cfg.GetTableCreateConstraint()) {
			oneParam := this_.ConstraintParam(handler, databaseName, schemaName, table.Name, one)
			if one.Comment != "" {
				this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetTableCreateConstraintComment(), oneParam)
			}
		} else {
			r1, r2 := this_.ConstraintAddSql(handler, databaseName, schemaName, table.Name, one)
			sqlList = append(sqlList, r1...)
			sqlArgsList = append(sqlArgsList, r2...)
		}
	}
	for _, one := range table.IndexList {
		if SqlTemplateHasContent(this_.cfg.GetTableCreateIndex()) {
			oneParam := this_.IndexParam(handler, databaseName, schemaName, table.Name, one)
			if one.Comment != "" {
				this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetTableCreateIndexComment(), oneParam)
			}
		} else {
			r1, r2 := this_.IndexAddSql(handler, databaseName, schemaName, table.Name, one)
			sqlList = append(sqlList, r1...)
			sqlArgsList = append(sqlArgsList, r2...)
		}
	}
	return
}

var (
	ErrorNotSupportCreateTable = errors.New("not support create table")
)

func (this_ *DialectService) TableCreate(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, table *Table) (err error) {
	sqlList, sqlArgsList := this_.TableCreateSql(handler, databaseName, schemaName, table)
	if len(sqlList) == 0 {
		err = ErrorNotSupportCreateTable
		return
	}
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) TableDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string) (sqlList []string, sqlArgsList [][]any) {
	var param = make(map[string]any)
	this_.AppendParamTableName(handler, param, databaseName, schemaName, tableName)
	param["name"] = tableName
	param["nameWrap"] = this_.GetDDLHandler(handler).WrapTableName(tableName)
	return this_.TemplateSqlList(this_.cfg.GetTableDelete(), param)
}
func (this_ *DialectService) TableDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string) (err error) {
	sqlList, sqlArgsList := this_.TableDeleteSql(handler, databaseName, schemaName, tableName)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) ColumnSelectSql(databaseName string, schemaName string, tableName string, columnName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["databaseName"] = databaseName
	param["schemaName"] = schemaName
	param["tableName"] = tableName
	param["columnName"] = columnName
	return this_.TemplateSql(this_.cfg.GetColumnSelect(), param)
}
func (this_ *DialectService) ColumnList(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (columnList []*Column, err error) {
	sqlInfo, sqlArgs := this_.ColumnSelectSql(databaseName, schemaName, tableName, "")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	for _, row := range list {
		column := ColumnByMap(row)
		err = this_.ColumnFull(sqlConn, databaseName, schemaName, tableName, column)
		if err != nil {
			return
		}
		columnList = append(columnList, column)
	}
	return
}
func (this_ *DialectService) Column(sqlConn SqlConn, databaseName string, schemaName string, tableName string, columnName string) (column *Column, err error) {
	sqlInfo, sqlArgs := this_.ColumnSelectSql(databaseName, schemaName, tableName, columnName)
	if len(sqlInfo) == 0 {
		return
	}
	row, err := this_.QueryOne(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	if row != nil {
		column = ColumnByMap(row)
		err = this_.ColumnFull(sqlConn, databaseName, schemaName, tableName, column)
		if err != nil {
			return
		}
	}
	return
}

func (this_ *DialectService) ColumnFull(sqlConn SqlConn, databaseName string, schemaName string, tableName string, column *Column) (err error) {
	if column.AutoIncrementName != "" {
		var s *Sequence
		s, err = this_.Sequence(sqlConn, databaseName, schemaName, column.AutoIncrementName)
		if err != nil {
			return
		}
		if s != nil {
			column.AutoIncrement = true
			column.AutoIncrementStart = s.Start
		}
	}
	return
}
func (this_ *DialectService) ColumnAddSql(handler DDLHandler, databaseName string, schemaName string, tableName string, column *Column) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ColumnParam(handler, databaseName, schemaName, tableName, column)
	sqlList, sqlArgsList = this_.TemplateSqlList(this_.cfg.GetColumnAdd(), param)
	if column.Comment != "" {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetColumnComment(), param)
	}
	return
}
func (this_ *DialectService) ColumnAdd(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, column *Column) (err error) {
	sqlList, sqlArgsList := this_.ColumnAddSql(handler, databaseName, schemaName, tableName, column)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) ColumnCommentSql(handler DDLHandler, databaseName string, schemaName string, tableName string, column *Column) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ColumnParam(handler, databaseName, schemaName, tableName, column)
	return this_.TemplateSqlList(this_.cfg.GetColumnComment(), param)
}
func (this_ *DialectService) ColumnComment(handler DDLHandler, databaseName string, schemaName string, tableName string, column *Column) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ColumnParam(handler, databaseName, schemaName, tableName, column)
	return this_.TemplateSqlList(this_.cfg.GetColumnComment(), param)
}
func (this_ *DialectService) ColumnUpdateSql(handler DDLHandler, databaseName string, schemaName string, tableName string, oldColumn *Column, column *Column) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ColumnParam(handler, databaseName, schemaName, tableName, column)
	param["oldName"] = oldColumn.Name
	param["oldNameWrap"] = this_.GetDDLHandler(handler).WrapColumnName(oldColumn.Name)
	param["newName"] = column.Name
	param["newNameWrap"] = this_.GetDDLHandler(handler).WrapColumnName(column.Name)

	sqlList, sqlArgsList = this_.TemplateSqlList(this_.cfg.GetColumnUpdate(), param)
	if oldColumn.Name != column.Name {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetColumnUpdateRename(), param)
	}
	oldColumnType := this_.GetColumnType(handler, oldColumn)
	newColumnType := this_.GetColumnType(handler, column)
	if oldColumnType != newColumnType {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetColumnUpdateType(), param)
	}
	if oldColumn.Default != column.Default {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetColumnUpdateDefault(), param)
	}
	if oldColumn.NotNull != column.NotNull {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetColumnUpdateNotNull(), param)
	}
	if column.Comment != "" {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetColumnComment(), param)
	}

	return
}
func (this_ *DialectService) ColumnUpdate(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, oldColumn *Column, column *Column) (err error) {
	sqlList, sqlArgsList := this_.ColumnUpdateSql(handler, databaseName, schemaName, tableName, oldColumn, column)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) ColumnDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string, columnName string) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ColumnParam(handler, databaseName, schemaName, tableName, &Column{Name: columnName})
	return this_.TemplateSqlList(this_.cfg.GetColumnDelete(), param)
}
func (this_ *DialectService) ColumnDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, columnName string) (err error) {
	sqlList, sqlArgsList := this_.ColumnDeleteSql(handler, databaseName, schemaName, tableName, columnName)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) ConstraintSelectSql(databaseName string, schemaName string, tableName string, constraintName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["databaseName"] = databaseName
	param["schemaName"] = schemaName
	param["tableName"] = tableName
	param["constraintName"] = constraintName
	return this_.TemplateSql(this_.cfg.GetConstraintSelect(), param)
}
func (this_ *DialectService) ConstraintList(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (constraintList []*Constraint, err error) {
	sqlInfo, sqlArgs := this_.ConstraintSelectSql(databaseName, schemaName, tableName, "")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	var columns []*ConstraintColumn
	for _, row := range list {
		columns = append(columns, ConstraintColumnByMap(row))
	}
	constraintList = ConstraintListByColumns(columns)
	return
}

func (this_ *DialectService) Constraint(sqlConn SqlConn, databaseName string, schemaName string, tableName string, constraintName string) (constraint *Constraint, err error) {
	sqlInfo, sqlArgs := this_.ConstraintSelectSql(databaseName, schemaName, tableName, constraintName)
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	var columns []*ConstraintColumn
	for _, row := range list {
		columns = append(columns, ConstraintColumnByMap(row))
	}
	constraintList := ConstraintListByColumns(columns)
	if len(constraintList) > 0 {
		constraint = constraintList[0]
	}
	return
}
func (this_ *DialectService) ConstraintAddSql(handler DDLHandler, databaseName string, schemaName string, tableName string, constraint *Constraint) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ConstraintParam(handler, databaseName, schemaName, tableName, constraint)
	sqlList, sqlArgsList = this_.TemplateSqlList(this_.cfg.GetConstraintAdd(), param)

	if constraint.Comment != "" {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetConstraintComment(), param)
	}
	return
}
func (this_ *DialectService) ConstraintAdd(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, constraint *Constraint) (err error) {
	sqlList, sqlArgsList := this_.ConstraintAddSql(handler, databaseName, schemaName, tableName, constraint)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) ConstraintCommentSql(handler DDLHandler, databaseName string, schemaName string, tableName string, constraint *Constraint) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ConstraintParam(handler, databaseName, schemaName, tableName, constraint)
	return this_.TemplateSqlList(this_.cfg.GetConstraintComment(), param)
}
func (this_ *DialectService) ConstraintComment(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, constraint *Constraint) (err error) {
	sqlList, sqlArgsList := this_.ConstraintCommentSql(handler, databaseName, schemaName, tableName, constraint)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) ConstraintDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string, constraintName string) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.ConstraintParam(handler, databaseName, schemaName, tableName, &Constraint{Name: constraintName})
	return this_.TemplateSqlList(this_.cfg.GetConstraintDelete(), param)
}
func (this_ *DialectService) ConstraintDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, constraintName string) (err error) {
	sqlList, sqlArgsList := this_.ConstraintDeleteSql(handler, databaseName, schemaName, tableName, constraintName)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) IndexSelectSql(databaseName string, schemaName string, tableName string, indexName string) (sqlInfo string, sqlArgs []any) {
	var param = map[string]any{}
	param["databaseName"] = databaseName
	param["schemaName"] = schemaName
	param["tableName"] = tableName
	param["indexName"] = indexName
	return this_.TemplateSql(this_.cfg.GetIndexSelect(), param)
}
func (this_ *DialectService) IndexList(sqlConn SqlConn, databaseName string, schemaName string, tableName string) (indexList []*Index, err error) {
	sqlInfo, sqlArgs := this_.IndexSelectSql(databaseName, schemaName, tableName, "")
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	var columns []*IndexColumn
	for _, row := range list {
		columns = append(columns, IndexColumnByMap(row))
	}
	indexList = IndexListByColumns(columns)
	return
}
func (this_ *DialectService) Index(sqlConn SqlConn, databaseName string, schemaName string, tableName string, indexName string) (index *Index, err error) {
	sqlInfo, sqlArgs := this_.IndexSelectSql(databaseName, schemaName, tableName, indexName)
	if len(sqlInfo) == 0 {
		return
	}
	list, err := this_.Query(sqlConn, sqlInfo, sqlArgs)
	if err != nil {
		return
	}
	var columns []*IndexColumn
	for _, row := range list {
		columns = append(columns, IndexColumnByMap(row))
	}
	indexList := IndexListByColumns(columns)
	if len(indexList) > 0 {
		index = indexList[0]
	}
	return
}
func (this_ *DialectService) IndexAddSql(handler DDLHandler, databaseName string, schemaName string, tableName string, index *Index) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.IndexParam(handler, databaseName, schemaName, tableName, index)
	sqlList, sqlArgsList = this_.TemplateSqlList(this_.cfg.GetIndexAdd(), param)
	if index.Comment != "" {
		this_.appendSql(&sqlList, &sqlArgsList, this_.cfg.GetIndexComment(), param)
	}
	return
}
func (this_ *DialectService) IndexAdd(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, index *Index) (err error) {
	sqlList, sqlArgsList := this_.IndexAddSql(handler, databaseName, schemaName, tableName, index)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) IndexCommentSql(handler DDLHandler, databaseName string, schemaName string, tableName string, index *Index) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.IndexParam(handler, databaseName, schemaName, tableName, index)
	return this_.TemplateSqlList(this_.cfg.GetIndexComment(), param)
}
func (this_ *DialectService) IndexComment(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, index *Index) (err error) {
	sqlList, sqlArgsList := this_.IndexCommentSql(handler, databaseName, schemaName, tableName, index)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}
func (this_ *DialectService) IndexDeleteSql(handler DDLHandler, databaseName string, schemaName string, tableName string, indexName string) (sqlList []string, sqlArgsList [][]any) {
	var param = this_.IndexParam(handler, databaseName, schemaName, tableName, &Index{Name: indexName})
	return this_.TemplateSqlList(this_.cfg.GetIndexDelete(), param)
}
func (this_ *DialectService) IndexDelete(sqlConn SqlConn, handler DDLHandler, databaseName string, schemaName string, tableName string, indexName string) (err error) {
	sqlList, sqlArgsList := this_.IndexDeleteSql(handler, databaseName, schemaName, tableName, indexName)
	return this_.Execs(sqlConn, sqlList, sqlArgsList)
}

func (this_ *DialectService) GetParam(handler DDLHandler, obj any) (param map[string]any) {
	param = map[string]any{}
	this_.AppendParam(handler, param, obj)

	return
}
func (this_ *DialectService) AppendParamTableName(handler DDLHandler, param map[string]any, databaseName string, schemaName string, tableName string) {
	delete(param, "databaseName")
	delete(param, "databaseNameWrap")
	if databaseName != "" && this_.GetDDLHandler(handler).IsAppendDatabaseName() {
		param["databaseName"] = databaseName
		param["databaseNameWrap"] = this_.GetDDLHandler(handler).WrapDatabaseName(databaseName)
	}
	delete(param, "schemaName")
	delete(param, "schemaNameWrap")
	if schemaName != "" && this_.GetDDLHandler(handler).IsAppendSchemaName() {
		param["schemaName"] = schemaName
		param["schemaNameWrap"] = this_.GetDDLHandler(handler).WrapDatabaseName(schemaName)
	}
	delete(param, "tableName")
	delete(param, "tableNameWrap")
	if tableName != "" {
		param["tableName"] = tableName
		param["tableNameWrap"] = this_.GetDDLHandler(handler).WrapTableName(tableName)
	}
	return
}
func (this_ *DialectService) AppendParam(handler DDLHandler, param map[string]any, obj any, ignoreNames ...string) {
	ignoreNameStr := "," + strings.ToLower(strings.Join(ignoreNames, ",")) + ","
	var extend map[string]any
	_, isDatabase := obj.(*Database)
	_, isUser := obj.(*User)
	_, isSchema := obj.(*Schema)
	_, isTable := obj.(*Table)
	_, isColumn := obj.(*Column)
	_, isConstraint := obj.(*Constraint)
	_, isIndex := obj.(*Index)
	_, isSequence := obj.(*Sequence)
	if obj != nil {
		o := reflect.ValueOf(obj)
		if o.Kind() == reflect.Ptr {
			o = o.Elem()
		}
		t := o.Type()
		fNum := t.NumField()
		for i := 0; i < fNum; i++ {
			fV := o.Field(i)
			fT := t.Field(i)
			var name = fT.Tag.Get("json")
			if strings.Contains(name, ",") {
				name = strings.Split(name, ",")[0]
			}
			if name == "" {
				name = strings.ToLower(fT.Name[0:1]) + fT.Name[1:]
			}
			if strings.Contains(ignoreNameStr, ","+strings.ToLower(name)+",") {
				continue
			}
			if !fV.CanInterface() {
				continue
			}

			//fmt.Println("append param name:"+name+", value:", fV.Interface())
			v := fV.Interface()

			switch vT := v.(type) {
			case string:
				param[name] = vT
				if vT != "" {
					var wrapName = vT
					if (strings.EqualFold(name, "name") && isUser) || (strings.EqualFold(name, "userName")) {
						wrapName = this_.GetDDLHandler(handler).WrapUserName(vT)
					} else if (strings.EqualFold(name, "name") && isDatabase) || (strings.EqualFold(name, "databaseName")) ||
						(strings.EqualFold(name, "referencedDatabaseName")) {
						wrapName = this_.GetDDLHandler(handler).WrapDatabaseName(vT)
					} else if (strings.EqualFold(name, "name") && isSchema) || (strings.EqualFold(name, "schemaName")) ||
						(strings.EqualFold(name, "referencedSchemaName")) {
						wrapName = this_.GetDDLHandler(handler).WrapSchemaName(vT)
					} else if (strings.EqualFold(name, "name") && isTable) || (strings.EqualFold(name, "tableName")) ||
						(strings.EqualFold(name, "referencedTableName")) {
						wrapName = this_.GetDDLHandler(handler).WrapTableName(vT)
					} else if (strings.EqualFold(name, "name") && isColumn) || (strings.EqualFold(name, "columnName")) ||
						(strings.EqualFold(name, "columnNames")) ||
						(strings.EqualFold(name, "columnNameList")) ||
						(strings.EqualFold(name, "referencedColumnName")) ||
						(strings.EqualFold(name, "referencedColumnNames")) ||
						(strings.EqualFold(name, "referencedColumnNameList")) {
						wrapName = this_.GetDDLHandler(handler).WrapColumnName(vT)
					} else if (strings.EqualFold(name, "name") && isConstraint) || (strings.EqualFold(name, "constraintName")) {
						wrapName = this_.GetDDLHandler(handler).WrapConstraintName(vT)
					} else if (strings.EqualFold(name, "name") && isIndex) || (strings.EqualFold(name, "indexName")) {
						wrapName = this_.GetDDLHandler(handler).WrapIndexName(vT)
					} else if (strings.EqualFold(name, "name") && isSequence) || (strings.EqualFold(name, "sequenceName")) ||
						(strings.EqualFold(name, "autoIncrementName")) {
						wrapName = this_.GetDDLHandler(handler).WrapSequenceName(vT)
					}

					param[name+"Wrap"] = wrapName
					param[name+"WrapString"] = this_.GetDDLHandler(handler).WrapString(vT)
				}
			case []string:
				param[name] = this_.Names(vT...)
				var warpNames = this_.Names(vT...)

				if (strings.EqualFold(name, "name") && isColumn) || (strings.EqualFold(name, "columnName")) ||
					(strings.EqualFold(name, "columnNames")) ||
					(strings.EqualFold(name, "columnNameList")) ||
					(strings.EqualFold(name, "referencedColumnName")) ||
					(strings.EqualFold(name, "referencedColumnNames")) ||
					(strings.EqualFold(name, "referencedColumnNameList")) {
					warpNames = this_.WrapNames(this_.GetDDLHandler(handler).WrapColumnName, vT...)
				}
				param[name+"Wrap"] = warpNames
			case map[string]any:
				extend = vT
			default:
				param[name] = v
			}
		}
	}
	this_.AppendParamExtend(handler, param, extend)

	return
}

func (this_ *DialectService) AppendParamExtend(handler DDLHandler, param map[string]any, extend map[string]any) {
	if len(extend) > 0 {
		for k, v := range extend {
			if v == nil {
				continue
			}
			param[k] = v
			if s, isS := v.(string); isS {
				param[k+"Wrap"] = this_.GetDDLHandler(handler).WrapName(s)
				param[k+"WrapString"] = this_.GetDDLHandler(handler).WrapString(s)
			}
		}
	}

	return
}
func (this_ *DialectService) TemplateSql(t *SqlTemplate, param map[string]any) (sqlInfo string, sqlArgs []any) {
	if t == nil {
		return
	}
	sqlInfo, sqlArgs = t.GetSql(nil, param)
	return
}

func (this_ *DialectService) TemplateSqlList(t *SqlTemplate, param map[string]any) (sqlList []string, sqlArgsList [][]any) {
	if t == nil {
		return
	}
	sqlList, sqlArgsList = t.GetSqlList(nil, param)
	return
}

func (this_ *DialectService) Query(sqlConn SqlConn, sqlInfo string, sqlArgs []any) (res []map[string]any, err error) {
	//fmt.Println("query sql:", sqlInfo)
	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	res, err = DoQuery(context.Background(), sqlConn, sqlInfo, sqlArgs, false, nil)
	if err != nil {
		err = errors.New("query sql:" + sqlInfo + " error:" + err.Error())
		return
	}
	return
}
func (this_ *DialectService) QueryOne(sqlConn SqlConn, sqlInfo string, sqlArgs []any) (res map[string]any, err error) {
	//fmt.Println("query one sql:", sqlInfo)
	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	res, err = DoQueryOne(context.Background(), sqlConn, sqlInfo, sqlArgs, false, nil)
	if err != nil {
		err = errors.New("query one sql:" + sqlInfo + " error:" + err.Error())
		return
	}
	return
}
func (this_ *DialectService) Exec(sqlConn SqlConn, sqlInfo string, sqlArgs []any) (err error) {
	//fmt.Println("exec sql:", sqlInfo)
	sqlInfo = this_.FormatSqlArgChar(sqlInfo)
	_, err = sqlConn.ExecContext(context.Background(), sqlInfo, sqlArgs...)
	if err != nil {
		err = errors.New("exec sql:" + sqlInfo + " error:" + err.Error())
		return
	}
	return
}
func (this_ *DialectService) Execs(sqlConn SqlConn, sqlList []string, sqlArgsList [][]any) (err error) {
	for i, sqlInfo := range sqlList {
		sqlArgs := sqlArgsList[i]
		//fmt.Println("exec sql:", sqlInfo)
		sqlInfo = this_.FormatSqlArgChar(sqlInfo)
		_, err = sqlConn.ExecContext(context.Background(), sqlInfo, sqlArgs...)
		if err != nil {
			err = errors.New("exec sql:" + sqlInfo + " error:" + err.Error())
			return
		}
	}
	return
}

func (this_ *DialectService) GetDDLHandler(handler DDLHandler) DDLHandler {
	if handler == nil {
		return this_
	}
	return handler
}

func (this_ *DialectService) FormatSqlArgChar(sqlInfo string) (res string) {
	var argChar = this_.ArgChar()
	strList := strings.Split(sqlInfo, "?")
	if len(strList) < 1 {
		res = sqlInfo
		return
	}
	res = strList[0]
	for i := 1; i < len(strList); i++ {
		if strings.EqualFold(argChar, "$num") || strings.EqualFold(argChar, "$name") {
			res += "$" + strconv.Itoa(i)
		} else if strings.EqualFold(argChar, ":num") || strings.EqualFold(argChar, ":name") {
			res += ":" + strconv.Itoa(i)
		} else if strings.EqualFold(argChar, "?num") || strings.EqualFold(argChar, "?name") {
			res += "?" + strconv.Itoa(i)
		} else {
			res += "?"
		}
		res += strList[i]
	}
	return
}

func (this_ *DialectService) FormatPageSql(sqlInfo string, pageSize int64, pageNo int64) (res string) {
	res = sqlInfo
	return
}

func (this_ *DialectService) WrapUserName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) IsAppendDatabaseName() bool {
	return false
}
func (this_ *DialectService) WrapDatabaseName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) IsAppendSchemaName() bool {
	return false
}
func (this_ *DialectService) WrapSchemaName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) WrapTableName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) WrapColumnName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) WrapConstraintName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) WrapIndexName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) WrapSequenceName(name string) string {
	return this_.WrapName(name)
}
func (this_ *DialectService) WrapName(name string) (res string) {
	var wrapChar = this_.WrapNameChar()
	return wrapChar + name + wrapChar
}
func (this_ *DialectService) Names(names ...string) (res string) {
	return strings.Join(names, ",")
}
func (this_ *DialectService) WrapNames(wrapName func(name string) (res string), names ...string) (res string) {
	var ss []string
	for _, name := range names {
		ss = append(ss, wrapName(name))
	}
	return strings.Join(ss, ",")
}

func (this_ *DialectService) WrapString(str string) (res string) {
	var wrapChar = this_.WrapStringChar()
	return wrapChar + str + wrapChar
}

func (this_ *DialectService) SqlConcat(args ...string) (res string) {
	c := this_.cfg.FindFunc("concat")
	if c == nil {
		c = map[string]any{
			"has": true,
		}
	}
	hasV, hasFind := c["has"]
	argSizeV, argSizeFind := c["argSize"]
	useV, useFind := c["use"]
	var has bool
	var argSize int
	var use string
	if hasFind && hasV != nil {
		has = IsTrue(hasV)
	}
	if argSizeFind && argSizeV != nil {
		argSize, _ = ToIntValue(argSizeV)
	}
	if useFind && useV != nil {
		use = GetStringValue(useV)
	}
	if has {
		if argSize == 2 {
			// CONCAT(arg1, arg2))
			// CONCAT(CONCAT(arg1, arg2), arg3)
			res = "CONCAT("
			var concatArgI int
			for _, one := range args {
				if concatArgI == 1 {
					res += `, `
				} else if concatArgI == 2 {
					res = "CONCAT(" + res + "), "
					concatArgI = 1
				}
				res += `"` + one + `"`
				concatArgI++
			}
			res += ")"
		} else {
			res = "CONCAT("
			for i, one := range args {
				if i > 0 {
					res += `, `
				}
				res += `'` + one + `'`
			}
			res += ")"
		}
	} else {
		for i, one := range args {
			if i > 0 {
				res += ` ` + use + ` `
			}
			res += `'` + one + `'`
		}
	}
	return
}
