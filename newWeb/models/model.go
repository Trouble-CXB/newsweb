package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

//用户表
type User struct {
	Id       int
	UserName string `orm:"unique"`
	//设置 唯一
	Pwd string
	//用户与文章  多对多
	Articles []*Article `orm:"rel(m2m)"`
}

//文章表
type Article struct {
	Id int `orm:"pk;auto"`
	//主键 自增
	Title string `orm:"size(100)"`
	//大小
	Content string `orm:"size(500)"`
	//设置时间类型   自动添加当前时间
	Time time.Time `orm:"type(datetime);auto_now"`
	//默认值0
	ReadCount int `orm:"default(0)"`
	//允许为空
	Image string `orm:"null"`
	//文章与类型	一对多
	ArticleType *ArticleType `orm:"rel(fk)"`
	//用户与文章  多对多
	Users []*User `orm:"reverse(many)"`
}

//类型表
type ArticleType struct {
	Id    int
	TypeName string `orm:"size(200)"`
	//文章与类型	一对多
	Articles []*Article `orm:"reverse(many)"` //设置一对多的反向关系
}

func init() {
	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/newsweb?charset=utf8")
	orm.RegisterModel(new(User), new(Article), new(ArticleType))
	orm.RunSyncdb("default", false, true)
}
