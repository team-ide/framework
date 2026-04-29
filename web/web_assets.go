package web

import (
	"github.com/team-ide/framework/util"
	"os"
	"path"
	"strings"
)

func (this_ *WebServer) toAssets(requestInfo string, request *WebRequest) (ok bool, err error) {
	assets := this_.webConfig.Assets
	if assets == nil || !assets.Open || assets.Dir == "" {
		return
	}
	parentDir := assets.Dir
	filePath := request.Path
	if filePath == "" || filePath == "/" {
		filePath = "index.html"
	}
	assetPath := path.Join(parentDir, filePath)
	isSub, err := util.IsSubPath(parentDir, assetPath)
	if err != nil || !isSub {
		return
	}
	isExists, err := util.PathExists(assetPath)
	if err != nil || !isExists {
		return
	}
	f, err := os.Open(assetPath)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	if strings.HasSuffix(filePath, ".html") {
		this_.SetHeader(request, "Content-Type", "text/html")
		this_.SetHeader(request, "Cache-Control", "no-cache")
	} else if strings.HasSuffix(filePath, ".css") {
		this_.SetHeader(request, "Content-Type", "text/css")
		// max-age 缓存 过期时间 秒为单位
		this_.SetHeader(request, "Cache-Control", "max-age=31536000")
	} else if strings.HasSuffix(filePath, ".js") {
		this_.SetHeader(request, "Content-Type", "application/javascript")
		// max-age 缓存 过期时间 秒为单位
		this_.SetHeader(request, "Cache-Control", "max-age=31536000")
	} else if strings.HasSuffix(filePath, ".woff") ||
		strings.HasSuffix(filePath, ".ttf") ||
		strings.HasSuffix(filePath, ".woff2") ||
		strings.HasSuffix(filePath, ".eot") {
		// max-age 缓存 过期时间 秒为单位
		this_.SetHeader(request, "Cache-Control", "max-age=31536000")
	}

	_, err = this_.ResponseWriteByReader(request, f)
	if err != nil {
		return
	}
	ok = true

	return
}
