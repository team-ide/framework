package httpx

import (
	"io"
	"net"
	"net/http"
)

type Config struct {
	// Disabled 禁用 上层 初始化服务时候 可以判断该属性 如果为 配置 true 则不去初始化服务
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	RootUrl string `json:"rootUrl" yaml:"rootUrl"`

	// 超时时间 单位 毫秒
	Timeout         int `json:"timeout" yaml:"timeout"`
	MaxIdleConns    int `json:"maxIdleConns" yaml:"maxIdleConns"`
	MaxConnsPerHost int `json:"maxConnsPerHost" yaml:"maxConnsPerHost"`
	// 超时时间 单位 毫秒
	IdleConnTimeout int `json:"idleConnTimeout" yaml:"idleConnTimeout"`

	Tls *TLSConfig `json:"tls,omitempty" yaml:"tls,omitempty"`

	connProxy ConnProxy
}

type TLSConfig struct {
	InsecureSkipVerify bool `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
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
	// Info 查看 zk 相关信息
	Info() (info *Info, err error)

	Request(method, url string, body io.Reader, sets ...Set) (resp *http.Response, err error)
	GetRequest(url string, sets ...Set) (resp *http.Response, err error)
	PostRequest(url string, body io.Reader, sets ...Set) (resp *http.Response, err error)
	GetUrl(path string) (res string)
	Get(path string, sets ...Set) (res []byte, err error)
	Post(path string, data any, sets ...Set) (res []byte, err error)
}

type Info struct {
}

type Request struct {
	*http.Request
}

func (this_ *Request) SetHeader(name, value string) *Request {
	this_.Request.Header.Set(name, value)
	return this_
}

type Response struct {
	*http.Response
}
