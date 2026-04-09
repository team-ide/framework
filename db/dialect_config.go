package db

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	dialectConfigList []*DialectConfig
)

func GetDialectConfig(name string) *DialectConfig {
	for _, one := range dialectConfigList {
		if strings.EqualFold(one.Type, name) {
			return one
		}
	}
	for _, one := range dialectConfigList {
		for _, m := range one.Match {
			if strings.EqualFold(m, name) {
				return one
			}
		}
	}
	return nil
}

func AddDialectConfig(cfg *DialectConfig) *DialectConfig {
	dialectConfigList = append(dialectConfigList, cfg)
	return cfg
}

func LoadDialectConfigByFile(configFile string) (cfg *DialectConfig, err error) {
	bs, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		err = errors.New("read dialect config [" + configFile + "] error:" + err.Error())
		return
	}
	cfg, err = ToDialectConfigByBytes(configFile, bs)
	if err != nil {
		return
	}
	return
}

func ToDialectConfigByBytes(configFile string, bs []byte) (cfg *DialectConfig, err error) {
	cfg = &DialectConfig{}
	err = yaml.Unmarshal(bs, cfg)
	if err != nil {
		err = errors.New("yaml unmarshal dialect config [" + configFile + "] error:" + err.Error())
		return
	}
	if cfg.Type == "" {
		err = errors.New("dialect config [" + configFile + "] type is empty")
		return
	}
	return
}
func AddDialectByConfigFile(configFile string) (err error) {
	var cfg *DialectConfig
	cfg, err = LoadDialectConfigByFile(configFile)
	if err != nil {
		return
	}
	cfg.filePath = configFile
	AddDialectConfig(cfg)

	err = cfg.Init()

	if err != nil {
		err = errors.New("dialect config [" + cfg.filePath + "] init error:" + err.Error())
		return
	}

	dia := NewDialect(cfg)
	AddDialect(dia)
	return
}
func AddDialectByConfigDir(configDir string) (err error) {
	fs, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		err = errors.New("read dir [" + configDir + "] error:" + err.Error())
		return
	}
	var cfgList []*DialectConfig
	for _, one := range fs {
		configFile := path.Join(configDir, one.Name())
		if one.IsDir() {
			continue
		}
		var cfg *DialectConfig
		cfg, err = LoadDialectConfigByFile(configFile)
		if err != nil {
			return
		}
		cfg.filePath = configFile
		AddDialectConfig(cfg)
		cfgList = append(cfgList, cfg)
	}

	for _, cfg := range cfgList {
		err = cfg.Init()
		if err != nil {
			err = errors.New("dialect config [" + cfg.filePath + "] init error:" + err.Error())
			return
		}
	}
	for _, cfg := range cfgList {
		dia := NewDialect(cfg)
		AddDialect(dia)
	}
	return
}

func (this_ *DialectConfig) NewPlaceSqlTemplate(place string, sqlInfo string) (res *SqlTemplate, err error) {
	if sqlInfo == "" {
		return
	}
	res = NewSqlTemplate(sqlInfo)
	res.BracketOptional = true
	err = res.Parse()
	if err != nil {
		err = errors.New("dialect config [" + this_.Type + "] parse " + place + " sql error:" + err.Error())
		return
	}
	return
}
func (this_ *DialectConfig) Init() (err error) {
	if this_.Extend != "" {
		this_.extend = GetDialectConfig(this_.Extend)
		if this_.extend == nil {
			err = errors.New("dialect config [" + this_.Type + "] extend [" + this_.Extend + "] not found ")
			return
		}
	}
	if this_.User != nil {
		if this_.userChange, err = this_.NewPlaceSqlTemplate("user change", this_.User.Change); err != nil {
			return
		}
		if this_.userSelect, err = this_.NewPlaceSqlTemplate("user select", this_.User.Select); err != nil {
			return
		}
		if this_.userCreate, err = this_.NewPlaceSqlTemplate("user create", this_.User.Create); err != nil {
			return
		}
		if this_.userDelete, err = this_.NewPlaceSqlTemplate("user delete", this_.User.Delete); err != nil {
			return
		}

		this_.systemUsers = "," + strings.Join(this_.User.Systems, ",") + ","
		this_.systemUsers = strings.ToLower(this_.systemUsers)
	}
	if this_.Database != nil {
		if this_.databaseChange, err = this_.NewPlaceSqlTemplate("database change", this_.Database.Change); err != nil {
			return
		}
		if this_.databaseSelect, err = this_.NewPlaceSqlTemplate("database select", this_.Database.Select); err != nil {
			return
		}
		if this_.databaseCreate, err = this_.NewPlaceSqlTemplate("database create", this_.Database.Create); err != nil {
			return
		}
		if this_.databaseDelete, err = this_.NewPlaceSqlTemplate("database delete", this_.Database.Delete); err != nil {
			return
		}
		this_.systemDatabases = "," + strings.Join(this_.Database.Systems, ",") + ","
		this_.systemDatabases = strings.ToLower(this_.systemDatabases)
	}
	if this_.Schema != nil {
		if this_.schemaChange, err = this_.NewPlaceSqlTemplate("schema change", this_.Schema.Change); err != nil {
			return
		}
		if this_.schemaSelect, err = this_.NewPlaceSqlTemplate("schema select", this_.Schema.Select); err != nil {
			return
		}
		if this_.schemaCreate, err = this_.NewPlaceSqlTemplate("schema create", this_.Schema.Create); err != nil {
			return
		}
		if this_.schemaDelete, err = this_.NewPlaceSqlTemplate("schema delete", this_.Schema.Delete); err != nil {
			return
		}

		this_.systemSchemas = "," + strings.Join(this_.Schema.Systems, ",") + ","
		this_.systemSchemas = strings.ToLower(this_.systemSchemas)
	}
	if this_.Sequence != nil {
		if this_.sequenceSelect, err = this_.NewPlaceSqlTemplate("sequence select", this_.Sequence.Select); err != nil {
			return
		}
		if this_.sequenceCreate, err = this_.NewPlaceSqlTemplate("sequence create", this_.Sequence.Create); err != nil {
			return
		}
		if this_.sequenceDelete, err = this_.NewPlaceSqlTemplate("sequence delete", this_.Sequence.Delete); err != nil {
			return
		}
	}
	if this_.Table != nil {
		if this_.tableSelect, err = this_.NewPlaceSqlTemplate("table select", this_.Table.Select); err != nil {
			return
		}
		if this_.tableCreate, err = this_.NewPlaceSqlTemplate("table create", this_.Table.Create); err != nil {
			return
		}
		if this_.tableCreateComment, err = this_.NewPlaceSqlTemplate("table create comment", this_.Table.CreateComment); err != nil {
			return
		}
		if this_.tableCreateColumn, err = this_.NewPlaceSqlTemplate("table create column", this_.Table.CreateColumn); err != nil {
			return
		}
		if this_.tableCreateColumnComment, err = this_.NewPlaceSqlTemplate("table create column comment", this_.Table.CreateColumnComment); err != nil {
			return
		}
		if this_.tableCreateConstraint, err = this_.NewPlaceSqlTemplate("table create constraint", this_.Table.CreateConstraint); err != nil {
			return
		}
		if this_.tableCreateConstraintComment, err = this_.NewPlaceSqlTemplate("table create constraint comment", this_.Table.CreateConstraintComment); err != nil {
			return
		}
		if this_.tableCreateIndex, err = this_.NewPlaceSqlTemplate("table create index", this_.Table.CreateIndex); err != nil {
			return
		}
		if this_.tableCreateIndexComment, err = this_.NewPlaceSqlTemplate("table create index comment", this_.Table.CreateIndexComment); err != nil {
			return
		}
		if this_.tableComment, err = this_.NewPlaceSqlTemplate("table comment", this_.Table.Comment); err != nil {
			return
		}
		if this_.tableCreateSequenceTrigger, err = this_.NewPlaceSqlTemplate("table create sequence trigger", this_.Table.CreateSequenceTrigger); err != nil {
			return
		}
		if this_.tableDelete, err = this_.NewPlaceSqlTemplate("table delete", this_.Table.Delete); err != nil {
			return
		}
	}
	if this_.Column != nil {
		if this_.columnSelect, err = this_.NewPlaceSqlTemplate("column select", this_.Column.Select); err != nil {
			return
		}
		if this_.columnAdd, err = this_.NewPlaceSqlTemplate("column add", this_.Column.Add); err != nil {
			return
		}
		if this_.columnComment, err = this_.NewPlaceSqlTemplate("column comment", this_.Column.Comment); err != nil {
			return
		}
		if this_.columnUpdate, err = this_.NewPlaceSqlTemplate("column update", this_.Column.Update); err != nil {
			return
		}
		if this_.columnUpdateRename, err = this_.NewPlaceSqlTemplate("column update rename", this_.Column.UpdateRename); err != nil {
			return
		}
		if this_.columnUpdateType, err = this_.NewPlaceSqlTemplate("column update type", this_.Column.UpdateType); err != nil {
			return
		}
		if this_.columnUpdateDefault, err = this_.NewPlaceSqlTemplate("column update default", this_.Column.UpdateDefault); err != nil {
			return
		}
		if this_.columnUpdateNotNull, err = this_.NewPlaceSqlTemplate("column update not null", this_.Column.UpdateNotNull); err != nil {
			return
		}
		if this_.columnDelete, err = this_.NewPlaceSqlTemplate("column delete", this_.Column.Delete); err != nil {
			return
		}
	}
	if this_.Constraint != nil {
		if this_.constraintSelect, err = this_.NewPlaceSqlTemplate("constraint select", this_.Constraint.Select); err != nil {
			return
		}
		if this_.constraintAdd, err = this_.NewPlaceSqlTemplate("constraint add", this_.Constraint.Add); err != nil {
			return
		}
		if this_.constraintComment, err = this_.NewPlaceSqlTemplate("constraint comment", this_.Constraint.Comment); err != nil {
			return
		}
		if this_.constraintDelete, err = this_.NewPlaceSqlTemplate("constraint delete", this_.Constraint.Delete); err != nil {
			return
		}
	}
	if this_.Index != nil {
		if this_.indexSelect, err = this_.NewPlaceSqlTemplate("index select", this_.Index.Select); err != nil {
			return
		}
		if this_.indexAdd, err = this_.NewPlaceSqlTemplate("index add", this_.Index.Add); err != nil {
			return
		}
		if this_.indexComment, err = this_.NewPlaceSqlTemplate("index comment", this_.Index.Comment); err != nil {
			return
		}
		if this_.indexDelete, err = this_.NewPlaceSqlTemplate("index delete", this_.Index.Delete); err != nil {
			return
		}
	}
	this_.funcCache = make(map[string]map[string]any)

	for name, one := range this_.FuncList {
		this_.funcCache[strings.ToLower(name)] = one
	}

	this_.typeCache = make(map[string]*DialectTypeConfig)
	for _, one := range this_.TypeList {
		var format = one.Name
		var name = one.Name
		index := strings.Index(one.Name, "(")
		var args []string
		if index > 0 {
			name = strings.TrimSpace(one.Name[:index])
			endS := strings.TrimSuffix(one.Name[index+1:], ")")
			ss := strings.Split(endS, ",")
			if len(ss) == 1 {
				args = append(args, "$l")
			} else if len(ss) == 2 {
				args = append(args, "$p")
				args = append(args, "$s")
			}
			//for i, s := range ss {
			//	switch strings.TrimSpace(s) {
			//	case "长度", "$L":
			//		s = "$l"
			//	case "精度", "$P":
			//		s = "$p"
			//	case "标度", "$S":
			//		s = "$s"
			//	}
			//	args = append(args, s)
			//}
			format = name + "(" + strings.Join(args, ",") + ")"
		}
		key := strings.ToLower(name)
		if this_.typeCache[name] != nil {
			err = errors.New("dialect config [" + this_.Type + "] type [" + one.Name + "] already exist")
			return
		}
		one.Name = name
		one.Format = format
		one.Args = args
		this_.typeCache[key] = one
	}
	return
}

type DialectConfig struct {
	extend *DialectConfig

	userChange *SqlTemplate
	userSelect *SqlTemplate
	userCreate *SqlTemplate
	userDelete *SqlTemplate

	databaseChange *SqlTemplate
	databaseSelect *SqlTemplate
	databaseCreate *SqlTemplate
	databaseDelete *SqlTemplate

	schemaChange *SqlTemplate
	schemaSelect *SqlTemplate
	schemaCreate *SqlTemplate
	schemaDelete *SqlTemplate

	sequenceSelect *SqlTemplate
	sequenceCreate *SqlTemplate
	sequenceDelete *SqlTemplate

	tableSelect                  *SqlTemplate
	tableCreate                  *SqlTemplate
	tableCreateComment           *SqlTemplate
	tableCreateSequenceTrigger   *SqlTemplate
	tableCreateColumn            *SqlTemplate
	tableCreateColumnComment     *SqlTemplate
	tableCreateConstraint        *SqlTemplate
	tableCreateConstraintComment *SqlTemplate
	tableCreateIndex             *SqlTemplate
	tableCreateIndexComment      *SqlTemplate
	tableComment                 *SqlTemplate
	tableDelete                  *SqlTemplate

	columnSelect        *SqlTemplate
	columnAdd           *SqlTemplate
	columnComment       *SqlTemplate
	columnUpdate        *SqlTemplate
	columnUpdateRename  *SqlTemplate
	columnUpdateType    *SqlTemplate
	columnUpdateDefault *SqlTemplate
	columnUpdateNotNull *SqlTemplate
	columnDelete        *SqlTemplate

	constraintSelect  *SqlTemplate
	constraintAdd     *SqlTemplate
	constraintComment *SqlTemplate
	constraintDelete  *SqlTemplate

	indexSelect  *SqlTemplate
	indexAdd     *SqlTemplate
	indexComment *SqlTemplate
	indexDelete  *SqlTemplate

	filePath string

	typeCache map[string]*DialectTypeConfig

	funcCache map[string]map[string]any

	systemUsers     string
	systemDatabases string
	systemSchemas   string

	// 方言类型
	Type string `yaml:"type,omitempty"`
	// 查找方言 传入的类型 匹配 type 或 以下配置的字符 则使用该方言
	Match []string `yaml:"match,omitempty"`
	// 使用 其它方言作为基础数据，可以配置覆盖
	Extend string `yaml:"extend,omitempty"`
	// 驱动
	Driver *DialectDriverConfig `yaml:"driver,omitempty"`
	// 名称包裹字符 如：`name_1`、"name_1"
	NameWrapChar string `yaml:"nameWrapChar,omitempty"`
	// 名称 没有包装 则自动转大写
	NameNoWrapIsUpper bool `yaml:"nameNoWrapIsUpper,omitempty"`
	// 字符串包裹字符 如：'string_1'
	StringWrapChar string `yaml:"stringWrapChar,omitempty"`
	/// 字符串转义字符 如：'xx\'x'
	StringEscapeChar string `yaml:"stringEscapeChar,omitempty"`
	// 变量占位符 如：?、$num、:name、?num
	ArgChar string `yaml:"argChar,omitempty"`
	// 结构关系 如：database = schema
	Relation string `yaml:"relation,omitempty"`
	// 表单
	FormList []*Form `yaml:"formList,omitempty"`

	// 树形配置
	Tree []*DialectNodeConfig `yaml:"tree,omitempty"`

	// 用户
	User *DialectUserConfig `yaml:"user,omitempty"`
	// 库
	Database *DialectDatabaseConfig `yaml:"database,omitempty"`
	// 模式
	Schema *DialectSchemaConfig `yaml:"schema,omitempty"`
	// 序列
	Sequence *DialectSequenceConfig `yaml:"sequence,omitempty"`
	// 表
	Table *DialectTableConfig `yaml:"table,omitempty"`
	// 字段
	Column *DialectColumnConfig `yaml:"column,omitempty"`
	// 约束
	Constraint *DialectConstraintConfig `yaml:"constraint,omitempty"`
	// 索引
	Index *DialectIndexConfig `yaml:"index,omitempty"`

	TypeExtend bool                 `yaml:"typeExtend,omitempty"`
	TypeList   []*DialectTypeConfig `yaml:"typeList,omitempty"`

	FuncExtend bool                      `yaml:"funcExtend,omitempty"`
	FuncList   map[string]map[string]any `yaml:"funcList,omitempty"`
}

type DialectDriverConfig struct {
	Name  string `yaml:"name,omitempty"`
	Dsn   string `yaml:"dsn,omitempty"`
	IsUrl bool   `yaml:"isUrl,omitempty"`

	// 驱动连接中 是否可以包含 数据库
	HasDatabase bool `yaml:"hasDatabase,omitempty"`
	// 驱动连接中 是否可以包含 模式
	HasSchema bool `yaml:"hasSchema,omitempty"`

	// 值是否可以 编码 用于特殊字符
	CanPathEscape bool `yaml:"canPathEscape,omitempty"`

	// 参数间隔 字符 非 url 使用，默认为 " "
	ParamSpaceChar string `yaml:"paramSpaceChar,omitempty"`

	Params map[string]string `yaml:"params,omitempty"`
}

func (this_ *DialectConfig) IsSystemUser(name string) bool {
	if len(this_.systemUsers) == 0 && this_.extend != nil {
		return this_.extend.IsSystemUser(name)
	}
	return strings.Contains(this_.systemUsers, ","+strings.ToLower(name)+",")
}
func (this_ *DialectConfig) IsSystemDatabase(name string) bool {
	if len(this_.systemDatabases) == 0 && this_.extend != nil {
		return this_.extend.IsSystemDatabase(name)
	}
	return strings.Contains(this_.systemDatabases, ","+strings.ToLower(name)+",")
}
func (this_ *DialectConfig) IsSystemSchema(name string) bool {
	if len(this_.systemSchemas) == 0 && this_.extend != nil {
		return this_.extend.IsSystemSchema(name)
	}
	return strings.Contains(this_.systemSchemas, ","+strings.ToLower(name)+",")
}
func (this_ *DialectConfig) GetDriver() *DialectDriverConfig {
	if this_.Driver != nil && this_.Driver.Name != "" {
		return this_.Driver
	}
	if this_.extend != nil {
		return this_.extend.GetDriver()
	}
	return nil
}
func (this_ *DialectConfig) GetNameWrapChar() string {
	if this_.NameWrapChar == "" && this_.extend != nil {
		return this_.extend.GetNameWrapChar()
	}
	return this_.NameWrapChar
}
func (this_ *DialectConfig) GetNameNoWrapIsUpper() bool {
	return this_.NameNoWrapIsUpper
}
func (this_ *DialectConfig) GetStringWrapChar() string {
	if this_.StringWrapChar == "" && this_.extend != nil {
		return this_.extend.GetStringWrapChar()
	}
	return this_.StringWrapChar
}
func (this_ *DialectConfig) GetStringEscapeChar() string {
	if this_.StringEscapeChar == "" && this_.extend != nil {
		return this_.extend.GetStringEscapeChar()
	}
	return this_.StringEscapeChar
}
func (this_ *DialectConfig) GetArgChar() string {
	if this_.ArgChar == "" && this_.extend != nil {
		return this_.extend.GetArgChar()
	}
	return this_.ArgChar
}

type DialectNodeConfig struct {
	Name    string `yaml:"name,omitempty"`
	Comment string `yaml:"comment,omitempty"`
	Rule    string `yaml:"rule,omitempty"`

	// 需要切换用户
	ShouldChangeUser bool `yaml:"shouldChangeUser,omitempty"`
	// 需要切换数据库
	ShouldChangeDatabase bool `yaml:"shouldChangeDatabase,omitempty"`
	// 需要切换模式
	ShouldChangeSchema bool `yaml:"shouldChangeSchema,omitempty"`

	Children []*DialectNodeConfig `yaml:"children,omitempty"`
}

type DialectUserConfig struct {
	// 创建用户 会自动创建 数据库
	AutoCreateDatabase bool `yaml:"autoCreateDatabase,omitempty"`
	// 创建用户 会自动创建 模式
	AutoCreateSchema bool `yaml:"autoCreateSchema,omitempty"`

	Change string `yaml:"change,omitempty"`
	// 查询 SQL 字段：userName；参数：userName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：userName、password、加上表单指定
	Create string `yaml:"create,omitempty"`
	// 字段
	CreateFieldList []*Field `yaml:"createFieldList,omitempty"`
	// 删除 SQL 参数：userName
	Delete string `yaml:"delete,omitempty"`

	Systems []string `yaml:"systems,omitempty"`
}

func (this_ *DialectConfig) CreateUserAutoCreateDatabase() bool {
	if this_.User != nil {
		return this_.User.AutoCreateDatabase
	}
	if this_.extend != nil {
		return this_.extend.CreateUserAutoCreateDatabase()
	}
	return false
}
func (this_ *DialectConfig) CreateUserAutoCreateSchema() bool {
	if this_.User != nil {
		return this_.User.AutoCreateSchema
	}
	if this_.extend != nil {
		return this_.extend.CreateUserAutoCreateSchema()
	}
	return false
}
func (this_ *DialectConfig) GetUserChange() (t *SqlTemplate) {
	t = this_.userChange
	if t == nil && this_.extend != nil {
		return this_.extend.GetUserChange()
	}
	return
}
func (this_ *DialectConfig) GetUserSelect() (t *SqlTemplate) {
	t = this_.userSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetUserSelect()
	}
	return
}
func (this_ *DialectConfig) GetUserCreate() (t *SqlTemplate) {
	t = this_.userCreate
	if t == nil && this_.extend != nil {
		return this_.extend.GetUserSelect()
	}
	return
}
func (this_ *DialectConfig) GetUserCreateFieldList() (res []*Field) {
	if this_.User != nil {
		res = this_.User.CreateFieldList
	}
	if res == nil && this_.extend != nil {
		return this_.extend.GetUserCreateFieldList()
	}
	return
}
func (this_ *DialectConfig) GetUserDelete() (t *SqlTemplate) {
	t = this_.userDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetUserDelete()
	}
	return
}

type DialectDatabaseConfig struct {
	Change string `yaml:"change,omitempty"`
	// 查询 SQL 字段：databaseName；参数：databaseName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：databaseName、加上表单字段
	Create string `yaml:"create,omitempty"`
	// 创建 SQL 附件 字段
	CreateFieldList []*Field `yaml:"createFieldList,omitempty"`
	// 删除 SQL 参数：databaseName
	Delete string `yaml:"delete,omitempty"`

	// 需要切换用户
	ShouldChangeUser bool `yaml:"shouldChangeUser,omitempty"`
	// 需要切换数据库
	ShouldChangeDatabase bool `yaml:"shouldChangeDatabase,omitempty"`
	// 需要切换模式
	ShouldChangeSchema bool `yaml:"shouldChangeSchema,omitempty"`

	Systems []string `yaml:"systems,omitempty"`
}

func (this_ *DialectConfig) GetDatabaseChange() (t *SqlTemplate) {
	t = this_.databaseChange
	if t == nil && this_.extend != nil {
		return this_.extend.GetDatabaseChange()
	}
	return
}
func (this_ *DialectConfig) GetDatabaseSelect() (t *SqlTemplate) {
	t = this_.databaseSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetDatabaseSelect()
	}
	return
}
func (this_ *DialectConfig) GetDatabaseCreate() (t *SqlTemplate) {
	t = this_.databaseCreate
	if t == nil && this_.extend != nil {
		return this_.extend.GetDatabaseCreate()
	}
	return
}
func (this_ *DialectConfig) GetDatabaseCreateFieldList() (res []*Field) {
	if this_.Database != nil {
		res = this_.Database.CreateFieldList
	}
	if res == nil && this_.extend != nil {
		return this_.extend.GetDatabaseCreateFieldList()
	}
	return
}
func (this_ *DialectConfig) GetDatabaseDelete() (t *SqlTemplate) {
	t = this_.databaseDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetDatabaseDelete()
	}
	return
}

type DialectSchemaConfig struct {
	Change string `yaml:"change,omitempty"`
	// 查询 SQL 字段：databaseName、schemaName；参数：databaseName、schemaName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：databaseName、schemaName、加上表单字段
	Create string `yaml:"create,omitempty"`
	// 创建 SQL 附件 字段
	CreateFieldList []*Field `yaml:"createFieldList,omitempty"`
	// 删除 SQL 参数：databaseName、schemaName
	Delete string `yaml:"delete,omitempty"`

	// 需要切换用户
	ShouldChangeUser bool `yaml:"shouldChangeUser,omitempty"`
	// 需要切换数据库
	ShouldChangeDatabase bool `yaml:"shouldChangeDatabase,omitempty"`
	// 需要切换模式
	ShouldChangeSchema bool `yaml:"shouldChangeSchema,omitempty"`

	Systems []string `yaml:"systems,omitempty"`
}

func (this_ *DialectConfig) GetSchemaChange() (t *SqlTemplate) {
	t = this_.schemaChange
	if t == nil && this_.extend != nil {
		return this_.extend.GetSchemaChange()
	}
	return
}
func (this_ *DialectConfig) GetSchemaSelect() (t *SqlTemplate) {
	t = this_.schemaSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetSchemaSelect()
	}
	return
}
func (this_ *DialectConfig) GetSchemaCreate() (t *SqlTemplate) {
	t = this_.schemaCreate
	if t == nil && this_.extend != nil {
		return this_.extend.GetSchemaCreate()
	}
	return
}
func (this_ *DialectConfig) GetSchemaCreateFieldList() (res []*Field) {
	if this_.Schema != nil {
		res = this_.Schema.CreateFieldList
	}
	if res == nil && this_.extend != nil {
		return this_.extend.GetSchemaCreateFieldList()
	}
	return
}
func (this_ *DialectConfig) GetSchemaDelete() (t *SqlTemplate) {
	t = this_.schemaDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetSchemaDelete()
	}
	return
}

type DialectSequenceConfig struct {
	// 查询 SQL 字段：databaseName、schemaName；参数：databaseName、schemaName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：databaseName、schemaName、加上表单字段
	Create string `yaml:"create,omitempty"`
	// 创建 SQL 附件 字段
	CreateFieldList []*Field `yaml:"createFieldList,omitempty"`
	// 删除 SQL 参数：databaseName、schemaName
	Delete string `yaml:"delete,omitempty"`
}

func (this_ *DialectConfig) GetSequenceSelect() (t *SqlTemplate) {
	t = this_.sequenceSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetSequenceSelect()
	}
	return
}
func (this_ *DialectConfig) GetSequenceCreate() (t *SqlTemplate) {
	t = this_.sequenceCreate
	if t == nil && this_.extend != nil {
		return this_.extend.GetSequenceCreate()
	}
	return
}
func (this_ *DialectConfig) GetSequenceCreateFieldList() (res []*Field) {
	if this_.Sequence != nil {
		res = this_.Sequence.CreateFieldList
	}
	if res == nil && this_.extend != nil {
		return this_.extend.GetSequenceCreateFieldList()
	}
	return
}
func (this_ *DialectConfig) GetSequenceDelete() (t *SqlTemplate) {
	t = this_.sequenceDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetSequenceDelete()
	}
	return
}

type DialectTableConfig struct {
	// 查询 SQL 字段：databaseName、schemaName、tableName；参数：databaseName、schemaName、tableName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：databaseName、schemaName、tableName、加上表单字段
	Create        string `yaml:"create,omitempty"`
	CreateComment string `yaml:"createComment,omitempty"`
	// 创建 SQL 附件 字段
	CreateFieldList []*Field `yaml:"createFieldList,omitempty"`
	// 建表 字段 SQL 参数：columnName、columnType、notNull、
	CreateColumn        string `yaml:"createColumn,omitempty"`
	CreateColumnComment string `yaml:"createColumnComment,omitempty"`
	// 建表 约束 SQL 参数：columnName、columnType、notNull、
	CreateConstraint        string `yaml:"createConstraint,omitempty"`
	CreateConstraintComment string `yaml:"createConstraintComment,omitempty"`

	CreateSequenceTrigger string `yaml:"createSequenceTrigger,omitempty"`

	// 建表 索引 SQL 参数：columnName、columnType、notNull、
	CreateIndex        string `yaml:"createIndex,omitempty"`
	CreateIndexComment string `yaml:"createIndexComment,omitempty"`
	Comment            string `yaml:"comment,omitempty"`
	// 删除 SQL 参数：databaseName、schemaName、tableName
	Delete string `yaml:"delete,omitempty"`

	// 需要切换用户
	ShouldChangeUser bool `yaml:"shouldChangeUser,omitempty"`
	// 需要切换数据库
	ShouldChangeDatabase bool `yaml:"shouldChangeDatabase,omitempty"`
	// 需要切换模式
	ShouldChangeSchema bool `yaml:"shouldChangeSchema,omitempty"`
}

func (this_ *DialectConfig) GetTableSelect() (t *SqlTemplate) {
	t = this_.tableSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableSelect()
	}
	return
}
func (this_ *DialectConfig) GetTableCreate() (t *SqlTemplate) {
	t = this_.tableCreate
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreate()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateSequenceTrigger() (t *SqlTemplate) {
	t = this_.tableCreateSequenceTrigger
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateSequenceTrigger()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateComment() (t *SqlTemplate) {
	t = this_.tableCreateComment
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateComment()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateColumn() (t *SqlTemplate) {
	t = this_.tableCreateColumn
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateColumn()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateColumnComment() (t *SqlTemplate) {
	t = this_.tableCreateColumnComment
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateColumnComment()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateConstraint() (t *SqlTemplate) {
	t = this_.tableCreateConstraint
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateConstraint()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateConstraintComment() (t *SqlTemplate) {
	t = this_.tableCreateConstraintComment
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateConstraintComment()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateIndex() (t *SqlTemplate) {
	t = this_.tableCreateIndex
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateIndex()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateIndexComment() (t *SqlTemplate) {
	t = this_.tableCreateIndexComment
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableCreateIndexComment()
	}
	return
}
func (this_ *DialectConfig) GetTableCreateFieldList() (res []*Field) {
	if this_.Table != nil {
		res = this_.Table.CreateFieldList
	}
	if res == nil && this_.extend != nil {
		return this_.extend.GetTableCreateFieldList()
	}
	return
}
func (this_ *DialectConfig) GetTableDelete() (t *SqlTemplate) {
	t = this_.tableDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetTableDelete()
	}
	return
}

type DialectColumnConfig struct {
	// 查询 SQL 字段：databaseName、schemaName、tableName；参数：databaseName、schemaName、tableName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：databaseName、schemaName、tableName、加上表单字段
	Add           string `yaml:"add,omitempty"`
	Comment       string `yaml:"comment,omitempty"`
	Update        string `yaml:"update,omitempty"`
	UpdateRename  string `yaml:"updateRename,omitempty"`
	UpdateType    string `yaml:"updateType,omitempty"`
	UpdateNotNull string `yaml:"updateNotNull,omitempty"`
	UpdateDefault string `yaml:"updateDefault,omitempty"`
	// 删除 SQL 参数：databaseName、schemaName、tableName
	Delete string `yaml:"delete,omitempty"`
}

func (this_ *DialectConfig) GetColumnSelect() (t *SqlTemplate) {
	t = this_.columnSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnSelect()
	}
	return
}
func (this_ *DialectConfig) GetColumnAdd() (t *SqlTemplate) {
	t = this_.columnAdd
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnAdd()
	}
	return
}
func (this_ *DialectConfig) GetColumnComment() (t *SqlTemplate) {
	t = this_.columnComment
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnComment()
	}
	return
}
func (this_ *DialectConfig) GetColumnUpdate() (t *SqlTemplate) {
	t = this_.columnUpdate
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnUpdate()
	}
	return
}
func (this_ *DialectConfig) GetColumnUpdateRename() (t *SqlTemplate) {
	t = this_.columnUpdateRename
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnUpdateRename()
	}
	return
}
func (this_ *DialectConfig) GetColumnUpdateType() (t *SqlTemplate) {
	t = this_.columnUpdateType
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnUpdateType()
	}
	return
}
func (this_ *DialectConfig) GetColumnUpdateDefault() (t *SqlTemplate) {
	t = this_.columnUpdateDefault
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnUpdateDefault()
	}
	return
}
func (this_ *DialectConfig) GetColumnUpdateNotNull() (t *SqlTemplate) {
	t = this_.columnUpdateNotNull
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnUpdateNotNull()
	}
	return
}
func (this_ *DialectConfig) GetColumnDelete() (t *SqlTemplate) {
	t = this_.columnDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetColumnDelete()
	}
	return
}

type DialectConstraintConfig struct {
	// 查询 SQL 字段：databaseName、schemaName、tableName；参数：databaseName、schemaName、tableName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：databaseName、schemaName、tableName、加上表单字段
	Add     string `yaml:"add,omitempty"`
	Comment string `yaml:"comment,omitempty"`
	// 删除 SQL 参数：databaseName、schemaName、tableName
	Delete string `yaml:"delete,omitempty"`
}

func (this_ *DialectConfig) GetConstraintSelect() (t *SqlTemplate) {
	t = this_.constraintSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetConstraintSelect()
	}
	return
}
func (this_ *DialectConfig) GetConstraintAdd() (t *SqlTemplate) {
	t = this_.constraintAdd
	if t == nil && this_.extend != nil {
		return this_.extend.GetConstraintAdd()
	}
	return
}
func (this_ *DialectConfig) GetConstraintComment() (t *SqlTemplate) {
	t = this_.constraintComment
	if t == nil && this_.extend != nil {
		return this_.extend.GetConstraintComment()
	}
	return
}
func (this_ *DialectConfig) GetConstraintDelete() (t *SqlTemplate) {
	t = this_.constraintDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetConstraintDelete()
	}
	return
}

type DialectIndexConfig struct {
	// 查询 SQL 字段：databaseName、schemaName、tableName；参数：databaseName、schemaName、tableName
	Select string `yaml:"select,omitempty"`
	// 创建 SQL 参数：databaseName、schemaName、tableName、加上表单字段
	Add     string `yaml:"add,omitempty"`
	Comment string `yaml:"comment,omitempty"`
	// 删除 SQL 参数：databaseName、schemaName、tableName
	Delete string `yaml:"delete,omitempty"`
}

func (this_ *DialectConfig) GetIndexSelect() (t *SqlTemplate) {
	t = this_.indexSelect
	if t == nil && this_.extend != nil {
		return this_.extend.GetIndexSelect()
	}
	return
}
func (this_ *DialectConfig) GetIndexAdd() (t *SqlTemplate) {
	t = this_.indexAdd
	if t == nil && this_.extend != nil {
		return this_.extend.GetIndexAdd()
	}
	return
}
func (this_ *DialectConfig) GetIndexComment() (t *SqlTemplate) {
	t = this_.indexComment
	if t == nil && this_.extend != nil {
		return this_.extend.GetIndexComment()
	}
	return
}
func (this_ *DialectConfig) GetIndexDelete() (t *SqlTemplate) {
	t = this_.indexDelete
	if t == nil && this_.extend != nil {
		return this_.extend.GetIndexDelete()
	}
	return
}

func (this_ *DialectConfig) FindFunc(name string) (r map[string]any) {
	if this_.funcCache != nil {
		r = this_.funcCache[strings.ToLower(name)]
	}
	if len(this_.FuncList) > 0 && !this_.FuncExtend {
		return
	}
	if r == nil && this_.extend != nil {
		return this_.extend.FindFunc(name)
	}
	return
}

type DialectTypeConfig struct {
	Name   string   `yaml:"name,omitempty"`
	Format string   `yaml:"format,omitempty"`
	Args   []string `yaml:"args,omitempty"`
	Match  []string `yaml:"match,omitempty"`
}

func (this_ *DialectConfig) FindType(name string) (r *DialectTypeConfig) {
	if this_.typeCache != nil {
		r = this_.typeCache[strings.ToLower(name)]
	}
	if len(this_.TypeList) > 0 && !this_.TypeExtend {
		return
	}
	if r == nil && this_.extend != nil {
		return this_.extend.FindType(name)
	}
	return
}

func (this_ *DialectConfig) MatchType(dataType string, length, precision, scale int) (r *DialectTypeConfig) {
	r = this_.FindType(dataType)
	if r != nil {
		return
	}
	for _, t := range this_.TypeList {
		var match bool
		for _, m := range t.Match {
			var name = m
			var binaryI = strings.Index(m, "&&")
			var binary string
			if binaryI > 0 {
				name = strings.TrimSpace(m[0:binaryI])
				binary = strings.TrimSpace(m[binaryI+2:])
			}
			if !strings.EqualFold(dataType, strings.TrimSpace(name)) {
				continue
			}
			if binary == "" {
				match = true
				break
			}
		}
		if match {
			r = t
		}
	}
	if r == nil && this_.extend != nil {
		return this_.extend.MatchType(dataType, length, precision, scale)
	}
	return
}
func (this_ *DialectTypeConfig) FormatColumnType(length, precision, scale int) (columnType string) {
	columnType = this_.Name
	var args []string
	for _, arg := range this_.Args {
		switch arg {
		case "$l":
			if length == 0 && precision > 0 {
				length = precision
			}
			if length > 0 {
				args = append(args, strconv.Itoa(length))
			}
		case "$p":
			if precision == 0 && length > 0 {
				precision = length
			}
			if precision > 0 {
				args = append(args, strconv.Itoa(precision))
			}
		case "$s":
			if scale > 0 {
				args = append(args, strconv.Itoa(scale))
			}
		}
	}

	if len(args) > 0 {
		columnType += "(" + strings.Join(args, ",") + ")"
	}
	return
}

type Form struct {
	// 表单名称
	Name string `yaml:"name,omitempty"`
	// 字段
	FieldList []*Field `yaml:"fieldList,omitempty"`
}

type Field struct {
	// 字段名称
	Name string `yaml:"name,omitempty"`
	// 标签
	Label string `yaml:"label,omitempty"`
	// 类型 text、password、select、redis、checkbox、json字符串，form
	Type string `yaml:"type,omitempty"`
	// 多个值
	List bool `yaml:"list,omitempty"`
	// 必填
	Required bool `yaml:"required,omitempty"`
	// 子表单
	FormName string `yaml:"formName,omitempty"`
	// 默认值
	Default   string `json:"default,omitempty"`
	IsInteger bool   `json:"isInteger,omitempty"`
}
