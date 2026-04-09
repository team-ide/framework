package db

import (
	"regexp"
	"strings"
	"unicode"
)

func NewSqlReader(str string) (r *SqlReader) {
	r = &SqlReader{}
	r.ReadState = &ReadState{}
	r.ReadScope = &ReadScope{}
	r.idx = -1
	r.offset = -1
	r.str = ""
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\r", "\n")
	r.body = strings.Split(str, "")
	r.len = len(r.body)
	return
}

type SqlReader struct {
	body []string
	len  int

	*ReadState

	*ReadScope
}
type ReadScope struct {
	InSelect       bool `json:"inSelect,omitempty"`
	InSelectColumn bool `json:"inSelectColumn,omitempty"`
	InSelectFrom   bool `json:"inSelectFrom,omitempty"`
	InInsert       bool `json:"inInsert,omitempty"`
	InInsertColumn bool `json:"inInsertColumn,omitempty"`
	InInsertValue  bool `json:"inInsertValue,omitempty"`
	InUpdate       bool `json:"inUpdate,omitempty"`
	InUpdateColumn bool `json:"inUpdateColumn,omitempty"`
	InDelete       bool `json:"inDelete,omitempty"`
	InFrom         bool `json:"inFrom,omitempty"`
	InJoin         bool `json:"inJoin,omitempty"`
	InJoinOn       bool `json:"inJoinOn,omitempty"`
	InWhere        bool `json:"inWhere,omitempty"`
	InGroupBy      bool `json:"inGroupBy,omitempty"`
	InOrderBy      bool `json:"inOrderBy,omitempty"`
	InValue        bool `json:"inValue,omitempty"`
	InDDL          bool `json:"inDDL,omitempty"`

	InExpression bool `json:"inExpression,omitempty"`

	// 解析 算术运算符: + - * / %
	ArithmeticOperator bool `json:"arithmeticOperator,omitempty"`
	// 解析 比较（关系）运算符: == != > < >= <=
	RelationalOperator bool `json:"relationalOperator,omitempty"`
	// 解析 逻辑运算符: && ||
	LogicalOperator bool `json:"logicalOperator,omitempty"`

	lastScope *ReadScope
}

func (this_ *SqlReader) openScope() {
	newScope := &ReadScope{}
	newScope.lastScope = this_.ReadScope
	this_.ReadScope = newScope
}

func (this_ *SqlReader) closeScope() {
	this_.ReadScope = this_.lastScope
}

type ReadExpression struct {
	StartExpression SqlExpression   `json:"startExpression,omitempty"`
	LastExpression  SqlExpression   `json:"lastExpression,omitempty"`
	Expressions     []SqlExpression `json:"expressions,omitempty"`
}

func (this_ *ReadExpression) Add(e SqlExpression) {

	if this_.StartExpression == nil {
		this_.StartExpression = e
	}
	this_.LastExpression = e
	this_.Expressions = append(this_.Expressions, e)
}
func (this_ *ReadExpression) ReplaceLast(replaceIndex int, e SqlExpression) {
	replaceFirst := this_.Expressions[replaceIndex]
	if this_.StartExpression == replaceFirst {
		this_.StartExpression = e
	}
	this_.LastExpression = e
	this_.Expressions = this_.Expressions[:replaceIndex]
	this_.Expressions = append(this_.Expressions, e)
}

type ReadState struct {
	// 起始位置
	idx int
	// 读取到的位置
	offset int
	// 读取的字符
	str string

	isEnd bool

	isWhiteSpace bool

	lastState *ReadState
}

func (this_ *SqlReader) Next() {
	this_.lastState = this_.ReadState
	this_.ReadState = this_.read()
}
func (this_ *SqlReader) NextHasChar() {
	this_.lastState = this_.ReadState
	for {
		this_.ReadState = this_.read()
		if this_.isEnd {
			return
		}
		str := strings.TrimSpace(this_.str)
		if str != "" {
			break
		}
	}
}
func (this_ *SqlReader) NextLine() (res string) {
	this_.lastState = this_.ReadState
	for {
		this_.ReadState = this_.read()
		if this_.isEnd {
			return
		}
		if this_.str == "\n" {
			break
		}
		res += this_.str
	}
	this_.Next()
	return
}
func (this_ *SqlReader) NextByCharEnd(endChar string, skipChar string) {
	this_.lastState = this_.ReadState
	var str string
	var idx int

	for !this_.isEnd {
		this_.ReadState = this_.read()
		if this_.isEnd {
			return
		}
		str += this_.str
		if this_.str == endChar {
			if skipChar != "" && strings.HasSuffix(str, skipChar) {

			} else {
				break
			}
		}
	}
	this_.str = str
	this_.idx = idx
}

var (
	identifierReg, _ = regexp.Compile(`^[a-zA-Z_$]+([a-zA-Z_$\d]+)?$`)
	numReg, _        = regexp.Compile(`^-?\d+$`)
	intReg, _        = regexp.Compile(`^\d+$`)
	floatReg, _      = regexp.Compile(`^-?\d+(\.\d+)?$`)
)

func IsIdentifier(s string) bool {
	return identifierReg.MatchString(s)
}
func IsNumber(s string) bool {
	return numReg.MatchString(s)
}
func IsInt(s string) bool {
	return intReg.MatchString(s)
}
func IsFloat(s string) bool {
	return floatReg.MatchString(s)
}

func (this_ *SqlReader) Scan() (res string) {
	readState := this_.ReadState
	for {
		this_.ReadState = this_.read()
		if this_.isEnd {
			return
		}
		if strings.TrimSpace(this_.str) != "" {
			break
		}
	}
	res = this_.str
	this_.ReadState = readState
	return
}

func (this_ *SqlReader) ScanMore(size int) (res []string) {
	readState := this_.ReadState
	for {
		if len(res) >= size {
			break
		}
		this_.ReadState = this_.read()
		if this_.isEnd {
			return
		}
		if strings.TrimSpace(this_.str) != "" {
			res = append(res, this_.str)
		}
	}
	this_.ReadState = readState
	return
}

func (this_ *SqlReader) read() (res *ReadState) {
	res = &ReadState{}
	res.idx = this_.offset + 1
	res.offset = res.idx
	var char string
	char, res.isEnd = this_.readChar(res.offset)
	if res.isEnd {
		return
	}
	res.str += char
	if identifierReg.MatchString(char) || char == "$" {
		for {
			char, res.isEnd = this_.readChar(res.offset + 1)
			if res.isEnd {
				break
			}
			if identifierReg.MatchString(char) || intReg.MatchString(char) {
				res.offset++
				res.str += char
				continue
			}
			break
		}
	} else if numReg.MatchString(char) || char == "-" {
		for {
			char, res.isEnd = this_.readChar(res.offset + 1)
			if res.isEnd {
				break
			}
			if numReg.MatchString(char) || char == "." {
				res.offset++
				res.str += char
				continue
			}
			break
		}
	} else if char == "=" || char == "!" || char == ">" || char == "<" {
		char, res.isEnd = this_.readChar(res.offset + 1)
		if !res.isEnd {
			if char == "=" {
				res.offset++
				res.str += char
			}
		}

	} else if char == "&" || char == "|" || char == "-" || char == "/" {
		lC := char
		char, res.isEnd = this_.readChar(res.offset + 1)
		if !res.isEnd {
			if char == lC {
				res.offset++
				res.str += char
			}
		}
	} else if IsWhiteSpace(char) {
		for {
			char, res.isEnd = this_.readChar(res.offset + 1)
			if res.isEnd {
				break
			}
			if IsWhiteSpace(char) {
				res.offset++
				res.str += char
				continue
			}
			break
		}
		res.isWhiteSpace = true
	}

	return
}
func IsWhiteSpace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			//fmt.Println("IsWhiteSpace not space r:", r)
			return false
		}
	}
	return true
}

//	func IsWhiteSpace(char string) bool {
//		switch char {
//		case " ", "\t", "\v", "\u00a0", "'\ufeff'":
//			return true
//		case "\n", "\r":
//			return true
//		}
//		return false
//	}
func HasNewLine(str string) bool {
	return strings.Contains(str, "\n")
}
func (this_ *SqlReader) readChar(index int) (char string, isEnd bool) {
	if index >= this_.len {
		isEnd = true
		return
	}
	char = this_.body[index]
	return
}
