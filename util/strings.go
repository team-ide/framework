package util

import (
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

// FirstToUpper 字符首字母大写
// @param str string "任意字符串"
// @return string
// FirstToUpper("abc")
func FirstToUpper(str string) (res string) {
	if str == "" {
		return
	}
	res = strings.ToUpper(str[0:1])
	res += str[1:]
	return
}

// FirstToLower 字符首字母小写
// @param str string "任意字符串"
// @return string
// FirstToLower("Abc")
func FirstToLower(str string) (res string) {
	if str == "" {
		return
	}
	res = strings.ToLower(str[0:1])
	res += str[1:]
	return
}

// Marshal 转换为大驼峰命名法则 首字母大写，“_” 忽略后大写
// Marshal("abc_def")
func Marshal(name string) string {
	if name == "" {
		return ""
	}

	temp := strings.Split(name, "_")
	var s string
	for _, v := range temp {
		vv := []rune(v)
		if len(vv) > 0 {
			if vv[0] >= 'a' && vv[0] <= 'z' { //首字母大写
				vv[0] -= 32
			}
			s += string(vv)
		}
	}

	return s
}

// Hump 转换为驼峰命名法则 “_”后的字母大写
// Hump("abc_def")
func Hump(name string) string {
	if name == "" {
		return ""
	}

	temp := strings.Split(name, "_")
	var s string
	for i, v := range temp {
		vv := []rune(v)
		if len(vv) > 0 {
			if i > 0 {
				if vv[0] >= 'a' && vv[0] <= 'z' { //首字母大写
					vv[0] -= 32
				}
			}
			s += string(vv)
		}
	}

	return s
}

// GetStringValue 将传入的值转为字符串
// @param value interface{} "任意值"
// @return string
// GetStringValue(arg)
func GetStringValue(value any) (valueString string) {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		valueString = v
		break
	case *string:
		valueString = *v
		break
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		valueString = fmt.Sprintf("%d", v)
		break
	case float32:
		valueString = decimal.NewFromFloat32(v).String()
		break
	case float64:
		valueString = decimal.NewFromFloat(v).String()
	case bool:
		if v {
			valueString = "1"
		} else {
			valueString = "0"
		}
		break
	case time.Time:
		if v.IsZero() {
			valueString = ""
		} else {
			//valueString = GetFormatByTime(v)
			valueString = v.Format("2006-01-02 15:04:05.0000000-07:00")
		}
		break
	case []byte:
		valueString = string(v)
		break
	case StringType:
		valueString = v.String()
	default:
		s, err := ObjToJson(value)
		if err != nil {
			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString = s
		}
		break
	}
	return
}

type StringType interface {
	String() string
}

// ToPinYin 将姓名转为拼音
// @param name string "姓名"
// @return string
// ToPinYin("张三")
//func ToPinYin(name string) (res string, err error) {
//	// InitialsInCapitals: 首字母大写, 不带音调
//	// WithoutTone: 全小写,不带音调
//	// Tone: 全小写带音调
//	res, err = pinyin.New(name).Split("").Mode(pinyin.WithoutTone).Convert()
//	if err != nil {
//		return
//	}
//	return
//}

var (
	RandChats = []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "c", "d", "e", "f", "g",
		"h", "i", "j", "k", "l", "m", "n",
		"o", "p", "q", "r", "s", "t", "u",
		"v", "w", "z", "y", "z",
		"A", "B", "C", "D", "E", "F", "G",
		"H", "I", "J", "K", "L", "M", "N",
		"O", "P", "Q", "R", "S", "T", "U",
		"V", "W", "Z", "Y", "Z",
		"_",
	}
	RandChatsSize = len(RandChats)
)

// StrPadLeft 在字符串 左侧补全 字符串 到 指定长度
// input string 原字符串
// padLength int 规定补齐后的字符串长度
// padString string 自定义填充字符串
// StrPadLeft("xx", 5, "0") 左侧补”0“达到5位长度
func StrPadLeft(input string, padLength int, padString string) string {

	output := ""
	inputLen := len(input)

	if inputLen >= padLength {
		return input
	}

	padStringLen := len(padString)
	needFillLen := padLength - inputLen

	if diffLen := padStringLen - needFillLen; diffLen > 0 {
		padString = padString[diffLen:]
	}

	for i := 1; i <= needFillLen; i += padStringLen {
		output += padString
	}
	return output + input
}

// StrPadRight 在字符串 右侧补全 字符串 到 指定长度
// input string 原字符串
// padLength int 规定补齐后的字符串长度
// padString string 自定义填充字符串
// StrPadRight("xx", 5, "0") 右侧补”0“达到5位长度
func StrPadRight(input string, padLength int, padString string) string {

	output := ""
	inputLen := len(input)

	if inputLen >= padLength {
		return input
	}

	padStringLen := len(padString)
	needFillLen := padLength - inputLen

	if diffLen := padStringLen - needFillLen; diffLen > 0 {
		padString = padString[diffLen:]
	}

	for i := 1; i <= needFillLen; i += padStringLen {
		output += padString
	}
	return input + output
}

// TrimSpace 去除 前后空格
func TrimSpace(arg string) string {
	return strings.TrimSpace(arg)
}

// TrimPrefix 去除 匹配的 前缀
func TrimPrefix(arg string, trim string) string {
	return strings.TrimPrefix(arg, trim)
}

// HasPrefix 匹配的 前缀
func HasPrefix(arg string, trim string) bool {
	return strings.HasPrefix(arg, trim)
}

// TrimSuffix 去除 匹配的 后缀
func TrimSuffix(arg string, trim string) string {
	return strings.TrimSuffix(arg, trim)
}

// HasSuffix 匹配的 后缀
func HasSuffix(arg string, trim string) bool {
	return strings.HasSuffix(arg, trim)
}

// TrimLeft 去除 所有 匹配的 前缀
func TrimLeft(arg string, trim string) string {
	return strings.TrimLeft(arg, trim)
}

// TrimRight 去除 所有 匹配的 后缀
func TrimRight(arg string, trim string) string {
	return strings.TrimRight(arg, trim)
}

// StringJoin 字符串拼接
func StringJoin(es []string, sep string) string {
	return strings.Join(es, sep)
}

// AnyJoin 任意切片拼接
func AnyJoin(sep string, es ...any) (res string) {
	if len(es) == 0 {
		return
	}
	for i, e := range es {
		if i > 0 {
			res += sep
		}
		res += GetStringValue(e)
	}
	return
}

// IntJoin int 拼接
func IntJoin(es []int, sep string) (res string) {
	if len(es) == 0 {
		return
	}
	for i, e := range es {
		if i > 0 {
			res += sep
		}
		res += fmt.Sprintf("%d", e)
	}
	return
}

// Int64Join int64 拼接
func Int64Join(es []int64, sep string) (res string) {
	if len(es) == 0 {
		return
	}
	for i, e := range es {
		if i > 0 {
			res += sep
		}
		res += fmt.Sprintf("%d", e)
	}
	return
}

// GenStringJoin 生成 字符串 拼接
// GenStringJoin(5, "xx", ",") 表示 生成 xx,xx,xx,xx,xx
func GenStringJoin(len int, str string, sep string) (res string) {
	if len <= 0 {
		return
	}
	for i := 0; i < len; i++ {
		if i > 0 {
			res += sep
		}
		res += str
	}
	return
}
