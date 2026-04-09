package db

import (
	"reflect"
	"strings"
)

func NewConditions() *Conditions {
	cs := &Conditions{}
	return cs
}

type Conditions []*Condition

// Op 操作符类型
type Op string

const (
	Eq        Op = "="
	Ne        Op = "!="
	Gt        Op = ">"
	Gte       Op = ">="
	Lt        Op = "<"
	Lte       Op = "<="
	In        Op = "IN"
	NotIn     Op = "NOT IN"
	Like      Op = "LIKE"
	NotLike   Op = "NOT LIKE"
	IsNull    Op = "IS NULL"
	IsNotNull Op = "IS NOT NULL"
	Custom    Op = "CUSTOM"
)

// Condition 表示一个条件（字段 + 操作符 + 值）
type Condition struct {
	WhereSql string
	Field    string
	Op       Op
	Value    any
	// 用于嵌套的子条件（AND/OR 组）
	Children *Conditions
	IsOr     bool // true 表示 OR 组，false 表示 AND 组（仅在有Children时有效）
}

// And 添加 AND 条件（最常用）
func (cs *Conditions) And(field string, op Op, value any) *Conditions {
	*cs = append(*cs, &Condition{
		Field: field,
		Op:    op,
		Value: value,
	})
	return cs
}

// Or 添加 OR 条件（单条件）
func (cs *Conditions) Or(field string, op Op, value any) *Conditions {
	*cs = append(*cs, &Condition{
		Field: field,
		Op:    op,
		Value: value,
		IsOr:  true, // 标记为 OR（但实际拼接时需要看上下文）
	})
	return cs
}
func (cs *Conditions) AndWhereSql(whereSql string) *Conditions {
	*cs = append(*cs, &Condition{
		WhereSql: whereSql,
	})
	return cs
}
func (cs *Conditions) OrWhereSql(whereSql string) *Conditions {
	*cs = append(*cs, &Condition{
		WhereSql: whereSql,
		IsOr:     true, // 标记为 OR（但实际拼接时需要看上下文）
	})
	return cs
}

// AndGroup 添加一组 AND 条件（嵌套）
func (cs *Conditions) AndGroup(sub *Conditions) *Conditions {

	if sub == nil || len(*sub) == 0 {
		return cs
	}
	*cs = append(*cs, &Condition{
		Children: sub,
	})
	return cs
}

// OrGroup 添加一组 OR 条件（嵌套）
func (cs *Conditions) OrGroup(sub *Conditions) *Conditions {
	if sub == nil || len(*sub) == 0 {
		return cs
	}
	*cs = append(*cs, &Condition{
		Children: sub,
		IsOr:     true,
	})
	return cs
}

// Where 快捷方式，等价于 And(field, Eq, value)
func (cs *Conditions) Where(field string, value any) *Conditions {
	return cs.And(field, Eq, value)
}

// ------------------ 常用快捷方法 ------------------

func (cs *Conditions) Eq(field string, value any) *Conditions   { return cs.And(field, Eq, value) }
func (cs *Conditions) OrEq(field string, value any) *Conditions { return cs.Or(field, Eq, value) }

func (cs *Conditions) Ne(field string, value any) *Conditions   { return cs.And(field, Ne, value) }
func (cs *Conditions) OrNe(field string, value any) *Conditions { return cs.Or(field, Ne, value) }

func (cs *Conditions) Gt(field string, value any) *Conditions   { return cs.And(field, Gt, value) }
func (cs *Conditions) OrGt(field string, value any) *Conditions { return cs.Or(field, Gt, value) }

func (cs *Conditions) Gte(field string, value any) *Conditions   { return cs.And(field, Gte, value) }
func (cs *Conditions) OrGte(field string, value any) *Conditions { return cs.Or(field, Gte, value) }

func (cs *Conditions) Lt(field string, value any) *Conditions   { return cs.And(field, Lt, value) }
func (cs *Conditions) OrLt(field string, value any) *Conditions { return cs.Or(field, Lt, value) }

func (cs *Conditions) Lte(field string, value any) *Conditions   { return cs.And(field, Lte, value) }
func (cs *Conditions) OrLte(field string, value any) *Conditions { return cs.Or(field, Lte, value) }

func (cs *Conditions) In(field string, values any) *Conditions   { return cs.And(field, In, values) }
func (cs *Conditions) OrIn(field string, values any) *Conditions { return cs.Or(field, In, values) }

func (cs *Conditions) NotIn(field string, values any) *Conditions {
	return cs.And(field, NotIn, values)
}
func (cs *Conditions) OrNotIn(field string, values any) *Conditions {
	return cs.Or(field, NotIn, values)
}

func (cs *Conditions) Like(field string, pattern any) *Conditions {
	return cs.And(field, Like, pattern)
}
func (cs *Conditions) OrLike(field string, pattern any) *Conditions {
	return cs.Or(field, Like, pattern)
}

func (cs *Conditions) NotLike(field string, pattern any) *Conditions {
	return cs.And(field, NotLike, pattern)
}
func (cs *Conditions) OrNotLike(field string, pattern any) *Conditions {
	return cs.Or(field, NotLike, pattern)
}

func (cs *Conditions) IsNull(field string) *Conditions   { return cs.And(field, IsNull, nil) }
func (cs *Conditions) OrIsNull(field string) *Conditions { return cs.Or(field, IsNull, nil) }

func (cs *Conditions) IsNotNull(field string) *Conditions   { return cs.And(field, IsNotNull, nil) }
func (cs *Conditions) OrIsNotNull(field string) *Conditions { return cs.Or(field, IsNotNull, nil) }

func (cs *Conditions) Custom(custom string, values []any) *Conditions {
	return cs.And(custom, Custom, values)
}
func (cs *Conditions) OrCustom(custom string, values []any) *Conditions {
	return cs.Or(custom, Custom, values)
}

// Build 构建 WHERE 子句和参数
// 返回：WHERE 字符串（不含 "WHERE" 关键字），参数切片
func (cs *Conditions) Build(b *OrmSqlBuilder, s IService) (string, []any) {
	if cs == nil || len(*cs) == 0 {
		return "", nil
	}

	var sb strings.Builder
	var args []any

	for i, c := range *cs {
		if i > 0 {
			if c.IsOr {
				sb.WriteString(" OR ")
			} else {
				sb.WriteString(" AND ")
			}
		}
		cs.buildCriterion(b, &sb, c, &args, s)
	}

	return sb.String(), args
}

func (cs *Conditions) buildCriterion(b *OrmSqlBuilder, sb *strings.Builder, c *Condition, args *[]any, s IService) {
	if c.Children != nil && len(*c.Children) > 0 {
		// 嵌套组
		sb.WriteByte('(')
		for i, child := range *c.Children {
			if i > 0 {
				if c.IsOr {
					sb.WriteString(" OR ")
				} else {
					sb.WriteString(" AND ")
				}
			}
			cs.buildCriterion(b, sb, child, args, s)
		}
		sb.WriteByte(')')
		return
	}

	if c.WhereSql != "" {
		sb.WriteString(c.WhereSql)
		return
	}

	// 普通条件
	var wrapColumn = c.Field
	if b != nil {
		if c.Op != Custom {
			// 防注入，简单实现可替换成你的转义逻辑
			wrapColumn = b.WrapColumnName(b.sqlParam, c.Field)
		}
	}

	switch c.Op {
	case IsNull:
		sb.WriteString(wrapColumn)
		sb.WriteString(" IS NULL")
	case IsNotNull:
		sb.WriteString(wrapColumn)
		sb.WriteString(" IS NOT NULL")
	case Custom:
		sb.WriteString(c.Field)
		vals, ok := c.Value.([]any)
		if ok {
			*args = append(*args, vals...)
		}
	case In, NotIn:
		sb.WriteString(wrapColumn)
		sb.WriteByte(' ')
		sb.WriteString(string(c.Op))
		sb.WriteString(" (")
		vs := reflect.ValueOf(c.Value)
		size := vs.Len()
		for i := 0; i < size; i++ {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteByte('?')
			v := vs.Index(i)
			*args = append(*args, v.Interface())
		}

		sb.WriteByte(')')
	default:
		sb.WriteString(wrapColumn)
		sb.WriteByte(' ')
		sb.WriteString(string(c.Op))
		sqlCV, isSCV := c.Value.(*SqlConcatValue)
		if isSCV && sqlCV != nil && s != nil {
			sb.WriteString(" " + s.GetDialect().SqlConcat(sqlCV.Values...))
		} else {
			sb.WriteString(" ?")
		}
		if isSCV && sqlCV != nil {
			*args = append(*args, sqlCV.Value)
		} else {
			*args = append(*args, c.Value)
		}
	}
}

type SqlConcatValue struct {
	Values []string
	Value  any
}
