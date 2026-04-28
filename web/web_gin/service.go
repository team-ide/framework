package web_gin

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/team-ide/framework/web"
)

func BindWebService(server *web.WebServer) {
	res := &WebServiceGin{}
	res.server = server
	server.IWebService = res
	return
}

type WebServiceGin struct {
	server *web.WebServer
	engine *gin.Engine
}

func (this_ *WebServiceGin) Init() (err error) {

	gin.DefaultWriter = this_.server.GetWebDefaultWriter()
	gin.DefaultErrorWriter = this_.server.GetWebErrorWriter()

	this_.engine = gin.Default()
	this_.engine.MaxMultipartMemory = this_.server.GetConfig().MaxMultipartMemory

	var Context = this_.server.GetConfig().Context
	routerGroup := this_.engine.Group(Context)

	// 绑定 注册的 API 路由
	routerGroup.Any("*path", func(c *gin.Context) {
		request := web.NewWebRequest(c)
		request.StartTime = time.Now()
		request.Method = c.Request.Method
		request.Path = c.Request.URL.Path
		this_.server.DoRequest(request)
	})
	return
}

func (this_ *WebServiceGin) Serve() (err error) {

	t := this_.server.GetConfig().Tls
	if t == nil {
		t = &web.ConfigTls{}
	}

	ss := &http.Server{
		Addr:    this_.server.GetAddr(),
		Handler: this_.engine,
	}
	ss.ErrorLog = log.New(this_.server.GetWebErrorWriter(), "", 0)
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

func (this_ *WebServiceGin) RequestRead(request *web.WebRequest) []byte {
	if request.Body == nil {
		c := request.GetRequestCtx().(*gin.Context)
		if c.Request.Body != nil {
			request.Body, _ = io.ReadAll(c.Request.Body)
		}
	}
	return request.Body
}

func (this_ *WebServiceGin) SetStatus(request *web.WebRequest, status int) {
	c := request.GetRequestCtx().(*gin.Context)
	c.Status(status)
}

func (this_ *WebServiceGin) GetHeader(request *web.WebRequest, key string) string {
	c := request.GetRequestCtx().(*gin.Context)
	return c.GetHeader(key)
}

func (this_ *WebServiceGin) SetHeader(request *web.WebRequest, key, value string) {
	c := request.GetRequestCtx().(*gin.Context)
	c.Header(key, value)
}

func (this_ *WebServiceGin) ResponseWrite(request *web.WebRequest, data []byte) {
	c := request.GetRequestCtx().(*gin.Context)
	_, _ = c.Writer.Write(data)
}

func (this_ *WebServiceGin) ResponseWriteByReader(request *web.WebRequest, reader io.Reader) (written int64, err error) {
	c := request.GetRequestCtx().(*gin.Context)
	written, err = io.Copy(c.Writer, reader)
	return
}

func (this_ *WebServiceGin) GetParam(request *web.WebRequest, key string) string {
	c := request.GetRequestCtx().(*gin.Context)
	return c.Query(key)
}

func (this_ *WebServiceGin) RawQuery(request *web.WebRequest) string {
	c := request.GetRequestCtx().(*gin.Context)
	return c.Request.RequestURI
}

func (this_ *WebServiceGin) ClientIP(request *web.WebRequest) string {
	c := request.GetRequestCtx().(*gin.Context)
	return c.ClientIP()
}

func (this_ *WebServiceGin) UserAgent(request *web.WebRequest) string {
	c := request.GetRequestCtx().(*gin.Context)
	return c.Request.UserAgent()
}

func (this_ *WebServiceGin) GetFiles(name string, request *web.WebRequest) (res []*web.UploadFile, err error) {
	c := request.GetRequestCtx().(*gin.Context)
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

func (this_ *WebServiceGin) GetWriter(request *web.WebRequest) io.Writer {
	c := request.GetRequestCtx().(*gin.Context)
	return c.Writer
}
