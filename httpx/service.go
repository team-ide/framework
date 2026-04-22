package httpx

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
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
	res = this_.RootUrl + path
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
func (this_ *Service) Get(path string, sets ...Set) (res []byte, err error) {
	resp, err := this_.GetRequest(path, sets...)
	if err != nil {
		return
	}
	res, err = this_.ReadResponse(resp)
	return
}

func (this_ *Service) Post(path string, data any, sets ...Set) (res []byte, err error) {
	body, err := this_.DataReader(data)
	if err != nil {
		return
	}
	resp, err := this_.PostRequest(path, body, sets...)
	if err != nil {
		return
	}
	res, err = this_.ReadResponse(resp)
	return
}

func (this_ *Service) DataReader(data any) (res io.Reader, err error) {
	if data == nil {
		return
	}
	var bs []byte
	bs, err = util.ObjToJsonBytes(data)
	if err != nil {
		err = errors.New("data reader data to json bytes error:" + err.Error())
		return
	}
	res = bytes.NewReader(bs)
	return
}
func (this_ *Service) ReadResponse(resp *http.Response) (res []byte, err error) {
	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status)
		return
	}
	if resp.Body == nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	res, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
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
func HttpGet[T any](ser IService, path string, sets ...Set) (res T, body []byte, err error) {
	body, err = ser.Get(path, sets...)
	if err != nil {
		return
	}
	res, err = HttpJsonToObj[T](body)
	return
}
func HttpPost[T any](ser IService, path string, in any, sets ...Set) (res T, body []byte, err error) {
	body, err = ser.Post(path, in, sets...)
	if err != nil {
		return
	}
	res, err = HttpJsonToObj[T](body)
	fmt.Println("body:" + string(body))
	return
}

type DataResponse[T any] struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func HttpPostDataResponse[T any](ser IService, path string, in any, sets ...Set) (res *DataResponse[T], body []byte, err error) {
	body, err = ser.Post(path, in, sets...)
	if err != nil {
		return
	}
	res, err = HttpJsonToObj[*DataResponse[T]](body)
	fmt.Println("body:" + string(body))
	return
}

func HttpJsonToObj[T any](bs []byte) (res T, err error) {
	// 创建新的实例并反序列化
	var result T
	// res 是指针，直接反序列化到指针
	err = util.JsonBytesToObj(bs, &result)
	if err != nil {
		err = fmt.Errorf("json unmarshal error: %w", err)
		return
	}
	res = result
	return
}
