package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	// Disabled 禁用 上层 初始化服务时候 可以判断该属性 如果为 配置 true 则不去初始化服务
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	Address        string `json:"address" yaml:"address"`
	Username       string `json:"username,omitempty" yaml:"username,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	MinPoolSize    int    `json:"minPoolSize,omitempty" yaml:"minPoolSize,omitempty"`
	MaxPoolSize    int    `json:"maxPoolSize,omitempty" yaml:"maxPoolSize,omitempty"`
	ConnectTimeout int    `json:"connectTimeout,omitempty" yaml:"connectTimeout,omitempty"` // 客户端连接超时 单位 毫秒
	CertPath       string `json:"certPath,omitempty" yaml:"certPath,omitempty"`
}

// New 创建 mongodb 客户端
func New(config *Config) (IService, error) {
	service := &Service{
		Config: config,
	}
	err := service.init()
	if err != nil {
		return nil, err
	}
	return service, nil
}

type IService interface {
	Close()
	Databases() (databases []*Database, totalSize int64, err error)
	DatabaseDelete(database string) (err error)
	Collections(database string) (collections []*Collection, err error)
	CollectionCreate(database string, collection string, opts ...*options.CreateCollectionOptions) (err error)
	CollectionDelete(database string, collection string) (err error)
	Indexes(database string, collection string) (indexes []map[string]interface{}, err error)
	IndexCreate(database string, collection string, index mongo.IndexModel) (name string, err error)
	IndexesCreate(database string, collection string, indexes []mongo.IndexModel) (names []string, err error)
	IndexDelete(database string, collection string, name string) (err error)
	IndexDeleteAll(database string, collection string) (err error)
	Insert(database string, collection string, document interface{}, opts ...*options.InsertOneOptions) (insertedId interface{}, err error)
	BatchInsert(database string, collection string, documents []interface{}, opts ...*options.InsertManyOptions) (insertedIds []interface{}, err error)
	Update(database string, collection string, id interface{}, update interface{}, opts ...*options.UpdateOptions) (updateResult *UpdateResult, err error)
	UpdateOne(database string, collection string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (updateResult *UpdateResult, err error)
	BatchUpdate(database string, collection string, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (updateResult *UpdateResult, err error)
	Count(database string, collection string, filter interface{}) (totalCount int64, err error)
	QueryMap(database string, collection string, filter interface{}, opts *options.FindOptions) (list []map[string]interface{}, err error)
	QueryMapPage(database string, collection string, filter interface{}, page *Page, opts *options.FindOptions) (list []map[string]interface{}, err error)
	QueryMapPageResult(database string, collection string, filter interface{}, page *Page, opts *options.FindOptions) (pageResult *Page, err error)
	DeleteOne(database string, collection string, filter interface{}) (deletedCount int64, err error)
	DeleteMany(database string, collection string, filter interface{}) (deletedCount int64, err error)
}
