package controllers

import (
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"newWeb/models"
)

type UserController struct {
	beego.Controller
}

//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

//处理注册信息
func (this *UserController) HandleRegister() {
	userName := this.GetString("userName")
	Pwd := this.GetString("password")

	if userName == "" || Pwd == "" {
		this.Data["errmsg"] = "用户名或密码为空"
		this.TplName = "register.html"
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.UserName = userName
	user.Pwd = Pwd
	_, err := o.Insert(&user)
	if err != nil {
		this.Data["errmsg"] = "服务器受到陨石攻击。注册失败，请重新注册"
		this.TplName = "register.html"
		return
	}
	//this.Ctx.Wri'teString("注册成功")
	/*Redirect和Tplame的区别

	*/
	this.Redirect("/login", 302)
	//this.TplName ="login.html"
}

//展示登陆页面
func (this *UserController) ShowLogin() {
	dec := this.Ctx.GetCookie("userName")
	userName,_:=base64.StdEncoding.DecodeString(dec)
	if string(userName) != "" {
		this.Data["userName"] = string(userName)
		this.Data["checked"] = "checked"
	} else {
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	}

	this.TplName = "login.html"
}

//处理登陆信息
func (this *UserController) HandleLogin() {
	/*
	1.接受数据
	2.校验数据
	3.处理数据
	4.返回数据
	*/
	userName := this.GetString("userName")
	Pwd := this.GetString("password")

	if userName == "" || Pwd == "" {
		this.Data["errmsg"] = "用户名或密码为空"
		this.TplName = "login.html"
		return
	}

	o := orm.NewOrm()
	var user models.User
	user.UserName = userName
	err := o.Read(&user, "UserName")
	if err != nil {
		this.Data["errmsg"] = "用户名不存在"
		this.TplName = "login.html"
		return
	}
	if user.Pwd != Pwd {
		this.Data["errmsg"] = "密码错误，请重新输入"
		this.TplName = "register.html"
		return
	}
	//获取是否记住用户名
	remember:=this.GetString("remember")
	enc:=base64.StdEncoding.EncodeToString([]byte(userName))
	if remember == "on" {
		this.Ctx.SetCookie("userName",enc,3600*1)
	}else {
		this.Ctx.SetCookie("userName",userName,-1)
	}
	//返回数据
	this.SetSession("userName",userName)
	//this.Ctx.WriteString("登陆成功")
	this.Redirect("/article/articleList", 302)
}
