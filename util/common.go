package util

import (
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	"io"
	"strings"
)

func GetUuid() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// GetMd5 获取MD5
func GetMd5(str string) (res string) {
	m := md5.New()
	_, _ = io.WriteString(m, str)
	bs := m.Sum(nil)
	res = fmt.Sprintf("%x", bs)
	return
}
