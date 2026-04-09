package zookeeper

import (
	"github.com/go-zookeeper/zk"
	"net"
)

type Config struct {
	// Disabled 禁用 上层 初始化服务时候 可以判断该属性 如果为 配置 true 则不去初始化服务
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	Address           string `json:"address" yaml:"address"`
	Username          string `json:"username,omitempty" yaml:"username,omitempty"`
	Password          string `json:"password,omitempty" yaml:"password,omitempty"`
	SessionTimeout    int    `json:"sessionTimeout,omitempty" yaml:"sessionTimeout,omitempty"`       // 会话超时 单位 毫秒
	ConnectionTimeout int    `json:"connectionTimeout,omitempty" yaml:"connectionTimeout,omitempty"` // 客户端连接超时 单位 毫秒

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

// New 创建zookeeper客户端
func New(config *Config) (IService, error) {
	service := &Service{
		Config: config,
	}
	err := service.init(config.connProxy)
	if err != nil {
		return nil, err
	}
	return service, nil
}

type IService interface {
	// Close 关闭 客户端
	Close()
	// GetConn 获取 zk Conn
	GetConn() *zk.Conn
	// Info 查看 zk 相关信息
	Info() (info *Info, err error)
	// Create 创建 永久 节点
	Create(path string, value string) (err error)
	// CreateIfNotExists 如果不存在 则创建 永久 节点 创建时候如果已存在不报错  如果 父节点不存在 则先创建父节点
	CreateIfNotExists(path string, value string) (err error)
	// CreateEphemeral 创建 临时 节点
	CreateEphemeral(path string, value string) (err error)
	// CreateEphemeralIfNotExists 如果不存在 则创建 临时 节点 创建时候如果已存在不报错 如果 父节点不存在 则先创建父节点
	CreateEphemeralIfNotExists(path string, value string) (err error)
	// Exists 查看节点是否存在
	Exists(path string) (isExist bool, err error)
	// Set 设置 节点 值
	Set(path string, value string) (err error)
	// Get 查看 节点 数据
	Get(path string) (value string, err error)
	// GetInfo 查看 节点 信息
	GetInfo(path string) (info *NodeInfo, err error)
	// Stat 节点 状态
	Stat(path string) (info *StatInfo, err error)
	// GetChildren 查询 子节点
	GetChildren(path string) (children []string, err error)
	// Delete 删除节点 如果 包含子节点 则先删除所有子节点
	Delete(path string) (err error)
	// WatchChildren 监听 子节点 只监听当前节点下的子节点 NodeEventError 监听异常 NodeEventStopped zk客户端关闭 NodeEventAdded 节点新增 NodeEventDeleted 节点删除 NodeEventNodeNotFound 节点不存在
	WatchChildren(path string, listen func(data *WatchChildrenListenData) (finish bool)) (err error)
}
