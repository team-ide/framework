package web

import (
	"errors"
	"fmt"
	"github.com/team-ide/framework"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/team-ide/framework/util"
)

func (this_ *WebServer) Bin(apiRouters ...*WebApiRouter) {
	if apiRouters == nil {
		return
	}
	for _, apiRouter := range apiRouters {
		if apiRouter == nil || apiRouter.GetHandle() == nil {
			return
		}
		this_.addApiRouter(apiRouter)
	}
	return
}

var (
	pathReplaceCompile, _ = regexp.Compile("/+")
)

var NotFoundError = errors.New("404 page not found")

func (this_ *WebServer) doApiRouterHandle(requestInfo string, request *WebRequest) (err error) {

	request.HandleStartTime = time.Now()
	if request.WebApiRouter == nil {
		// 跳转到 资源文件
		var ok bool
		ok, err = this_.toAssets(requestInfo, request)
		if err != nil {
			return
		}
		if !ok {
			ok, err = this_.toFiles(requestInfo, request)
			if err != nil {
				return
			}
		}
		request.Response = WebNotResponse
		if !ok {
			err = NotFoundError
			return
		}
	} else {
		request.Response, err = request.WebApiRouter.GetHandle()(request)
	}
	request.HandleEndTime = time.Now()
	if err != nil {
		this_.Error(requestInfo+" api router handle error", zap.Error(err))
		return
	}
	return
}

func (this_ *WebServer) DoRequest(request *WebRequest) {
	request.webServer = this_

	request.Path = pathReplaceCompile.ReplaceAllLiteralString(request.Path, "/")
	if this_.webConfig.Context != "/" {
		if this_.webConfig.Context == request.Path+"/" {
			request.Path = "/"
		} else {
			request.Path = "/" + strings.TrimPrefix(request.Path, this_.webConfig.Context)
		}
	}

	requestInfo := "request [" + request.Path + "] [" + request.Method + "]"

	this_.Debug(requestInfo + " start")
	var err error
	var isNotFoundError bool
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
			this_.Error(requestInfo + " panic error:" + fmt.Sprint(e))
		}
		// 如果 结果 是不输出 则跳过
		if !isNotFoundError && request.Response != WebNotResponse {
			this_.ResponseJsonData(request, request.Response, err)
		}
	}()
	request.WebApiRouter = this_.findApiRouter(request)
	// this_.Debug(requestInfo + " do filters")
	err = this_.doFilters(requestInfo, request)
	if err != nil {
		isNotFoundError = errors.Is(err, NotFoundError)
		if isNotFoundError {
			err = nil
			this_.SetStatus(request, http.StatusNotFound)
			this_.SetHeader(request, "Content-Type", "text/plain; charset=utf-8")
			this_.ResponseWrite(request, []byte("404 page ["+request.Path+"] not found"))
			return
		}
		return
	}

	request.EndTime = time.Now()
	useTime := request.EndTime.UnixMilli() - request.StartTime.UnixMilli()
	var handleUseTime int64
	if !request.HandleEndTime.IsZero() {
		handleUseTime = request.HandleEndTime.UnixMilli() - request.HandleStartTime.UnixMilli()
	}
	var log = fmt.Sprintf(", useTime:%dms, apiHandleUseTime:%dms", useTime, handleUseTime)
	this_.Info(requestInfo + " end " + log)
	return
}

func (this_ *WebServer) ResponseJsonData(request *WebRequest, data any, err error) {
	response := WebResponse{
		Code: "0",
		Data: data,
	}
	if err != nil {
		var toE framework.TError
		if errors.As(err, &toE) {
			response.Code = toE.GetCode()
			response.Msg = toE.GetMsg()
		} else {
			response.Code = "-1"
			response.Msg = err.Error()
		}
	} else {
		response.Data = data
	}

	resBs, e := util.ObjToJsonBytes(response)
	if e != nil {
		response.Msg = "response json data to json error:" + e.Error()
		this_.Error(response.Msg)
		response.Data = nil
		response.Code = "-1"
		resBs, _ = util.ObjToJsonBytes(response)
	}
	this_.SetStatus(request, http.StatusOK)
	this_.SetHeader(request, "Content-Type", "application/json; charset=utf-8")
	this_.ResponseWrite(request, resBs)
}

func (this_ *WebServer) doInterceptors(requestInfo string, request *WebRequest) (err error) {
	// this_.Debug(requestInfo + " do interceptors before")
	doContinue, err := this_.doInterceptorsBefore(requestInfo, request)
	if err != nil {
		this_.Error(requestInfo+" doInterceptorsBefore error", zap.Error(err))
		return
	}
	// 如果 InterceptorsBefore 返回 false，则直接返回
	if !doContinue {
		return
	}
	// this_.Debug(requestInfo + " do api router handle")
	err = this_.doApiRouterHandle(requestInfo, request)
	if err != nil {
		return
	}
	// this_.Debug(requestInfo + " do interceptors after")
	doContinue, err = this_.doInterceptorsAfter(requestInfo, request)
	if err != nil {
		this_.Error(requestInfo+" doInterceptorsAfter error", zap.Error(err))
		return
	}
	// 如果 InterceptorsAfter 返回 false，则直接返回
	if !doContinue {
		return
	}
	return
}

func (this_ *WebServer) doFilters(requestInfo string, request *WebRequest) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
			this_.Error(requestInfo+" doFilters panic error", zap.Error(err))
		}
	}()

	err = this_.doFilter(requestInfo, request, 0)
	return
}

func (this_ *WebServer) doFilter(requestInfo string, request *WebRequest, index int) (err error) {
	if index < len(this_.filters) {
		nextIndex := index + 1
		filter := this_.filters[index]
		if filter.match(request) {
			// this_.Debug(requestInfo + " do filter [" + filter.GetName() + "]")
			err = filter.DoFilter(request, func(request *WebRequest) (err error) {
				err = this_.doFilter(requestInfo, request, nextIndex)
				return err
			})
			if err != nil {
				this_.Error(requestInfo+" doFilter ["+filter.GetName()+"] error", zap.Error(err))
				return
			}
			return
		}
		err = this_.doFilter(requestInfo, request, nextIndex)
		return
	}
	err = this_.doInterceptors(requestInfo, request)
	return
}

func (this_ *WebServer) doInterceptorsBefore(requestInfo string, request *WebRequest) (res bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
			this_.Error(requestInfo+" doInterceptorsBefore panic error", zap.Error(err))
		}
	}()

	res = true

	for _, one := range this_.interceptors {
		if !one.match(request) {
			continue
		}
		res, err = one.Before(request)
		if err != nil {
			this_.Error(requestInfo+" doInterceptorsBefore ["+one.GetName()+"] error", zap.Error(err))
			return
		}
		if !res {
			break
		}
	}
	return
}

func (this_ *WebServer) doInterceptorsAfter(requestInfo string, request *WebRequest) (res bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
			this_.Error(requestInfo+" doInterceptorsAfter panic error", zap.Error(err))
		}
	}()

	res = true

	for _, one := range this_.interceptors {
		if !one.match(request) {
			continue
		}
		res, err = one.After(request)
		if err != nil {
			this_.Error(requestInfo+" doInterceptorsAfter ["+one.GetName()+"] error", zap.Error(err))
			return
		}
		if !res {
			break
		}
	}
	return
}
