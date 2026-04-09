package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// StringInt64 自定义 int64 类型 序列化JSON为string，反序列化支持string、int64输入
type StringInt64 int64

// MarshalJSON 将 int64 序列化为字符串
func (s StringInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(s), 10))
}

// UnmarshalJSON 反序列化，支持字符串和数字
func (s *StringInt64) UnmarshalJSON(b []byte) error {
	var value interface{}
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64: // JSON 数字默认解析为 float64
		*s = StringInt64(int64(v))
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		*s = StringInt64(i)
	default:
		return fmt.Errorf("invalid type for StringInt64: %T", value)
	}
	return nil
}

// StringInt64Slice 自定义 []int64 类型
type StringInt64Slice []int64

// MarshalJSON 将 []int64 序列化为字符串数组
func (s StringInt64Slice) MarshalJSON() ([]byte, error) {
	strSlice := make([]string, len(s))
	for i, v := range s {
		strSlice[i] = strconv.FormatInt(v, 10)
	}
	return json.Marshal(strSlice)
}

// UnmarshalJSON 反序列化，支持字符串数组和数字数组
func (s *StringInt64Slice) UnmarshalJSON(b []byte) error {
	var value interface{}
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case []interface{}: // JSON 数组
		intSlice := make([]int64, len(v))
		for i, item := range v {
			switch itemVal := item.(type) {
			case float64: // 数字
				intSlice[i] = int64(itemVal)
			case string: // 字符串
				num, err := strconv.ParseInt(itemVal, 10, 64)
				if err != nil {
					return err
				}
				intSlice[i] = num
			default:
				return fmt.Errorf("invalid type in array for StringInt64Slice: %T", item)
			}
		}
		*s = intSlice
	default:
		return fmt.Errorf("invalid type for StringInt64Slice: %T", value)
	}
	return nil
}
