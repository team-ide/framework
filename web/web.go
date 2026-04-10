package web

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/team-ide/framework"
	"github.com/team-ide/framework/util"
)

type Option func(*WebServer)

func New(name string, webConfig *Config, opts ...Option) *WebServer {
	web := &WebServer{
		name:           name,
		webConfig:      webConfig,
		pathApiRouters: make(map[string][]*ApiRouter),
		Logger:         framework.Skip1Logger,
	}
	for _, opt := range opts {
		opt(web)
	}
	return web
}

type Config struct {
	// Disabled 禁用 上层 初始化服务时候 可以判断该属性 如果为 配置 true 则不去初始化服务
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	Host           string `json:"host,omitempty"`
	Port           int    `json:"port,omitempty"`
	Context        string `json:"context,omitempty"`
	AssetsDir      string `json:"assetsDir,omitempty"`
	AssetsSeparate string `json:"assetsSeparate,omitempty"` // assets 分割字符 默认 assets/
	FilesDir       string `json:"filesDir,omitempty"`
	FilesSeparate  string `json:"filesSeparate,omitempty"` // files 分割字符 默认 files/

	MaxMultipartMemory int64 `json:"maxMultipartMemory,omitempty"`

	Tls *ConfigTls `json:"tls,omitempty"`

	Locations     []*ConfigLocation         `json:"locations,omitempty"`
	Replaces      map[string]*ConfigReplace `json:"replaces,omitempty"`
	GinDefaultLog bool                      `json:"ginDefaultLog,omitempty"`
	GinErrorLog   bool                      `json:"ginErrorLog,omitempty"`
}

type ConfigTls struct {
	Open     bool   `json:"open,omitempty"`
	CertFile string `json:"certFile,omitempty"`
	KeyFile  string `json:"keyFile,omitempty"`
}

type ConfigLocation struct {
	Path string `json:"path,omitempty"`
	To   string `json:"to,omitempty"`
}

type ConfigReplace struct {
	Path        string `json:"path,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Content     string `json:"content,omitempty"`
	Append      string `json:"append,omitempty"`
}

func (this_ *Config) Init() {
	if this_.Context == "" {
		this_.Context = "/"
	}
	if !strings.HasPrefix(this_.Context, "/") {
		this_.Context = "/" + this_.Context
	}
	if !strings.HasSuffix(this_.Context, "/") {
		this_.Context += "/"
	}
	if this_.AssetsDir == "" {
		this_.AssetsDir = "assets/"
	}
	if !strings.HasSuffix(this_.AssetsDir, "/") {
		this_.AssetsDir += "/"
	}
	if this_.AssetsSeparate == "" {
		this_.AssetsSeparate = "assets/"
	}
	if this_.FilesDir == "" {
		this_.FilesDir = "files/"
	}
	if !strings.HasSuffix(this_.FilesDir, "/") {
		this_.FilesDir += "/"
	}
	if this_.FilesSeparate == "" {
		this_.FilesSeparate = "files/"
	}
	if this_.MaxMultipartMemory == 0 {
		this_.MaxMultipartMemory = 1024 << 20 // 1 G 最大上传大小
	}
	if this_.Host == "" {
		this_.Host = "0.0.0.0"
	}
}

func (this_ *WebServer) Start() (err error) {
	err = this_.init()
	if err != nil {
		return
	}
	err = this_.IWebService.Init()
	if err != nil {
		return
	}

	var Context = this_.webConfig.Context

	var Host = this_.webConfig.Host
	var Port = this_.webConfig.Port
	if Port <= 0 {
		err = fmt.Errorf("web server port error")
		this_.Info("web server port error")
		return
	}

	var ins []net.Interface
	ins, err = net.Interfaces()
	if err != nil {
		return
	}
	var serverUrl string
	s := "http"
	var t = &ConfigTls{}
	if this_.webConfig != nil && this_.webConfig.Tls != nil {
		t = this_.webConfig.Tls
	}
	if t.Open {
		s = "https"
	}
	if Host == "0.0.0.0" || Host == "::" {
		address := fmt.Sprintf("%s://127.0.0.1:%d%s", s, Port, Context)
		serverUrl = address
		this_.serverUrls = append(this_.serverUrls, address)
		this_.Info("web server url:" + address)
		for _, in := range ins {
			if in.Flags&net.FlagUp == 0 {
				continue
			}
			if in.Flags&net.FlagLoopback != 0 {
				continue
			}
			var adders []net.Addr
			adders, err = in.Addrs()
			if err != nil {
				return
			}
			for _, addr := range adders {
				ip := util.GetIpFromAddr(addr)
				if ip == nil {
					continue
				}
				address = fmt.Sprintf("%s://%s:%d%s", s, ip.String(), Port, Context)
				this_.serverUrls = append(this_.serverUrls, address)
				this_.Info("web server url:" + address)
			}
		}
	} else {
		address := fmt.Sprintf("%s://%s:%d%s", s, Host, Port, Context)
		serverUrl = address
		this_.serverUrls = append(this_.serverUrls, address)
		this_.Info("web server url:" + address)
	}
	addr := fmt.Sprintf("%s:%d", Host, Port)
	this_.Info("web server start", zap.Any("addr", addr))

	this_.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	this_.serverUrl = serverUrl

	err = this_.Serve()
	if err != nil {
		return
	}
	return
}

type IWebService interface {
	Init() (err error)
	Serve() (err error)
	RequestRead(request *WebRequest) []byte
	SetStatus(request *WebRequest, status int)
	GetHeader(request *WebRequest, key string) string
	SetHeader(request *WebRequest, key, value string)
	ResponseWrite(request *WebRequest, data []byte)
	ResponseWriteByReader(request *WebRequest, reader io.Reader) (written int64, err error)
	GetParam(request *WebRequest, key string) string
	RawQuery(request *WebRequest) string
	ClientIP(request *WebRequest) string
	UserAgent(request *WebRequest) string
}

type WebServer struct {
	name       string
	serverUrl  string
	serverUrls []string

	webConfig *Config

	filters []*Filter

	interceptors []*Interceptor

	apiRouters      []*ApiRouter
	pathApiRouters  map[string][]*ApiRouter
	matchApiRouters []*ApiRouter

	webDefaultWriter io.Writer
	webErrorWriter   io.Writer

	addr     string
	listener net.Listener

	IWebService

	framework.Logger

	webFilters      []WebFilter
	webInterceptors []WebInterceptor
	webApiRouters   []*WebApiRouter
}

func (this_ *WebServer) AddWebFilters(webFilters ...WebFilter) *WebServer {
	this_.webFilters = append(this_.webFilters, webFilters...)
	return this_
}

func (this_ *WebServer) AddWebInterceptors(webInterceptors ...WebInterceptor) *WebServer {
	this_.webInterceptors = append(this_.webInterceptors, webInterceptors...)
	return this_
}

func (this_ *WebServer) AddWebApiRouters(webApiRouters ...*WebApiRouter) *WebServer {
	this_.webApiRouters = append(this_.webApiRouters, webApiRouters...)
	return this_
}

func (this_ *WebServer) AddWebApis(webApis ...*WebApi) *WebServer {
	for _, webApi := range webApis {
		this_.AddWebApi(webApi)
	}
	return this_
}

func (this_ *WebServer) AddWebApi(webApi *WebApi) *WebServer {
	for _, router := range webApi.routers {
		router.Path = webApi.Path + router.Path
		if router.Method == "" {
			router.Method = webApi.Method
		}
		this_.webApiRouters = append(this_.webApiRouters, router)
	}
	return this_
}

func (this_ *WebServer) Close() {
	listener := this_.listener
	if listener != nil {
		_ = listener.Close()
	}
}
func (this_ *WebServer) GetAddr() string                { return this_.addr }
func (this_ *WebServer) GetListener() net.Listener      { return this_.listener }
func (this_ *WebServer) GetConfig() *Config             { return this_.webConfig }
func (this_ *WebServer) GetWebDefaultWriter() io.Writer { return this_.webDefaultWriter }
func (this_ *WebServer) GetWebErrorWriter() io.Writer   { return this_.webErrorWriter }

type Filter struct {
	WebFilter
	*MatchRule
}

type Interceptor struct {
	WebInterceptor
	*MatchRule
}

type ApiRouter struct {
	*WebApiRouter
	*MatchRule
}

func (this_ *WebServer) GetServerUrl() string {
	return this_.serverUrl
}
func (this_ *WebServer) GetServerUrls() []string {
	return this_.serverUrls
}

func (this_ *WebServer) init() (err error) {
	if this_.Logger == nil {
		this_.Logger = framework.DefaultLogger
	}
	this_.webConfig.Init()
	if this_.webConfig.GinDefaultLog {
		this_.webDefaultWriter = WebDefaultWriter
	} else {
		this_.webDefaultWriter = WebDefaultNotWriter
	}
	if this_.webConfig.GinErrorLog {
		this_.webErrorWriter = WebErrorWriter
	} else {
		this_.webErrorWriter = WebErrorNotWriter
	}

	err = this_.initFilters()
	if err != nil {
		return
	}
	err = this_.initInterceptors()
	if err != nil {
		return
	}
	err = this_.initApiRouters()
	if err != nil {
		return
	}
	return
}

func (this_ *WebServer) initFilters() (err error) {
	list := GetWebFilterList(this_.name)
	list = append(list, this_.webFilters...)

	// Order 正序 排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})
	for _, one := range list {
		to := &Filter{
			WebFilter: one,
		}
		match := one.GetMatch()
		to.MatchRule, err = this_.toMatchRule(match.Includes, match.Excludes, match.Methods)
		if err != nil {
			this_.Error("filter ["+one.GetName()+"] init error", zap.Error(err))
			return
		}
		if to.MatchRule == nil {
			this_.Warn("filter ["+one.GetName()+"] init matchRule is null", zap.Any("match", match))
			continue
		}
		this_.filters = append(this_.filters, to)
	}
	return
}

func (this_ *WebServer) initInterceptors() (err error) {
	list := GetWebInterceptorList(this_.name)
	list = append(list, this_.webInterceptors...)

	// Order 正序 排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})
	for _, one := range list {
		to := &Interceptor{
			WebInterceptor: one,
		}
		match := one.GetMatch()
		to.MatchRule, err = this_.toMatchRule(match.Includes, match.Excludes, match.Methods)
		if err != nil {
			this_.Error("interceptor ["+one.GetName()+"] init error", zap.Error(err))
			return
		}
		if to.MatchRule == nil {
			this_.Warn("interceptor ["+one.GetName()+"] init matchRule is null", zap.Any("match", match))
			continue
		}
		this_.interceptors = append(this_.interceptors, to)
	}
	return
}

func (this_ *WebServer) initApiRouters() (err error) {
	list := GetWebApiRouterList(this_.name)
	list = append(list, this_.webApiRouters...)

	for _, one := range list {
		this_.addApiRouter(one)
	}
	return
}

func (this_ *WebServer) addApiRouter(apiRouter *WebApiRouter) {
	to := &ApiRouter{
		WebApiRouter: apiRouter,
	}
	var includes []string
	var methods []string
	if apiRouter.Path != "" {
		includes = strings.Split(apiRouter.Path, ",")
	}
	if apiRouter.Method != "" {
		methods = strings.Split(apiRouter.Method, ",")
	}
	var err error
	to.MatchRule, err = this_.toMatchRule(includes, []string{}, methods)
	if err != nil {
		this_.Error("api router ["+apiRouter.Path+"] init error", zap.Error(err))
		return
	}
	if to.MatchRule == nil {
		this_.Warn("api router ["+apiRouter.Path+"] init matchRule is null", zap.Any("includes", includes), zap.Any("methods", methods))
		return
	}
	if to.includePaths != "" {
		pathList := strings.Split(to.MatchRule.includePaths, ",")
		for _, path := range pathList {
			this_.Info("bind api router path [" + path + "]")
			this_.pathApiRouters[path] = append(this_.pathApiRouters[path], to)
		}
	}
	if len(to.includeRegexps) > 0 {
		this_.matchApiRouters = append(this_.matchApiRouters, to)
	}

	return
}

func (this_ *WebServer) findApiRouter(request *WebRequest) (apiRouter *WebApiRouter) {
	finds := this_.pathApiRouters[request.Path]
	for _, find := range finds {

		if find.methodAny {
			apiRouter = find.WebApiRouter
			return
		}
		if strings.Contains(find.methodNames, strings.ToLower(request.Method)+",") {
			apiRouter = find.WebApiRouter
			return
		}
	}
	for _, one := range this_.matchApiRouters {
		if one.match(request) {
			apiRouter = one.WebApiRouter
			return
		}
	}

	return
}

type MatchRule struct {
	includePaths   string
	includeRegexps []*regexp.Regexp
	excludePaths   string
	excludeRegexps []*regexp.Regexp
	methodNames    string
	methodAny      bool
}

func (this_ *WebServer) toMatchRule(includes []string, excludes []string, methods []string) (matchRule *MatchRule, err error) {
	if len(includes) == 0 {
		return
	}
	matchRule = &MatchRule{}
	if len(methods) > 0 {
		for i, one := range methods {
			methods[i] = strings.ToLower(strings.TrimSpace(one))
			if strings.EqualFold(methods[i], "any") {
				matchRule.methodAny = true
			}
		}
		matchRule.methodNames = strings.Join(methods, ",") + ","
	} else {
		matchRule.methodAny = true
	}
	var re *regexp.Regexp
	for _, one := range includes {
		one = strings.TrimSpace(one)
		if !strings.HasPrefix(one, "/") {
			one = "/" + one
		}

		one = pathReplaceCompile.ReplaceAllLiteralString(one, "/")
		if strings.Contains(one, "*") {
			// 将通配符模式中的 * 替换为 .*
			pattern := strings.ReplaceAll(one, "*", ".*")
			// 编译正则表达式
			compiledPattern := strings.ReplaceAll(pattern, "/", `\/`)
			compiledPattern = "^" + compiledPattern + "$"
			re, err = regexp.Compile(compiledPattern)
			if err != nil {
				return
			}
			matchRule.includeRegexps = append(matchRule.includeRegexps, re)
		} else {
			matchRule.includePaths += one + ","
		}
	}

	for _, one := range excludes {
		one = strings.TrimSpace(one)
		if !strings.HasPrefix(one, "/") {
			one = "/" + one
		}
		one = pathReplaceCompile.ReplaceAllLiteralString(one, "/")
		if strings.Contains(one, "*") {
			// 将通配符模式中的 * 替换为 .*
			pattern := strings.ReplaceAll(one, "*", ".*")
			// 编译正则表达式
			compiledPattern := strings.ReplaceAll(pattern, "/", `\/`)
			compiledPattern = "^" + compiledPattern + "$"
			re, err = regexp.Compile(compiledPattern)
			if err != nil {
				return
			}
			matchRule.excludeRegexps = append(matchRule.excludeRegexps, re)
		} else {
			matchRule.excludePaths += one + ","
		}
	}

	return
}

func (this_ *MatchRule) match(request *WebRequest) (res bool) {

	if !this_.methodAny && this_.methodNames != "" {
		if !strings.Contains(this_.methodNames, strings.ToLower(request.Method)+",") {
			return
		}
	}
	path := request.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 匹配 忽略 地址 直接 返回
	if this_.excludePaths != "" {
		if strings.Contains(this_.excludePaths, path+",") {
			return
		}
	}

	// 匹配 忽略 地址 直接 返回
	for _, one := range this_.excludeRegexps {
		if one.MatchString(path) {
			return
		}
	}

	if this_.includePaths != "" {
		if strings.Contains(this_.includePaths, path+",") {
			res = true
			return
		}
	}
	for _, one := range this_.includeRegexps {
		if one.MatchString(path) {
			res = true
			return
		}
	}

	return
}

type WebRequestMatch struct {
	Includes []string // 包含路径 表达式 匹配 路径
	Excludes []string // 忽略路径 表达式 匹配 路径
	Methods  []string // HTTP 方法 GET、POST 等
}

type WebFilter interface {
	GetName() string           // 定义名称
	Order() int                // 顺序 正序 排序 值越小 则优先
	GetMatch() WebRequestMatch // 请求 匹配 配置
	DoFilter(request *WebRequest, chain WebFilterChain) (err error)
}

type WebFilterChain func(request *WebRequest) (err error)

var (
	webFilterList = map[string][]WebFilter{}
)

func AppendWebFilter(webName string, webFilter WebFilter) {
	webFilterList[webName] = append(webFilterList[webName], webFilter)
}

func GetWebFilterList(webName string) []WebFilter {
	return webFilterList[webName]
}

type WebInterceptor interface {
	GetName() string           // 定义名称
	Order() int                // 顺序 正序 排序 值越小 则优先
	GetMatch() WebRequestMatch // 请求 匹配 配置
	Before(request *WebRequest) (toContinue bool, err error)
	After(request *WebRequest) (toContinue bool, err error)
}

var (
	webInterceptorList = map[string][]WebInterceptor{}
)

func AppendWebInterceptor(webName string, webInterceptor WebInterceptor) {
	webInterceptorList[webName] = append(webInterceptorList[webName], webInterceptor)
}

func GetWebInterceptorList(webName string) []WebInterceptor {
	return webInterceptorList[webName]
}

func NewWebApi(path string) *WebApi {
	res := &WebApi{}
	res.Path = path
	return res
}

type WebApi struct {
	Path    string // 路由
	Method  string // HTTP 方法 GET、POST 等
	routers []*WebApiRouter
}

func (this_ *WebApi) SetMethod(method string) *WebApi {
	this_.Method = method
	return this_
}

func (this_ *WebApi) Add(path string, handle WebApiHandleFunc) *WebApiRouter {
	res := &WebApiRouter{}
	res.Path = path
	res.handle = handle
	this_.routers = append(this_.routers, res)
	return res
}

func (this_ *WebApi) Router(router *WebApiRouter) *WebApi {

	this_.routers = append(this_.routers, router)
	return this_
}

type WebApiRouter struct {
	Path     string // 路由
	Method   string // HTTP 方法 GET、POST 等
	Comment  string // 说明
	NotLogin bool   // 不需要登录
	NotLog   bool   // 不需要日志

	handle WebApiHandleFunc
}

func NewApiRouter(path string) *WebApiRouter {
	res := &WebApiRouter{}
	res.Path = path
	return res
}

var (
	webApiRouterList = map[string][]*WebApiRouter{}
)

func AppendWebApiRouter(webName string, webApiRouter *WebApiRouter, handle WebApiHandleFunc) {
	webApiRouter.handle = handle
	webApiRouterList[webName] = append(webApiRouterList[webName], webApiRouter)
}

func GetWebApiRouterList(webName string) []*WebApiRouter {
	return webApiRouterList[webName]
}

type WebApiHandleFunc func(request *WebRequest) (res any, err error)

func (this_ *WebApiRouter) SetHandle(handle WebApiHandleFunc) *WebApiRouter {
	this_.handle = handle
	return this_
}
func (this_ *WebApiRouter) SetNotLogin() *WebApiRouter {
	this_.NotLogin = true
	return this_
}
func (this_ *WebApiRouter) SetNotLog() *WebApiRouter {
	this_.NotLog = true
	return this_
}
func (this_ *WebApiRouter) SetGet() *WebApiRouter {
	this_.Method = "GET"
	return this_
}
func (this_ *WebApiRouter) SetPost() *WebApiRouter {
	this_.Method = "POST"
	return this_
}
func (this_ *WebApiRouter) SetMethod(method string) *WebApiRouter {
	this_.Method = method
	return this_
}
func (this_ *WebApiRouter) SetComment(comment string) *WebApiRouter {
	this_.Comment = comment
	return this_
}

func (this_ *WebApiRouter) GetHandle() WebApiHandleFunc {
	return this_.handle
}

func NewWebRequest(requestCtx any) *WebRequest {
	res := new(WebRequest)
	res.requestCtx = requestCtx
	return res
}

type WebRequest struct {
	Path            string        `json:"path,omitempty"`
	Method          string        `json:"method,omitempty"`
	WebSession      *WebSession   `json:"webSession,omitempty"`
	StartTime       time.Time     `json:"startTime,omitempty"`
	EndTime         time.Time     `json:"endTime,omitempty"`
	HandleStartTime time.Time     `json:"handleStartTime,omitempty"`
	HandleEndTime   time.Time     `json:"handleEndTime,omitempty"`
	Response        any           `json:"response,omitempty"`
	WebApiRouter    *WebApiRouter `json:"webApiRouter,omitempty"`

	Body []byte `json:"-"`

	requestCtx any

	webServer *WebServer
}

func (this_ *WebRequest) GetHeader(key string) string {
	return this_.webServer.GetHeader(this_, key)
}

func (this_ *WebRequest) GetParam(key string) string {
	return this_.webServer.GetParam(this_, key)
}

func (this_ *WebRequest) ClientIP() string {
	return this_.webServer.ClientIP(this_)
}

func (this_ *WebRequest) UserAgent() string {
	return this_.webServer.UserAgent(this_)
}

func (this_ *WebRequest) RawQuery() string {
	return this_.webServer.RawQuery(this_)
}

func (this_ *WebRequest) GetData() []byte {
	return this_.webServer.RequestRead(this_)
}

func (this_ *WebRequest) RequestJSON(data any) (err error) {
	bs := this_.webServer.RequestRead(this_)
	if len(bs) > 0 {
		err = util.JsonBytesToObj(bs, data)
	}
	return
}

func (this_ *WebRequest) GetRequestCtx() any {
	return this_.requestCtx
}

type WebSession struct {
	// @Tag:web_session_struct_fields
}

type WebResponse struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

var (
	WebNotResponse = new(any) // 接口 返回该结果 表示 web 框架 不输出 结果
)

func FormatTime(t time.Time) (res string) {
	return t.Format("2006-01-02 15:04:05.000")
}

var (
	WebDefaultWriter    = &webDefaultWriter{outLog: true}
	WebDefaultNotWriter = &webDefaultWriter{outLog: false}

	WebErrorWriter    = &webErrorWriter{outLog: true}
	WebErrorNotWriter = &webErrorWriter{outLog: false}
)

type webDefaultWriter struct {
	outLog bool
}

func (this_ *webDefaultWriter) Write(bs []byte) (int, error) {
	if this_.outLog {
		framework.Debug(string(bs))
	}
	return len(bs), nil
}

type webErrorWriter struct {
	outLog bool
}

func (this_ *webErrorWriter) Write(bs []byte) (int, error) {
	if this_.outLog {
		framework.Error(string(bs))
	}
	return len(bs), nil
}
