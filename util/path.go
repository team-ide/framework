package util

import (
	"os"
	"path/filepath"
	"strings"
)

// FormatPath 格式化路径
// FormatPath("/x/x/xxx\xx\xx")
func FormatPath(path string) string {

	var abs string
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	res := filepath.ToSlash(abs)
	return res
}

// PathExists 路径文件是否存在
// PathExists("/x/x/xxx\xx\xx")
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func PathJoin(path ...string) (res string) {
	var ss []string
	for _, v := range path {
		if v == "" {
			continue
		}
		ss = append(ss, v)
	}
	res = filepath.Join(ss...)
	res = filepath.ToSlash(res)
	return res
}

// IsSubPath child是否是parent子路径
// IsSubPath("/a/b", "/a/b/c")
func IsSubPath(parent, child string) (isSub bool, err error) {
	parentPath, err := filepath.Abs(parent)
	if err != nil {
		return
	}
	parentPath = filepath.ToSlash(parentPath)
	if !strings.HasSuffix(parentPath, "/") {
		parentPath += "/"
	}
	childPath, err := filepath.Abs(child)
	if err != nil {
		return
	}
	childPath = filepath.ToSlash(childPath)
	isSub = strings.HasPrefix(childPath, parentPath)
	return
}
