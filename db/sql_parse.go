package db

import (
	"errors"
	"strconv"
	"strings"
)

func (this_ *SqlTemplate) Parse() (err error) {
	//fmt.Println("parse sql:" + this_.GetTemplateSql())
	this_.SqlList, err = this_.ParseSqlList(this_.GetTemplateSql())
	if err != nil {
		return
	}

	return
}
func (this_ *SqlTemplate) GetTemplateSql() (str string) {

	for _, sqlInfo := range this_.sqlList {
		if sqlInfo == "" {
			continue
		}
		str += " " + sqlInfo
	}

	return
}

func (this_ *SqlTemplate) ParseSqlList(str string) (res []*SqlNodeSql, err error) {
	r := NewSqlReader(str)
	r.NextHasChar()

	for !r.isEnd {
		if r.str == ";" {
			r.NextHasChar()
			continue
		}
		parseSql := &SqlNodeSql{}
		parseSql, err = this_.ParseSql(r)
		if err != nil {
			return
		}
		res = append(res, parseSql)
		//fmt.Println("ParseSqlList Next Sql reader ReadState:", r.ReadState)
		//fmt.Println("ParseSqlList Next Sql reader ReadScope:", r.ReadScope)
	}
	return
}
func (this_ *SqlTemplate) ParseSql(r *SqlReader) (res *SqlNodeSql, err error) {

	res = &SqlNodeSql{}
	switch strings.ToLower(r.str) {
	case "select", "show", "desc", "explain", "pragma", "with":
		res.IsQuery = true
		res.Root, err = this_.parseSelect(r)
	case "insert":
		res.IsInsert = true
		res.Root, err = this_.parseInsert(r)
	case "update":
		res.IsUpdate = true
		res.Root, err = this_.parseUpdate(r)
	case "delete":
		res.IsDelete = true
		res.Root, err = this_.parseDelete(r)
	case "create", "alter", "drop":
		res.IsDDL = true
		res.Root, err = this_.parseDDL(r)
	default:
		res.IsOther = true
		res.Root, err = this_.parseBlock(r)
	}
	if err != nil {
		return
	}
	return
}
func (this_ *SqlTemplate) parseSelect(r *SqlReader) (res *SqlNodeSelect, err error) {
	r.openScope()
	defer r.closeScope()

	r.InSelect = true

	res = &SqlNodeSelect{}
	res.SqlNodeBlock = &SqlNodeBlock{}
	res.Children, err = this_.parseChildren(r)

	return
}
func (this_ *SqlTemplate) parseInsert(r *SqlReader) (res *SqlNodeInsert, err error) {
	r.openScope()
	defer r.closeScope()

	r.InInsert = true

	res = &SqlNodeInsert{}
	res.SqlNodeBlock = &SqlNodeBlock{}
	res.Children, err = this_.parseChildren(r)
	return
}
func (this_ *SqlTemplate) parseUpdate(r *SqlReader) (res *SqlNodeUpdate, err error) {
	r.openScope()
	defer r.closeScope()

	r.InUpdate = true

	res = &SqlNodeUpdate{}
	res.SqlNodeBlock = &SqlNodeBlock{}
	res.Children, err = this_.parseChildren(r)

	return
}
func (this_ *SqlTemplate) parseDelete(r *SqlReader) (res *SqlNodeDelete, err error) {
	r.openScope()
	defer r.closeScope()

	r.InDelete = true

	res = &SqlNodeDelete{}
	res.SqlNodeBlock = &SqlNodeBlock{}
	res.Children, err = this_.parseChildren(r)
	return
}

type ReadText struct {
	text, textL, scan, scanL string
}

func getReadText(r *SqlReader) (res *ReadText) {
	text := r.str
	// 如果是 ' " ` 开头 解析到结尾 跳过 \' \" \`
	if text == "'" || text == "\"" || text == "`" {
		r.NextByCharEnd(text, "\\"+r.str)
		text += r.str
	}

	res = &ReadText{}
	res.text = text
	res.textL = strings.ToLower(res.text)
	res.scan = r.Scan()
	res.scanL = strings.ToLower(res.scan)
	return
}
func (this_ *SqlTemplate) parseDDL(r *SqlReader) (res *SqlNodeDDL, err error) {
	r.openScope()
	defer r.closeScope()

	r.InDDL = true

	res = &SqlNodeDDL{}
	res.SqlNodeBlock = &SqlNodeBlock{}
	res.Children, err = this_.parseChildren(r)
	return
}
func (this_ *SqlTemplate) parseBlock(r *SqlReader) (res *SqlNodeBlock, err error) {
	r.openScope()
	defer r.closeScope()

	res = &SqlNodeBlock{}
	if r.str == "(" {
		res.Children = append(res.Children, &SqlNodeText{Text: r.str})
		r.NextHasChar()
		var t *SqlNodeSql
		t, err = this_.ParseSql(r)
		if err != nil {
			return
		}
		res.Children = append(res.Children, t)
		if r.str != ")" {
			err = errors.New("sql 快 必须 [()] 成对出现，以[)]结尾")
			return
		}
		res.Children = append(res.Children, &SqlNodeText{Text: r.str})
		r.Next()
	} else {
		res.Children, err = this_.parseChildren(r)
	}
	return
}
func (this_ *SqlTemplate) parseChildren(r *SqlReader) (res []SqlNode, err error) {
	var lastTextT *SqlNodeText
	for !r.isEnd {
		var t SqlNode
		if r.str == "}" || r.str == ";" || (this_.BracketOptional && r.str == "]") {
			break
		}
		//if r.InExpression && r.str == ")" {
		//	return
		//}
		t, err = this_.parseNode(r)
		if err != nil {
			return
		}
		text, isText := t.(*SqlNodeText)
		if !isText {
			if lastTextT != nil && lastTextT.Text != "" {
				res = append(res, lastTextT)
			}
			res = append(res, t)
			lastTextT = nil
		} else {
			if lastTextT == nil {
				lastTextT = &SqlNodeText{}
			}
			lastTextT.Text += text.Text
		}
	}
	if lastTextT != nil && lastTextT.Text != "" {
		res = append(res, lastTextT)
	}
	return
}

type ChangeSqlReader func(r *SqlReader)

func (this_ *SqlTemplate) parseBaseNode(r *SqlReader, rt *ReadText) (res SqlNode, err error) {
	if rt.text == "--" {
		c := &SqlNodeComment{}
		c.Comment = r.NextLine()
		res = c
		return
	}
	if rt.text == "$" && rt.scan == "{" {
		r.NextHasChar()
		r.NextHasChar()
		res, err = this_.parseArgParam(r)
		return
	}
	if (rt.text == "#" && rt.scan == "{") || rt.text == "{" {
		if rt.text == "#" {
			r.NextHasChar()
		}
		r.NextHasChar()
		res, err = this_.parseParam(r)
		return
	}
	if this_.BracketOptional && rt.text == "[" {
		r.NextHasChar()
		res, err = this_.parseOptional(r)
		return
	}
	if rt.text == "if" && rt.scan == "(" {
		r.NextHasChar()
		res, err = this_.parseIf(r)
		return
	}
	return
}
func (this_ *SqlTemplate) parseNode(r *SqlReader) (res SqlNode, err error) {

	var isWhiteSpace = IsWhiteSpace(r.str)
	//fmt.Println("parseNode start str:", r.str, " isWhiteSpace:", isWhiteSpace)
	// 如果 是 空白 直接返回 解析下一个有内容的节点
	if isWhiteSpace {
		whiteSpace := r.str
		r.Next()
		return &SqlNodeText{Text: whiteSpace}, nil
	}

	rt := getReadText(r)

	res, err = this_.parseBaseNode(r, rt)
	if err != nil {
		return
	}
	if res != nil {
		return
	}
	r.Next()
	return &SqlNodeText{Text: rt.text}, nil
}
func (this_ *SqlTemplate) parseParam(r *SqlReader) (res *SqlNodeParam, err error) {
	//fmt.Println("parseParam start")
	param, err := this_.parseExpression(r)
	if err != nil {
		return
	}
	if r.str != "}" {
		err = errors.New("拼接参数应该是`${xx}`，以`}`结尾")
		return
	}
	r.Next()
	res = &SqlNodeParam{}
	res.Param = param
	//fmt.Println("parseParam end")
	return
}
func (this_ *SqlTemplate) parseArgParam(r *SqlReader) (res *SqlNodeArgParam, err error) {
	param, err := this_.parseExpression(r)
	if err != nil {
		return
	}
	if r.str != "}" {
		err = errors.New("占位参数应该是`${xx}`，以`}`结尾")
		return
	}
	r.Next()
	res = &SqlNodeArgParam{}
	res.Param = param
	return
}

func (this_ *SqlTemplate) parseOptional(r *SqlReader) (res *SqlNodeOptional, err error) {
	var name string
	res = &SqlNodeOptional{}
	res.Children, err = this_.parseChildren(r)
	if err != nil {
		return
	}
	if r.str != "]" {
		name += r.str
		r.Next()
		if r.isEnd {
			err = errors.New("格式应该是`[xxx]`，以`]`结尾")
			return
		}
	}
	r.Next()
	return
}

func (this_ *SqlTemplate) parseIf(r *SqlReader) (res *SqlNodeIf, err error) {
	test, err := this_.parseExpression(r)
	if err != nil {
		err = errors.New("解析`if test` 异常:" + err.Error())
		return
	}
	//fmt.Println("parseIf test:", reflect.TypeOf(test))
	//fmt.Println("parseIf test end str:", r.str)
	if r.str != "{" {
		r.NextHasChar()
	}
	if r.str != "{" {
		err = errors.New("格式应该是`if(xx){`，内容以`{`开始")
		return
	}
	r.NextHasChar()
	children, err := this_.parseChildren(r)
	if err != nil {
		return
	}
	if r.str != "}" {
		err = errors.New("格式应该是`if(xx){}`，内容以`}`结尾")
		return
	}
	var elseIfs []*SqlNodeElseIf
	var elseIf *SqlNodeElseIf
	var else_ *SqlNodeElse
	for {
		if strings.EqualFold(r.Scan(), "else") {
			r.NextHasChar()
			if !strings.EqualFold(r.str, "else") {
				err = errors.New("格式应该是`}else`，内容以`else`开头")
				return
			}
			if strings.EqualFold(r.Scan(), "if") {
				r.NextHasChar()
				if !strings.EqualFold(r.str, "if") {
					err = errors.New("格式应该是`}else if`，内容以`else if`开头")
					return
				}
				r.NextHasChar()
				elseIf, err = this_.parseElseIf(r)
				if err != nil {
					return
				}
				elseIfs = append(elseIfs, elseIf)
			} else {
				r.NextHasChar()
				else_, err = this_.parseElse(r)
				if err != nil {
					return
				}
				break
			}
		}
		break
	}
	r.Next()
	res = &SqlNodeIf{}
	res.Test = test
	res.ElseIfs = elseIfs
	res.Else = else_
	res.Children = children
	return
}

func (this_ *SqlTemplate) parseElseIf(r *SqlReader) (res *SqlNodeElseIf, err error) {
	test, err := this_.parseExpression(r)
	if err != nil {
		err = errors.New("解析`if test` 异常:" + err.Error())
		return
	}
	//fmt.Println("parseIf test:", reflect.TypeOf(test))
	//fmt.Println("parseIf test end str:", r.str)
	if r.str != "{" {
		r.NextHasChar()
	}
	if r.str != "{" {
		err = errors.New("格式应该是`else if(xx){`，内容以`{`开始")
		return
	}
	r.NextHasChar()
	children, err := this_.parseChildren(r)
	if err != nil {
		return
	}
	if r.str != "}" {
		err = errors.New("格式应该是`else if(xx){}`，内容以`}`结尾")
		return
	}
	r.Next()
	res = &SqlNodeElseIf{}
	res.Test = test
	res.Children = children
	return
}

func (this_ *SqlTemplate) parseElse(r *SqlReader) (res *SqlNodeElse, err error) {

	//fmt.Println("parseIf test:", reflect.TypeOf(test))
	//fmt.Println("parseIf test end str:", r.str)
	if r.str != "{" {
		r.NextHasChar()
	}
	if r.str != "{" {
		err = errors.New("格式应该是`if(xx){`，内容以`{`开始")
		return
	}
	r.NextHasChar()
	children, err := this_.parseChildren(r)
	if err != nil {
		return
	}
	if r.str != "}" {
		err = errors.New("格式应该是`else{}`，内容以`}`结尾")
		return
	}
	r.Next()
	res = &SqlNodeElse{}
	res.Children = children
	return
}
func ToSqlNodeString(str string) (res *SqlNodeString) {
	v := str[1 : len(str)-1]
	res = &SqlNodeString{
		Value: v,
	}
	return
}
func ToSqlNodeInt(str string) (res *SqlNodeInt, err error) {
	v, e := strconv.ParseInt(str, 10, 64)
	if e != nil {
		err = errors.New("value:" + str + " 无法转为 int64")
		return
	}
	res = &SqlNodeInt{
		Value:  v,
		String: str,
	}
	return
}
func ToSqlNodeFloat(str string) (res *SqlNodeFloat, err error) {
	v, e := strconv.ParseFloat(str, 64)
	if e != nil {
		err = errors.New("value:" + str + " 无法转为 float64")
		return
	}
	res = &SqlNodeFloat{
		Value:  v,
		String: str,
	}
	return
}

func (this_ *SqlTemplate) parseDotExpression(r *SqlReader, left SqlExpression) (res SqlExpression, err error) {
	if r.str != "." {
		return
	}
	r.NextHasChar()
	//fmt.Println("parseDotExpression left:", util.GetStringValue(left))
	//fmt.Println("parseDotExpression left next str:", r.str)
	right, err := this_.parseBaseExpression(r)
	if err != nil {
		return
	}
	//fmt.Println("parseDotExpression right:", util.GetStringValue(right))
	//fmt.Println("parseDotExpression right next str:", r.str)
	t := &SqlNodeDotExpression{
		Left:  left,
		Right: right,
	}
	res = t
	return
}

func (this_ *SqlTemplate) parseCallExpression(r *SqlReader, left SqlExpression) (res SqlExpression, err error) {
	if r.str != "(" {
		return
	}
	r.NextHasChar()
	right, err := this_.parseExpression(r)
	if err != nil {
		return
	}
	t := &SqlNodeDotExpression{
		Left:  left,
		Right: right,
	}
	res = t
	return
}
func (this_ *SqlTemplate) parseBaseExpression(r *SqlReader) (res SqlExpression, err error) {
	var canHasDot bool
	var canHasCall bool

	if r.str == "(" {
		res, err = this_.parseParenthesisExpression(r)
		if err != nil {
			return
		}
		canHasDot = true
		canHasCall = true
	} else if IsIdentifier(r.str) {
		res = &SqlNodeIdentifier{
			Name: r.str,
		}
		r.NextHasChar()
		canHasDot = true
		canHasCall = true
	} else if IsInt(r.str) {
		res, err = ToSqlNodeInt(r.str)
		if err != nil {
			return
		}
		r.NextHasChar()
	} else if IsFloat(r.str) {
		res, err = ToSqlNodeFloat(r.str)
		if err != nil {
			return
		}
		r.NextHasChar()
	} else if r.str == "'" || r.str == "\"" || r.str == "`" {
		s := r.str
		r.NextByCharEnd(s, "\\"+s)
		//fmt.Println("s:", s)
		s += r.str
		//fmt.Println("s:", s)
		res = ToSqlNodeString(s)
		r.NextHasChar()
		canHasDot = true
	} else {
		return
	}
	if canHasDot && r.str == "." {
		res, err = this_.parseDotExpression(r, res)
		if err != nil {
			return
		}
	}
	if canHasCall && r.str == "." {
		res, err = this_.parseCallExpression(r, res)
		if err != nil {
			return
		}
	}
	return
}
func (this_ *SqlTemplate) parseExpression(r *SqlReader) (res SqlExpression, err error) {
	r.openScope()
	defer r.closeScope()
	r.InExpression = true

	for !r.isEnd {
		if r.str == ")" || r.str == "]" || r.str == "}" || r.str == "{" || r.str == ":" {
			break
		}
		if res == nil {
			res, err = this_.parseBaseExpression(r)
			if err != nil {
				return
			}
		} else {
			var right SqlExpression
			switch r.str {
			case "+", "-", "*", "/", "%":
				op := r.str
				r.NextHasChar()
				right, err = this_.parseBaseExpression(r)
				if err != nil {
					err = errors.New("解析 [" + op + "] 表达式右边失败:" + err.Error())
					return
				}
				if right == nil {
					err = errors.New("解析 [" + op + "] 表达式 右侧 [" + r.str + "] 失败")
					return
				}

				lo, isLO := res.(*SqlNodeLogicalOperate)
				ro, isRO := res.(*SqlNodeRelationalOperate)
				// 如果 前一个是 逻辑运算 && ||
				if isLO {
					// 在 将运算追加到 逻辑运算符后边
					t := &SqlNodeArithmeticOperate{
						Left:     lo.Right,
						Operator: op,
						Right:    right,
					}
					lo.Right = t
					// 如果 前一个是 比较（关系）运算: == != > < >= <=
				} else if isRO {
					// 在 将运算追加到 逻辑运算符后边
					t := &SqlNodeArithmeticOperate{
						Left:     ro.Right,
						Operator: op,
						Right:    right,
					}
					ro.Right = t
				} else {
					t := &SqlNodeArithmeticOperate{
						Left:     res,
						Operator: op,
						Right:    right,
					}
					res = t
				}
				continue
			case "==", "!=", ">", "<", ">=", "<=":
				op := r.str
				r.NextHasChar()
				right, err = this_.parseBaseExpression(r)
				if err != nil {
					err = errors.New("解析 [" + op + "] 表达式右边失败:" + err.Error())
					return
				}
				if right == nil {
					err = errors.New("解析 [" + op + "] 表达式 右侧 [" + r.str + "] 失败")
					return
				}
				lo, isLO := res.(*SqlNodeLogicalOperate)
				// 如果 前一个是 逻辑运算 && ||
				if isLO {
					// 在 将运算追加到 逻辑运算符后边
					t := &SqlNodeRelationalOperate{
						Left:     lo.Right,
						Operator: op,
						Right:    right,
					}
					lo.Right = t
				} else {
					t := &SqlNodeRelationalOperate{
						Left:     res,
						Operator: op,
						Right:    right,
					}
					res = t
				}
				continue
			case "&&", "||":
				op := r.str
				r.NextHasChar()
				right, err = this_.parseBaseExpression(r)
				if err != nil {
					err = errors.New("解析 [" + op + "] 表达式右边失败:" + err.Error())
					return
				}
				if right == nil {
					err = errors.New("解析 [" + op + "] 表达式 右侧 [" + r.str + "] 失败")
					return
				}

				t := &SqlNodeLogicalOperate{
					Left:     res,
					Operator: op,
					Right:    right,
				}
				res = t
				continue
			// 三元表达式
			case "?":
				r.NextHasChar()
				t := &SqlNodeTernary{
					Test: res,
				}
				t.Left, err = this_.parseExpression(r)
				if err != nil {
					err = errors.New("解析 [三元] 表达式左侧值失败:" + err.Error())
					return
				}
				if r.str != ":" {
					err = errors.New("[三元] 表达式必须使用 [:] 隔开左右值")
					return
				}
				r.NextHasChar()
				t.Right, err = this_.parseExpression(r)
				if err != nil {
					err = errors.New("解析 [三元] 表达式左侧值失败:" + err.Error())
					return
				}
				res = t
				//r.ReadExpression.ReplaceLast(replaceIndex, res)
				continue
			default:
				err = errors.New("不支持 [" + r.str + "] 表达式")
				return
			}
		}
	}

	//fmt.Println("parseExpression left type:", reflect.TypeOf(left))
	//fmt.Println("parseExpression left json:", GetStringValue(left))
	//fmt.Println("parseExpression left end str:", r.str)

	return
}

func (this_ *SqlTemplate) parseParenthesisExpression(r *SqlReader) (res SqlExpression, err error) {
	r.NextHasChar()

	t := &SqlNodeParenthesis{}
	res = t
	t.Expression, err = this_.parseExpression(r)
	if err != nil {
		return
	}
	if r.str != ")" {
		err = errors.New("表达式中`()`必须 以 `)`结尾")
		return
	}
	r.NextHasChar()
	return
}
