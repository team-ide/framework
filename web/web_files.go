package web

import (
	"github.com/team-ide/framework/util"
	"os"
	"path"
)

func (this_ *WebServer) toFiles(requestInfo string, request *WebRequest) (ok bool, err error) {
	files := this_.webConfig.Files
	if files == nil || !files.Open || files.Dir == "" {
		return
	}
	parentDir := files.Dir
	filePath := request.Path
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

	_, err = this_.ResponseWriteByReader(request, f)
	if err != nil {
		return
	}
	ok = true

	return
}
