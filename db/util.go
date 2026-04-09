package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type DmBlobType interface {
	Read(dest []byte) (n int, err error)
	ReadAt(pos int, dest []byte) (n int, err error)
	Write(pos int, src []byte) (n int, err error)
	Truncate(length int64) error
	Scan(src interface{}) error
}

type DmClobType interface {
	GetLength() (int64, error)
	ReadString(pos int, length int) (result string, err error)
	WriteString(pos int, s string) (n int, err error)
	Truncate(length int64) error
	Scan(src interface{}) error
}

type HasStringFuncType interface {
	String() string
}

func GetSqlValue(columnType *sql.ColumnType, data any) (value any) {
	if data == nil {
		return
	}
	if dmB, isDMB := data.(DmBlobType); isDMB {
		var bs []byte
		var readBs = make([]byte, 1024*1024)
		for {
			n, err := dmB.Read(readBs)
			if n > 0 {
				bs = append(bs, readBs[0:0]...)
			}
			if err != nil {
				if err == io.EOF {
					break
				} else {
					panic("GetSqlValue DmBlob Read error:" + err.Error())
				}
			}
		}
		value = string(bs)
		return
	}
	if dmC, isDMC := data.(DmClobType); isDMC {
		l, err := dmC.GetLength()
		if err != nil {
			panic("GetSqlValue DmClob GetLength error:" + err.Error())
			return
		}
		if l > 0 {
			value, err = dmC.ReadString(1, int(l))
			if err != nil {
				panic("GetSqlValue DmClob ReadString error:" + err.Error())
				return
			}
		}
		return
	}
	if sFun, isSFun := data.(HasStringFuncType); isSFun {
		value = sFun.String()
		return
	}
	vOf := reflect.ValueOf(data)
	if vOf.Kind() == reflect.Ptr {
		if vOf.IsNil() {
			return nil
		}
		return GetSqlValue(columnType, vOf.Elem().Interface())
	}
	//if columnType.Name() == "NESTING_EVENT_TYPE" {
	//	fmt.Println("NESTING_EVENT_TYPE value type", reflect.TypeOf(data).String(), " value is ", data)
	//}
	switch v := data.(type) {
	case sql.NullString:
		if !v.Valid {
			return nil
		}
		value = (v).String
		break
	case sql.NullBool:
		if !v.Valid {
			return nil
		}
		value = (v).Bool
		break
	case sql.NullByte:
		if !v.Valid {
			return nil
		}
		value = (v).Byte
		break
	case sql.NullFloat64:
		if !v.Valid {
			return nil
		}
		value = (v).Float64
		break
	case sql.NullInt16:
		if !v.Valid {
			return nil
		}
		value = (v).Int16
		break
	case sql.NullInt32:
		if !v.Valid {
			return nil
		}
		value = (v).Int32
		break
	case sql.NullInt64:
		if !v.Valid {
			return nil
		}
		value = (v).Int64
		break
	case sql.NullTime:
		if !v.Valid {
			return nil
		}
		value = (v).Time
		break
	case sql.RawBytes:
		value = string(v)
		break
	case string, int, int8, int16, int32, int64, float32, float64, bool, uint, uint8, uint16, uint32, uint64:
		value = v
		break
	case []byte:
		value = string(v)
		break
	case time.Time:
		value = v
		break
	default:
		baseValue, isBaseType := GetBaseTypeValue(value)
		if isBaseType {
			value = baseValue
			return
		}
		value = v
		//panic("GetSqlValue data [" + fmt.Sprint(data) + "] data type [" + reflect.TypeOf(data).String() + "] name [" + columnType.Name() + "] databaseType [" + columnType.DatabaseTypeName() + "] not support")
		break
	}
	return
}

func GetBaseTypeValue(data any) (res any, is bool) {
	if data == nil {
		return
	}
	switch v := data.(type) {
	case string, int, int8, int16, int32, int64, float32, float64, bool, uint, uint8, uint16, uint32, uint64, []byte:
		res = v
		is = true
		return
	}
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		if dataValue.IsNil() {
			return
		}
		return GetBaseTypeValue(dataValue.Elem().Interface())
	}

	res = dataValue.Interface()
	return
}

func FormatDriverDSN(driverDSN string, dbCfg *Config, isUrl bool, paramSpaceChar string, canPathEscape bool, defaultParams map[string]string, params map[string]string) string {
	if paramSpaceChar == "" {
		paramSpaceChar = " "
	}
	var ps = make(map[string]string)
	if defaultParams != nil {
		for k, v := range defaultParams {
			ps[k] = v
		}
	}
	if params != nil {
		for k, v := range params {
			ps[k] = v
		}
	}
	var dsnParams = map[string]string{}
	for k, v := range ps {
		var pv = v
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			pv = GetParamValue(dbCfg, strings.TrimSpace(v[1:len(v)-1]))
			if pv == "" {
				continue
			}
		}
		if canPathEscape {
			pv = url.PathEscape(pv)
		}
		dsnParams[k] = pv
	}
	formatStr, names := ParserToFormat(driverDSN)
	//fmt.Println("GetDriverDSN formatStr:", formatStr)
	//fmt.Println("GetDriverDSN names:", names)
	var dsn = formatStr
	for i, name := range names {
		var argChar = fmt.Sprintf("{%d}", i)
		var argValue = GetParamValue(dbCfg, name)
		if canPathEscape {
			argValue = url.PathEscape(argValue)
		}
		dsn = strings.ReplaceAll(dsn, argChar, argValue)
	}
	var appendParams string
	if isUrl {
		appendParams = ToUrlParams(dsnParams)
		if appendParams != "" {
			if strings.Contains(dsn, "?") {
				dsn += appendParams
			} else {
				dsn += "?" + appendParams
			}
		}
	} else {
		appendParams = ToKeyValueParams(dsnParams, paramSpaceChar)
		if appendParams != "" {
			dsn += paramSpaceChar + appendParams
		}

	}
	if dbCfg.DsnAppend != "" {
		dsn += dbCfg.DsnAppend
	}
	//fmt.Println("GetDriverDSN dsn:", dsn)
	return dsn
}

func ToUrlParams(params map[string]string) (res string) {
	if params == nil {
		return
	}
	for k, v := range params {
		res += fmt.Sprintf("&%s=%s", k, v)
	}
	return
}

func ToKeyValueParams(params map[string]string, paramSpaceChar string) (res string) {
	if params == nil {
		return
	}
	for k, v := range params {
		res += fmt.Sprintf(paramSpaceChar+"%s=%s", k, v)
	}
	return
}

func GetParamValue(dbCfg *Config, name string) string {
	switch name {
	case "type":
		return dbCfg.Type
	case "dialectType":
		return dbCfg.DialectType
	case "user":
		return dbCfg.Username
	case "password":
		return dbCfg.Password
	case "host":
		return dbCfg.Host
	case "port":
		return fmt.Sprintf("%d", dbCfg.Port)
	case "database":
		return dbCfg.Database
	case "schema":
		return dbCfg.Schema
	case "databasePath":
		if dbCfg.getDatabasePath != nil {
			return dbCfg.getDatabasePath(dbCfg)
		}
		if dbCfg.DatabasePath != "" {
			return dbCfg.DatabasePath
		}
		return dbCfg.Database
	case "serverName":
		return dbCfg.ServerName
	case "address":
		if dbCfg.Address != "" {
			return dbCfg.Address
		}
		if dbCfg.Port != 0 {
			return fmt.Sprintf("%s:%d", dbCfg.Host, dbCfg.Port)
		}
		return dbCfg.Host
	}
	return ""
}

// ParserToFormat 解析 字符串参数 将 `xx{name}xx{xx}` 转为 `xx{0}xx{1}`
// 如：{user}:{password}@tcp({host}:{port})/{database}?charset=utf8&parseTime=true&{urlParams}
func ParserToFormat(str string) (formatStr string, names []string) {
	var result []byte
	i := 0
	for i < len(str) {
		if str[i] == '{' {
			// 找到参数名的开始
			nameStart := i + 1
			nameEnd := nameStart

			// 查找结束的 }
			for nameEnd < len(str) && str[nameEnd] != '}' {
				nameEnd++
			}

			if nameEnd < len(str) && str[nameEnd] == '}' {
				// 提取参数名
				name := str[nameStart:nameEnd]

				// 检查参数名是否已经存在
				index := len(names)
				names = append(names, name)
				// 替换为 {index}
				result = append(result, '{')
				result = append(result, strconv.Itoa(index)...)
				result = append(result, '}')

				// 跳过已处理的部分
				i = nameEnd + 1
				continue
			}
		}

		// 普通字符或未匹配的 {，直接添加
		result = append(result, str[i])
		i++
	}

	formatStr = string(result)
	return
}

func IsInt64Value(value any) (res int64, ok bool) {
	if value == nil {
		return
	}
	ok = true
	switch t := value.(type) {
	case int64:
		res = t
	case int:
		res = int64(t)
	case int8:
		res = int64(t)
	case int16:
		res = int64(t)
	case int32:
		res = int64(t)
	case uint:
		res = int64(t)
	case uint8:
		res = int64(t)
	case uint16:
		res = int64(t)
	case uint32:
		res = int64(t)
	case uint64:
		res = int64(t)
	default:
		ok = false
	}
	return
}

func ToInt64Value(value any) (res int64, err error) {
	res, ok := IsInt64Value(value)
	if ok {
		return
	}
	stringV := GetStringValue(value)
	if stringV == "" {
		return
	}
	res, err = strconv.ParseInt(stringV, 10, 64)
	return
}

func ToIntValue(value any) (res int, err error) {
	i, err := ToInt64Value(value)
	if err != nil {
		return
	}
	res = int(i)
	return
}

func GetStringValue(value any) (valueString string) {
	if value == nil {
		return
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
		bs, err := json.Marshal(v)
		if err == nil {
			valueString = string(bs)
		} else {
			valueString = fmt.Sprintf("%v", value)
		}
		break
	}
	return
}

type StringType interface {
	String() string
}

func IsTrue(arg any) bool {
	if arg == nil {
		return false
	}
	switch v := arg.(type) {
	case bool:
		return v
	case string:
		str := strings.ToLower(strings.TrimSpace(v))
		return str == "true" || str == "1" || str == "t" || str == "yes" || str == "y" || str == "on"
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	case complex64, complex128:
		return reflect.ValueOf(v).Complex() != complex(0, 0)
	}
	return false
}

func FormatColumnType(columnType string) (dataType string, length int, precision int, scale int) {
	index := strings.Index(columnType, "(")
	if index <= 0 {
		dataType = columnType
		return
	}
	dataType = columnType[:index]
	lengthStr := columnType[index+1 : strings.Index(columnType, ")")]
	var vs []int
	ss := strings.Split(lengthStr, ",")
	vs = make([]int, len(ss))
	for i, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		vs[i], _ = strconv.Atoi(s)
	}
	if len(vs) == 1 {
		length = vs[0]
	} else if len(vs) == 2 {
		precision = vs[0]
		scale = vs[1]
	} else if len(vs) == 3 {
		length = vs[0]
		precision = vs[1]
		scale = vs[2]
	}
	return
}
