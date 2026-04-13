package db

import (
	"fmt"
	"github.com/team-ide/framework"
	"testing"

	"github.com/team-ide/framework/util"
)

func TestSqlTemplate(t *testing.T) {
	var list []string
	list = append(list, `
if(name.length > 0){
  name_length  = ${name.length}
}
`)
	//list = append(list, `${(status== 22)?11:1}`)
	//list = append(list, `${user.status==5 && (status== 22||status==3)?11:1}`)
	//list = append(list, `${user.base.aaa==5 && (status== 22||status==3)?11:1}`)
	//list = append(list, `select * from user where name like ${name} and status = ${status!= null?status:1}`)
	//list = append(list, `select * from user join aa  where name like ${name}`)
	//list = append(list, `select *,name as aaa,(select 1) aaa,(select bb from bbb) bb from user join aa left join bb inner join cc right join dd cross join ee where name like ${name}`)

	var sqlHandler = NewSqlOption()
	sqlHandler.WrapNameOpen = util.NewBool(true)
	sqlHandler.WrapNameChar = util.NewString("`")

	for _, one := range list {
		st := NewSqlTemplate(one)
		st.Append()
		err := st.Parse()
		if err != nil {
			panic(err)
			return
		}
		var param = map[string]any{}
		param["name"] = "%三%"
		param["status"] = 3
		param["user"] = map[string]any{
			"name":   "张三",
			"status": 5,
			"base": map[string]any{
				"xx":  "张三",
				"aaa": 5,
			},
		}
		sqlInfo, sqlArgs := st.GetSql(sqlHandler, param)
		if err != nil {
			panic(err)
			return
		}
		fmt.Println("sql template:", one)
		fmt.Println("sql info:", sqlInfo)
		fmt.Println("sql args:", util.GetStringValue(sqlArgs))
	}
}

func TestMapper(t *testing.T) {
	var err error
	testUserMapper := NewTestUserMapper()
	err = testUserMapper.Insert(&TestUser{Name: "xxx"})
	if err != nil {
		panic(err)
		return
	}
	err = testUserMapper.Update(1, &TestUser{Name: "xxx"})
	if err != nil {
		panic(err)
		return
	}

}

var (
	testSqlHandler = NewSqlOption().Set(func(o *SqlOption) {
		o.WrapNameOpen = util.NewBool(true)
		o.WrapNameChar = util.NewString("`")
	})

	TableNameTestUser = "test_user"
)

type TestUser struct {
	// 用户 id 唯一 不可重复
	UserId int64 `json:"userId" column:"user_id"`
	// 用户名
	Name string `json:"name" column:"name"`
	// 用户登录账号 唯一 不可重复
	Account string `json:"account" column:"account"`
	// 密码 md5(salt + '' + md5(password))
	Password string `json:"password" column:"password"`
	// 密码 盐值
	Salt string `json:"salt" column:"salt"`
	// 用户状态 1 正常 2 停用  9删除
	Status int `json:"status" column:"status"`
	// 创建 时间戳 单位毫秒
	CreateAt int64 `json:"createAt" column:"create_at"`
	// 更新 时间戳 单位毫秒
	UpdateAt int64 `json:"updateAt" column:"update_at"`
	// 删除 时间戳 单位毫秒
	DeleteAt int64 `json:"deleteAt" column:"delete_at"`
}

func (this_ *TestUser) GetTableName() string {
	return TableNameTestUser
}
func (this_ *TestUser) GetPrimaryKey() []string {
	return []string{"user_id"}
}
func NewTestUserMapper() (res *TestUserMapper) {
	res = &TestUserMapper{}
	return
}

type TestUserMapper struct {
}

func (this_ *TestUserMapper) Query(user *TestUser) (err error) {
	m := NewModelSelect(user)
	m.SetSqlHandler(testSqlHandler)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("select sql:" + sqlInfo)
	framework.Info("select sql args:" + util.GetStringValue(args))
	return
}

func (this_ *TestUserMapper) Count(user *TestUser) (err error) {
	m := NewModelCount(user)
	m.SetSqlHandler(testSqlHandler)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("count sql:" + sqlInfo)
	framework.Info("count sql args:" + util.GetStringValue(args))
	return
}

func (this_ *TestUserMapper) GetById(userId int64) (res *TestUser, err error) {
	m := NewModelSelect(&TestUser{})
	m.SetSqlHandler(testSqlHandler)
	m.Where().Eq("user_id", userId)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("select sql:" + sqlInfo)
	framework.Info("select sql args:" + util.GetStringValue(args))
	return
}

func (this_ *TestUserMapper) GetByIds(userIds []int64) (res []*TestUser, err error) {
	m := NewModelSelect(&TestUser{})
	m.SetSqlHandler(testSqlHandler)
	m.Where().In("user_id", userIds)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("select sql:" + sqlInfo)
	framework.Info("select sql args:" + util.GetStringValue(args))
	return
}

func (this_ *TestUserMapper) QueryByAccount(account string) (res []*TestUser, err error) {
	m := NewModelSelect(&TestUser{})
	m.SetSqlHandler(testSqlHandler)
	m.Where().Eq("account", account)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("select sql:" + sqlInfo)
	framework.Info("select sql args:" + util.GetStringValue(args))
	return
}

func (this_ *TestUserMapper) Insert(user *TestUser) (err error) {
	m := NewModelInsert(user)
	m.SetSqlHandler(testSqlHandler)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("insert sql:" + sqlInfo)
	framework.Info("insert sql args:" + util.GetStringValue(args))
	return
}

func (this_ *TestUserMapper) Update(userId int64, user *TestUser) (err error) {
	m := NewModelUpdate(user)
	m.SetSqlHandler(testSqlHandler)
	m.Where().Eq("user_id", userId)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("update sql:" + sqlInfo)
	framework.Info("update sql args:" + util.GetStringValue(args))
	return
}

func (this_ *TestUserMapper) Delete(userId int64) (err error) {
	m := NewModelDelete(nil)
	m.SetSqlHandler(testSqlHandler)
	m.Where().Eq("user_id", userId)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("delete sql:" + sqlInfo)
	framework.Info("delete sql args:" + util.GetStringValue(args))
	return
}
func (this_ *TestUserMapper) BatchDelete(userIds []int64) (err error) {
	m := NewModelDelete(nil)
	m.SetSqlHandler(testSqlHandler)
	m.Where().In("user_id", userIds)
	sqlInfo, args, err := m.GetSql()
	if err != nil {
		panic(err.Error())
		return
	}
	framework.Info("batch delete sql:" + sqlInfo)
	framework.Info("batch delete sql args:" + util.GetStringValue(args))
	return
}
