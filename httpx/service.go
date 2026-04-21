package httpx

import (
	"crypto/tls"
	"errors"
	"github.com/team-ide/framework/util"
	"io"
	"net/http"
	"time"
)

type Service struct {
	*Config
	httpClient *http.Client
	isClosed   bool
}

func (this_ *Service) init(connProxy ConnProxy) (err error) {
	// 创建传输对象
	transport := &http.Transport{
		MaxIdleConns:    10,
		MaxConnsPerHost: 10,
		IdleConnTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			// 指定不校验 SSL/TLS 证书
			InsecureSkipVerify: false,
		},
	}
	if this_.MaxIdleConns > 0 {
		transport.MaxIdleConns = this_.MaxIdleConns
	}
	if this_.MaxConnsPerHost > 0 {
		transport.MaxConnsPerHost = this_.MaxConnsPerHost
	}
	if this_.IdleConnTimeout > 0 {
		transport.IdleConnTimeout = time.Millisecond * time.Duration(this_.IdleConnTimeout)
	}
	if this_.Tls != nil {
		if this_.Tls.InsecureSkipVerify {
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Transport: transport,
	}
	if this_.Timeout > 0 {
		client.Timeout = time.Millisecond * time.Duration(this_.Timeout)
	}

	this_.httpClient = client

	this_.isClosed = false
	return
}

type Set func(in *http.Request)

func (this_ *Service) GetUrl(path string) (res string) {
	res = util.PathJoin(this_.RootUrl, path)
	return
}
func (this_ *Service) Request(method, path string, body io.Reader, sets ...Set) (resp *http.Response, err error) {
	if this_.isClosed {
		err = errors.New("http [" + this_.RootUrl + "] service is closed")
		return
	}
	url := this_.GetUrl(path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		err = errors.New("http [" + this_.RootUrl + "] service NewRequest error:" + err.Error())
		return
	}
	for _, set := range sets {
		set(req)
	}
	resp, err = this_.httpClient.Do(req)
	if err != nil {
		return
	}
	return
}
func (this_ *Service) GetRequest(path string, sets ...Set) (resp *http.Response, err error) {
	resp, err = this_.Request("GET", path, nil, sets...)
	if err != nil {
		return
	}
	return
}
func (this_ *Service) PostRequest(path string, body io.Reader, sets ...Set) (resp *http.Response, err error) {
	resp, err = this_.Request("POST", path, body, sets...)
	if err != nil {
		return
	}
	return
}
func (this_ *Service) Get(path string, sets ...Set) (res string, err error) {
	resp, err := this_.GetRequest(path, sets...)
	if err != nil {
		return
	}
	if resp.Body == nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	res = string(bs)
	return
}

func (this_ *Service) Post(path string, body io.Reader, sets ...Set) (res string, err error) {
	resp, err := this_.PostRequest(path, body, sets...)
	if err != nil {
		return
	}
	if resp.Body == nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	res = string(bs)
	return
}

func (this_ *Service) Info() (res *Info, err error) {
	return
}
func (this_ *Service) Close() {

	if this_ == nil {
		return
	}
	if this_.isClosed {
		return
	}
	this_.isClosed = true

	client := this_.httpClient
	this_.httpClient = nil
	if client != nil {
		client.CloseIdleConnections()
	}
}
