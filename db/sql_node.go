package db

import (
	"fmt"
	"github.com/team-ide/framework"
	"github.com/team-ide/framework/util"
	"reflect"
	"strings"
)

type SqlNode interface {
	isSqlNode()
	Append(b *SqlBuilder)
}

type SqlNodeSql struct {

	// 是 读取
	IsQuery bool `json:"isQuery,omitempty"`
	// 是 插入
	IsInsert bool `json:"isInsert,omitempty"`
	// 是 插入
	IsUpdate bool `json:"isUpdate,omitempty"`
	// 是 插入
	IsDelete bool `json:"isDelete,omitempty"`
	// 是 插入
	IsDDL bool `json:"isDDL,omitempty"`
	// 是 其它语句
	IsOther bool `json:"isOther,omitempty"`

	Root SqlNode `json:"root"`
}

func (this_ *SqlNodeSql) isSqlNode() {}
func (this_ *SqlNodeSql) Append(b *SqlBuilder) {
	this_.Root.Append(b)
}

type SqlNodeText struct {
	Text string
}

func (this_ *SqlNodeText) isSqlNode() {}
func (this_ *SqlNodeText) Append(b *SqlBuilder) {
	b.Append(this_.Text)
}

type SqlNodeComment struct {
	Comment string `json:"comment"`
}

func (this_ *SqlNodeComment) isSqlNode() {}
func (this_ *SqlNodeComment) Append(b *SqlBuilder) {
}

type SqlNodeBlock struct {
	Children []SqlNode `json:"children"`
}

func (this_ *SqlNodeBlock) isSqlNode() {}
func (this_ *SqlNodeBlock) Append(b *SqlBuilder) {
	for _, one := range this_.Children {
		one.Append(b)
	}
}

type SqlNodeSelect struct {
	*SqlNodeBlock
}

type SqlNodeInsert struct {
	*SqlNodeBlock
}

type SqlNodeUpdate struct {
	*SqlNodeBlock
}

type SqlNodeDelete struct {
	*SqlNodeBlock
}

type SqlNodeDDL struct {
	*SqlNodeBlock
}

type SqlNodeParam struct {
	Param SqlExpression `json:"param,omitempty"`
}

func (this_ *SqlNodeParam) isSqlNode() {}
func (this_ *SqlNodeParam) Append(b *SqlBuilder) {
	var v any
	if this_.Param != nil {
		v = this_.Param.GetValue(b)
	}
	b.Append(GetStringValue(v))
}

type SqlNodeArgParam struct {
	Param SqlExpression `json:"param,omitempty"`
}

func (this_ *SqlNodeArgParam) isSqlNode() {}
func (this_ *SqlNodeArgParam) Append(b *SqlBuilder) {
	var v any
	if this_.Param != nil {
		v = this_.Param.GetValue(b)
	}
	//fmt.Println("SqlNodeArgParam arg param name:", this_.Name)
	//fmt.Println("SqlNodeArgParam arg param value:", v)
	b.Append("?", v)
}

type SqlNodeOptional struct {
	Children []SqlNode `json:"children"`
	names    []string
}

func (this_ *SqlNodeOptional) isSqlNode() {}

func findSqlNodeIdentifier(list ...SqlNode) (names []string) {
	for _, one := range list {
		if one != nil {
			switch t := one.(type) {
			case *SqlNodeParam:
				names = append(names, findSqlExpressionIdentifier(t.Param)...)
			case *SqlNodeArgParam:
				names = append(names, findSqlExpressionIdentifier(t.Param)...)
			}
		}
	}
	return
}

func findSqlExpressionIdentifier(list ...SqlExpression) (names []string) {
	for _, one := range list {
		if one != nil {
			switch t := one.(type) {
			case *SqlNodeArithmeticOperate:
				names = append(names, findSqlExpressionIdentifier(t.Left, t.Right)...)
			case *SqlNodeRelationalOperate:
				names = append(names, findSqlExpressionIdentifier(t.Left, t.Right)...)
			case *SqlNodeLogicalOperate:
				names = append(names, findSqlExpressionIdentifier(t.Left, t.Right)...)
			case *SqlNodeTernary:
				names = append(names, findSqlExpressionIdentifier(t.Test, t.Left, t.Right)...)
			case *SqlNodeIdentifier:
				names = append(names, t.Name)
			}
		}
	}
	return
}
func (this_ *SqlNodeOptional) Append(b *SqlBuilder) {
	if this_.names == nil {
		this_.names = findSqlNodeIdentifier(this_.Children...)
	}

	var canAdd = func(name string) (res bool) {
		v := b.GetParam(name)
		if v == nil || v == "" || v == false {
			return false
		}
		return true
	}

	var add bool
	if len(this_.names) > 0 {
		for _, name := range this_.names {
			//fmt.Println("SqlNodeOptional find param name:", name)
			add = canAdd(name)
			//fmt.Println("SqlNodeOptional find param find:", add)
			if add {
				break
			}
		}
	} else {
		var str string
		for _, one := range this_.Children {
			switch t := one.(type) {
			case *SqlNodeText:
				str += " " + t.Text
			}
		}
		ss := strings.Split(str, " ")
		var name string
		for _, s := range ss {
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				continue
			}
			if len(name) == 0 {
				name = strings.ToLower(s)
			} else {
				name += strings.ToUpper(s[0:1]) + strings.ToLower(s[1:])
			}
		}
		//fmt.Println("SqlNodeOptional find param name:", name)
		add = canAdd(name)
		//fmt.Println("SqlNodeOptional find param find:", add)
	}
	if add {
		for _, c := range this_.Children {
			c.Append(b)
		}
	}
}

type SqlNodeIf struct {
	Test     SqlExpression    `json:"test"`
	Children []SqlNode        `json:"children"`
	ElseIfs  []*SqlNodeElseIf `json:"elseIfs"`
	Else     *SqlNodeElse     `json:"else"`
}

func (this_ *SqlNodeIf) isSqlNode() {}
func (this_ *SqlNodeIf) Append(b *SqlBuilder) {
	if b.Test(this_.Test) {
		for _, c := range this_.Children {
			c.Append(b)
		}
		return
	}
	for _, one := range this_.ElseIfs {
		if b.Test(one.Test) {
			for _, c := range one.Children {
				c.Append(b)
			}
			return
		}
	}
	if this_.Else != nil {
		for _, c := range this_.Else.Children {
			c.Append(b)
		}
		return
	}
}

type SqlNodeElseIf struct {
	Test     SqlExpression `json:"test"`
	Children []SqlNode     `json:"children"`
}

type SqlNodeElse struct {
	Children []SqlNode `json:"children"`
}

type SqlExpression interface {
	isSqlExpression()
	GetValue(b *SqlBuilder) (res any)
}

type SqlNodeParenthesis struct {
	Expression SqlExpression `json:"expression"`
}

func (this_ *SqlNodeParenthesis) isSqlExpression() {}
func (this_ *SqlNodeParenthesis) GetValue(b *SqlBuilder) (res any) {
	if this_.Expression == nil {
		return
	}
	res = this_.Expression.GetValue(b)
	return
}

type SqlNodeDotExpression struct {
	Left  SqlExpression `json:"left"`
	Right SqlExpression `json:"right"`
}

func (this_ *SqlNodeDotExpression) isSqlExpression() {}
func (this_ *SqlNodeDotExpression) GetObjectField(b *SqlBuilder, object any) (res any) {
	getObjectField, is := this_.Left.(IGetObjectField)
	if !is {
		framework.Error("sql dot expression left [" + reflect.TypeOf(this_.Right).String() + "] not to [GetObjectField] ")
		return
	}
	objectValue := getObjectField.GetObjectField(b, object)

	getObjectField, is = this_.Right.(IGetObjectField)
	if !is {
		framework.Error("sql dot expression right [" + reflect.TypeOf(this_.Right).String() + "] not to [GetObjectField] ")
		return
	}
	res = getObjectField.GetObjectField(b, objectValue)
	return
}

type IGetObjectField interface {
	GetObjectField(b *SqlBuilder, object any) (res any)
}

func (this_ *SqlNodeDotExpression) GetValue(b *SqlBuilder) (res any) {
	objectValue := b.GetValue(this_.Left)
	fmt.Println("SqlNodeDotExpression GetValue left:", util.GetStringValue(this_.Left))
	fmt.Println("SqlNodeDotExpression GetValue left value:", objectValue)
	if objectValue == nil {
		return
	}
	fmt.Println("SqlNodeDotExpression GetValue right:", util.GetStringValue(this_.Right))
	getObjectField, is := this_.Right.(IGetObjectField)
	if !is {
		framework.Error("sql dot expression right [" + reflect.TypeOf(this_.Right).String() + "] not to [GetObjectField] ")
		return
	}
	res = getObjectField.GetObjectField(b, objectValue)
	fmt.Println("SqlNodeDotExpression GetValue right value:", res)
	return
}

type SqlNodeIdentifier struct {
	Name string `json:"name"`
}

func (this_ *SqlNodeIdentifier) isSqlExpression() {}
func (this_ *SqlNodeIdentifier) GetObjectField(b *SqlBuilder, object any) (res any) {
	res = b.GetObjectField(object, this_.Name)
	return
}
func (this_ *SqlNodeIdentifier) GetValue(b *SqlBuilder) (res any) {
	//fmt.Println("SqlNodeIdentifier name:", this_.Name)
	switch strings.ToLower(this_.Name) {
	case "true":
		return true
	case "false":
		return false
	case "null", "nil":
		return nil
	}
	//if this_.Name == "autoIncrement" {
	//	for k, v := range b.Param {
	//		fmt.Println("key:", k)
	//		fmt.Println("key v:", v)
	//	}
	//}
	res = b.GetParam(this_.Name)
	//fmt.Println("SqlNodeIdentifier value:", res)
	return
}

type SqlNodeInt struct {
	Value  int64  `json:"value"`
	String string `json:"string"`
}

func (this_ *SqlNodeInt) isSqlNode() {}
func (this_ *SqlNodeInt) Append(b *SqlBuilder) {
	b.Append(this_.String)
}
func (this_ *SqlNodeInt) isSqlExpression() {}
func (this_ *SqlNodeInt) GetValue(b *SqlBuilder) (res any) {
	res = this_.Value
	return
}

type SqlNodeFloat struct {
	Value  float64 `json:"value"`
	String string  `json:"string"`
}

func (this_ *SqlNodeFloat) isSqlNode() {}
func (this_ *SqlNodeFloat) Append(b *SqlBuilder) {
	b.Append(this_.String)
}
func (this_ *SqlNodeFloat) isSqlExpression() {}
func (this_ *SqlNodeFloat) GetValue(b *SqlBuilder) (res any) {
	res = this_.Value
	return
}

type SqlNodeString struct {
	Value string `json:"value"`
}

func (this_ *SqlNodeString) isSqlNode() {}
func (this_ *SqlNodeString) Append(b *SqlBuilder) {
	b.Append(this_.Value)
}
func (this_ *SqlNodeString) isSqlExpression() {}
func (this_ *SqlNodeString) GetValue(b *SqlBuilder) (res any) {
	res = this_.Value
	return
}

// SqlNodeTernary 三元运算
type SqlNodeTernary struct {
	Test  SqlExpression `json:"test"`
	Left  SqlExpression `json:"left"`
	Right SqlExpression `json:"right"`
}

func (this_ *SqlNodeTernary) isSqlExpression() {}
func (this_ *SqlNodeTernary) GetValue(b *SqlBuilder) (res any) {
	if b.Test(this_.Test) {
		res = b.GetValue(this_.Left)
	} else {
		res = b.GetValue(this_.Right)
	}
	return
}

// SqlNodeArithmeticOperate 算术运算: + - * / %
type SqlNodeArithmeticOperate struct {
	Left SqlExpression `json:"left"`
	// 算术运算符: + - * / %
	Operator string        `json:"operator"`
	Right    SqlExpression `json:"right"`
}

func (this_ *SqlNodeArithmeticOperate) isSqlNode()       {}
func (this_ *SqlNodeArithmeticOperate) isSqlExpression() {}
func (this_ *SqlNodeArithmeticOperate) GetValue(b *SqlBuilder) (res any) {
	var leftV = b.GetValue(this_.Left)
	var rightV = b.GetValue(this_.Right)

	if this_.Operator == "+" {
		var leftIsS bool
		if leftV != nil {
			_, leftIsS = leftV.(string)
		}
		var rightIsS bool
		if rightV != nil {
			_, rightIsS = rightV.(string)
		}
		if (leftV == nil || leftIsS) || rightIsS {
			res = GetStringValue(leftV) + GetStringValue(rightV)
			return
		}
	}
	leftInt, e := ToInt64Value(leftV)
	if e != nil {
		fmt.Println(fmt.Sprintf("binary left value:%v 无法转为int64:"+e.Error(), leftV))
		return
	}
	rightInt, e := ToInt64Value(rightV)
	if e != nil {
		fmt.Println(fmt.Sprintf("binary right value:%v 无法转为int64:"+e.Error(), rightV))
		return
	}
	switch this_.Operator {
	case "+":
		res = leftInt + rightInt
	case "-":
		res = leftInt - rightInt
	case "*":
		res = leftInt * rightInt
	case "/":
		res = leftInt / rightInt
	case "%":
		res = leftInt % rightInt
	}
	return
}

// SqlNodeRelationalOperate 比较（关系）运算: == != > < >= <=
type SqlNodeRelationalOperate struct {
	Left SqlExpression `json:"left"`
	// 比较（关系）运算符: == != > < >= <=
	Operator string        `json:"operator"`
	Right    SqlExpression `json:"right"`
}

func (this_ *SqlNodeRelationalOperate) isSqlNode()       {}
func (this_ *SqlNodeRelationalOperate) isSqlExpression() {}
func (this_ *SqlNodeRelationalOperate) GetValue(b *SqlBuilder) (res any) {
	var leftV = b.GetValue(this_.Left)
	var rightV = b.GetValue(this_.Right)
	if this_.Operator == "==" {
		if reflect.TypeOf(leftV) == reflect.TypeOf(rightV) {
			if leftV == rightV {
				return true
			}
		}
		if GetStringValue(leftV) == GetStringValue(rightV) {
			return true
		}
		return false
	}
	if this_.Operator == "!=" {
		if reflect.TypeOf(leftV) == reflect.TypeOf(rightV) {
			if leftV != rightV {
				return true
			}
		}
		leftS := GetStringValue(leftV)
		rightS := GetStringValue(rightV)
		//fmt.Println("SqlNodeBinary Operator:", this_.Operator)
		//fmt.Println("SqlNodeBinary leftV:", leftV)
		//fmt.Println("SqlNodeBinary rightV:", rightV)
		//fmt.Println("SqlNodeBinary leftS:", leftS)
		//fmt.Println("SqlNodeBinary rightS:", rightS)
		//fmt.Println("SqlNodeBinary leftS != rightS:", leftS != rightS)

		if leftS != rightS {
			return true
		}
		return false
	}
	leftInt, e := ToInt64Value(leftV)
	if e != nil {
		fmt.Println(fmt.Sprintf("binary left value:%v 无法转为int64:"+e.Error(), leftV))
		return
	}
	rightInt, e := ToInt64Value(rightV)
	if e != nil {
		fmt.Println(fmt.Sprintf("binary right value:%v 无法转为int64:"+e.Error(), rightV))
		return
	}
	switch this_.Operator {
	case "<":
		res = leftInt < rightInt
	case ">":
		res = leftInt > rightInt
	case "<=":
		res = leftInt <= rightInt
	case ">=":
		res = leftInt >= rightInt
	}
	return
}

// SqlNodeLogicalOperate 逻辑运算: && ||
type SqlNodeLogicalOperate struct {
	Left SqlExpression `json:"left"`
	// 逻辑运算符: && ||
	Operator string        `json:"operator"`
	Right    SqlExpression `json:"right"`
}

func (this_ *SqlNodeLogicalOperate) isSqlNode()       {}
func (this_ *SqlNodeLogicalOperate) isSqlExpression() {}
func (this_ *SqlNodeLogicalOperate) GetValue(b *SqlBuilder) (res any) {
	var leftV = b.GetValue(this_.Left)
	if this_.Operator == "||" {
		if IsTrue(leftV) {
			return true
		}
		var rightV = b.GetValue(this_.Right)
		if IsTrue(rightV) {
			return true
		}
		return false
	}
	if this_.Operator == "&&" {
		if !IsTrue(leftV) {
			return false
		}
		var rightV = b.GetValue(this_.Right)
		if !IsTrue(rightV) {
			return false
		}
		return true
	}
	return
}
