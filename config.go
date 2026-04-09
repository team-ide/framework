package framework

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/team-ide/framework/util"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
)

func ReadConfig[C any](conf string) (cfg C, err error) {
	cType := reflect.TypeOf(cfg)
	isPtr := cType.Kind() == reflect.Ptr
	if isPtr {
		cType = cType.Elem()
	}
	cValue := reflect.New(cType)
	if isPtr {
		cfg = cValue.Interface().(C)
	} else {
		cfg = cValue.Elem().Interface().(C)
	}

	Info("read config [" + conf + "] start")
	var exists bool
	exists, err = util.PathExists(conf)
	if err != nil {
		Error("配置文件检查异常:" + err.Error())
		return
	}
	if !exists {
		err = errors.New(fmt.Sprint("配置文件[", conf, "]不存在"))
		Error(err.Error())
		return
	}
	var f *os.File
	f, err = os.Open(conf)
	if err != nil {
		return
	}
	bs, err := io.ReadAll(f)
	if err != nil {
		Error("配置文件[" + conf + "]读取异常:" + err.Error())
		return
	}
	var configMap = map[string]any{}
	if strings.HasSuffix(conf, ".json") {
		err = json.Unmarshal(bs, &configMap)
	} else if strings.HasSuffix(conf, ".yml") || strings.HasSuffix(conf, ".yaml") {
		err = yaml.Unmarshal(bs, &configMap)
	} else if strings.HasSuffix(conf, ".toml") {
		err = toml.Unmarshal(bs, &configMap)
	}
	if err != nil {
		Error("配置文件[" + conf + "] 加载异常:" + err.Error())
		return
	}
	//fmt.Println("format config before:" + util.GetStringValue(configMap))
	FormatConfigMap(configMap, cType)
	//fmt.Println("format config after:" + util.GetStringValue(configMap))
	err = util.ObjToObjByJson(configMap, cfg)
	if err != nil {
		Error("config map data to config error:" + err.Error())
		return
	}
	Info("read config [" + conf + "] success")
	return
}

func FormatConfigMap(mapValue map[string]any, mapType reflect.Type) {
	if mapValue == nil {
		return
	}
	var filedMap = make(map[string]reflect.StructField)
	var jsonMap = make(map[string]reflect.StructField)
	var yamlMap = make(map[string]reflect.StructField)
	var tomlMap = make(map[string]reflect.StructField)
	var appendField func(t reflect.Type)
	appendField = func(t reflect.Type) {
		for t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for num := range t.NumField() {
			field := t.Field(num)
			if field.Anonymous {
				appendField(field.Type)
				continue
			}
			filedMap[strings.ToLower(field.Name)] = field
			jsonName := field.Tag.Get("json")
			if strings.Contains(jsonName, ",") {
				jsonName = strings.TrimSpace(strings.Split(jsonName, ",")[0])
			}
			jsonMap[strings.ToLower(jsonName)] = field

			yamlName := field.Tag.Get("yaml")
			if strings.Contains(yamlName, ",") {
				yamlName = strings.TrimSpace(strings.Split(yamlName, ",")[0])
			}
			yamlMap[strings.ToLower(yamlName)] = field

			tomlName := field.Tag.Get("toml")
			if strings.Contains(tomlName, ",") {
				tomlName = strings.TrimSpace(strings.Split(tomlName, ",")[0])
			}
			tomlMap[strings.ToLower(tomlName)] = field
		}
	}
	if mapType != nil {
		appendField(mapType)
	}

	for key, value := range mapValue {
		field, find := filedMap[strings.ToLower(key)]
		if !find {
			field, find = yamlMap[strings.ToLower(key)]
		}
		if !find {
			field, find = tomlMap[strings.ToLower(key)]
		}
		if !find {
			field, find = jsonMap[strings.ToLower(key)]
		}
		var fieldType reflect.Type
		if find {
			fieldType = field.Type
		}

		switch v := value.(type) {
		case map[string]any:
			FormatConfigMap(v, fieldType)
		case []any:
			for i, one := range v {
				switch oneV := one.(type) {
				case map[string]any:
					FormatConfigMap(oneV, fieldType)
					v[i] = oneV
				default:
					res := getEnvValue(oneV)
					if find {
						v[i] = valueToTypeValue(res, fieldType)
					} else {
						v[i] = res
					}
				}
			}
		default:
			res := getEnvValue(value)
			if find {
				mapValue[key] = valueToTypeValue(res, fieldType)
			} else {
				mapValue[key] = res
			}
		}
	}

}

func valueToTypeValue(v any, t reflect.Type) (res any) {
	res = v
	if t.Kind() == reflect.String {
		res = util.GetStringValue(v)
		return
	}
	if t.Kind() >= reflect.Int && t.Kind() <= reflect.Int64 {
		res = util.StringToInt64(util.GetStringValue(v))
		return
	}
	if t.Kind() >= reflect.Uint && t.Kind() <= reflect.Uint64 {
		res = util.StringToUint64(util.GetStringValue(v))
		return
	}
	if t.Kind() == reflect.Bool {
		str := util.GetStringValue(v)
		res = str == "true" || str == "1" || str == "t" || str == "yes" || str == "y" || str == "on"
		return
	}
	return
}

var (
	configValueReg, _ = regexp.Compile(`[$]+{(.+?)}`)
)

func getEnvValue(value any) (v any) {
	if value == nil {
		return
	}
	stringValue, stringValueOk := value.(string)
	if !stringValueOk {
		v = value
		return
	}
	var formatValue string
	indexList := configValueReg.FindAllIndex([]byte(stringValue), -1)
	var lastIndex = 0
	for _, indexes := range indexList {
		formatValue += stringValue[lastIndex:indexes[0]]

		lastIndex = indexes[1]

		str := stringValue[indexes[0]+2 : indexes[1]-1]
		var envName = str
		var defaultValue string
		idx := strings.Index(str, ":")
		if idx > 0 {
			envName = str[:idx]
			defaultValue = str[idx+1:]
		}
		envName = strings.TrimSpace(envName)
		defaultValue = strings.TrimSpace(defaultValue)

		envValue := GetFromSystem(envName)
		if envValue == "" {
			envValue = defaultValue
		}
		formatValue += envValue
	}
	formatValue += stringValue[lastIndex:]

	v = formatValue
	return
}

func GetFromSystem(key string) string {
	return os.Getenv(key)
}
