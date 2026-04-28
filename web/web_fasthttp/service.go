package web_fasthttp

import (
	"fmt"
	"io"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/team-ide/framework/web"
)

func BindWebService(server *web.WebServer) {
	res := &WebServiceFast{}
	res.server = server
	server.IWebService = res
	return
}

type WebServiceFast struct {
	server *web.WebServer
}

func (this_ *WebServiceFast) Init() (err error) {
	return
}

func (this_ *WebServiceFast) Printf(format string, args ...any) {
	_, _ = this_.server.GetWebDefaultWriter().Write([]byte(fmt.Sprintf(format, args...)))
}
func (this_ *WebServiceFast) Serve() (err error) {

	t := this_.server.GetConfig().Tls
	if t == nil {
		t = &web.ConfigTls{}
	}
	ss := &fasthttp.Server{
		Handler: this_.handler,
	}
	ss.Logger = this_
	go func() {
		if t.Open {
			err = ss.ServeTLS(this_.server.GetListener(), t.CertFile, t.KeyFile)
		} else {
			err = ss.Serve(this_.server.GetListener())
		}
		if err != nil {
			//this_.Error("web server serve error", zap.Error(err))
			return
		}
	}()
	// 等待 100 毫秒
	time.Sleep(time.Millisecond * 100)
	return
}

func (this_ *WebServiceFast) handler(c *fasthttp.RequestCtx) {
	request := web.NewWebRequest(c)
	request.StartTime = time.Now()
	request.Method = string(c.Method())
	request.Path = string(c.Path())
	this_.server.DoRequest(request)

	return
}

func (this_ *WebServiceFast) RequestRead(request *web.WebRequest) []byte {
	if request.Body == nil {
		c := request.GetRequestCtx().(*fasthttp.RequestCtx)
		request.Body = c.PostBody()
	}
	return request.Body
}

func (this_ *WebServiceFast) SetStatus(request *web.WebRequest, status int) {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	c.Response.SetStatusCode(status)
}

func (this_ *WebServiceFast) GetHeader(request *web.WebRequest, key string) string {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	return string(c.Request.Header.Peek(key))
}

func (this_ *WebServiceFast) SetHeader(request *web.WebRequest, key, value string) {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	c.Response.Header.Set(key, value)
}

func (this_ *WebServiceFast) ResponseWrite(request *web.WebRequest, data []byte) {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	c.Response.SetBody(data)
}

func (this_ *WebServiceFast) ResponseWriteByReader(request *web.WebRequest, reader io.Reader) (written int64, err error) {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	written, err = io.Copy(c.Response.BodyWriter(), reader)
	return
}

func (this_ *WebServiceFast) GetParam(request *web.WebRequest, key string) string {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	return string(c.QueryArgs().Peek(key))
}

func (this_ *WebServiceFast) RawQuery(request *web.WebRequest) string {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	return string(c.RequestURI())
}

func (this_ *WebServiceFast) ClientIP(request *web.WebRequest) string {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	return c.RemoteIP().String()
}

func (this_ *WebServiceFast) UserAgent(request *web.WebRequest) string {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	return string(c.UserAgent())
}

func (this_ *WebServiceFast) GetFiles(name string, request *web.WebRequest) (res []*web.UploadFile, err error) {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	// 获取表单数据中的文件字段，这里假设字段名为"files"
	form, err := c.MultipartForm()
	if err != nil {
		return
	}

	files := form.File[name] // 获取名为"files"的所有文件
	for _, file := range files {
		f := &web.UploadFile{}
		f.Filename = file.Filename
		f.Size = file.Size
		f.ReadCloser, err = file.Open()
		if err != nil {
			return
		}
		res = append(res, f)
	}
	return
}

func (this_ *WebServiceFast) GetWriter(request *web.WebRequest) io.Writer {
	c := request.GetRequestCtx().(*fasthttp.RequestCtx)
	return c
}
