package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"math"
	"newWeb/models"
	"path"
	"strconv"
	"time"
)

type ArticleController struct {
	beego.Controller
}

//展示文章列表页
func (this *ArticleController) ShowArticleList() {
	//用户名获取
	userName:=this.GetSession("userName")
	this.Data["userName"] = userName.(string)
	//获取所有类型
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)
	this.Data["articleTypes"] = articleTypes

	//查询数据库，取出数据，传给视图
	var articles []models.Article
	//查询数据    qs:queryseter  高级查询使用的数据类型
	qs := o.QueryTable("Article")
	//查询所有数据(接收对象)	相当于select * from Article
	//qs.All(&articles)
	//beego.Info(articles)

	/////实现分页/////
	//获取总记录数
	count, _ := qs.Count()
	//每页多少条记录
	pageSize := int64(2)
	//获总页数
	pageCount := math.Ceil(float64(count) / float64(pageSize))

	//把数据传递给视图
	this.Data["count"] = count
	this.Data["pageCount"] = pageCount

	//获取首页末页数据
	//pageIndex  当前页
	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	//获取分页的数据
	start := pageSize * (int64(pageIndex) - 1)
	//RelatedSel	一对多关系表查询中，用来制定
	qs.Limit(pageSize,start).RelatedSel("ArticleType").All(&articles)

	this.Data["pageIndex"] = pageIndex
	this.Data["articles"] = articles

	errmsg := this.GetString("errmsg")
	if errmsg != "" {
		this.Data["errmsg"] = errmsg
	}

	//根据传递的类型获取相应的文章
	//获取数据
	typeName:=this.GetString("select")
	this.Data["typeName"] = typeName

	//qs.Limit(pageSize,start).RelatedSel("ArticleType").Filter("ArticleType_TypeName",typeName).All(&articles)
	qs.Limit(pageSize,start).RelatedSel("ArticleType").Filter("ArticleType__TypeName",typeName).All(&articles)

	this.Layout = "layout.html"
	this.TplName = "index.html"
}

//展示文章添加页面
func (this *ArticleController) ShowAddArticle() {
	//用户名获取
	userName:=this.GetSession("userName")
	this.Data["userName"] = userName.(string)
	//获取所有类型
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)
	this.Data["articleTypes"] = articleTypes
	//渲染页面
	this.Layout = "layout.html"
	this.TplName = "add.html"
}

//处理添加文章业务
func (this *ArticleController) HandleAddArticle() {
	//储存信息
	articleName := this.GetString("articleName")
	content := this.GetString("content")

	//校验信息
	if articleName == "" || content == "" {
		this.Data["errmsg"] = "文章标题或内容不能为空"
		this.TplName = "add.html"
		return
	}
	//调用函数  接收图片并校验图片
	fileName := UploadFile(this, "uploadname")

	//获取信息储存到数据库
	o := orm.NewOrm()
	var article models.Article
	article.Title = articleName
	article.Content = content
	article.Image = fileName
	//类型储存
	typeName := this.GetString("select")
	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Read(&articleType, "TypeName")
	article.ArticleType = &articleType
	//储存进数据库
	_, err := o.Insert(&article)

	if err != nil {
		beego.Error("文章上传失败", err)
		this.Data["errmsg"] = "文章上传失败，请重新上传。"
		this.TplName = "add.html"
		return
	}
	//数据库储存成功后进行文件存储
	this.SaveToFile("uploadname", "./static/image/"+fileName)

	//返回页面
	this.Redirect("/article/articleList", 302)
}

//展示文章详情页
func (this *ArticleController) ShowArticleDetail() {
	articleId, err := this.GetInt("id")
	if err != nil {
		this.Data["errmsg"] = "请求路径错误！"
		this.TplName = "index.html"
	}

	o := orm.NewOrm()
	var article models.Article
	article.Id = articleId
	err = o.Read(&article)
	if err != nil {
		this.Data["errmsg"] = "读取错误"
		this.TplName = "index.html"
		return
	}
	//增加阅读次数
	article.ReadCount ++
	o.Update(&article)


	m2m:=o.QueryM2M(&article,"Users")
	var user models.User
	userName:=this.GetSession("userName")
	user.UserName = userName.(string)
	o.Read(&user,"UserName")
	//添加用户在关系表中
	m2m.Add(user)

	//第一种多对多查询
	o.LoadRelated(&article,"Users")

	////第二种多对多关系查询
	////filter  过滤器  指定查询条件，进行过滤查找
	var users []models.User
	o.QueryTable("User").Filter("Articles__Article__Id",articleId).Distinct().All(&users)
	this.Data["users"] = users

	this.Data["userName"] = userName.(string)
	this.Data["article"] = article

	this.Layout = "layout.html"
	this.TplName = "content.html"
}

//展示编辑文章页面
func (this *ArticleController) ShowUpdateArticle() {
	//用户名获取
	userName:=this.GetSession("userName")
	this.Data["userName"] = userName.(string)

	articleId, err := this.GetInt("id")
	if err != nil {
		errmsg := "请求路径错误"
		this.Redirect("/article/articleList?errmsg="+errmsg, 302)
		return
	}

	errmsg := this.GetString("errmsg")
	if errmsg != "" {
		this.Data["errmsg"] = errmsg
	}

	o := orm.NewOrm()
	var article models.Article
	article.Id = articleId
	o.Read(&article)

	this.Data["article"] = article

	this.Layout = "layout.html"
	this.TplName = "update.html"
}

//处理编辑文章页面
func (this *ArticleController) HandleUpdateArticle() {
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	fileName := UploadFile(this, "uploadname")
	articleId, err2 := this.GetInt("id")
	//检验数据
	if articleName == "" || content == "" || fileName == "" || err2 != nil {
		errmsg := "内容不能为空"
		this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"errmsg="+errmsg, 302)
		return
	}

	//信息储存到数据库
	o := orm.NewOrm()
	var article models.Article
	//更新前先查找数据存不存在
	article.Id = articleId
	err := o.Read(&article)
	if err != nil {
		errmsg := "更新文章不存在"
		this.Redirect("/article/updateArticle?id="+strconv.Itoa(articleId)+"errmsg="+errmsg, 302)
		return
	}
	article.Title = articleName
	article.Content = content
	article.Image = fileName
	//更新数据库
	o.Update(&article)

	//数据库储存成功后进行文件存储
	this.SaveToFile("uploadname", "./static/image/"+fileName)

	//返回页面
	this.Redirect("/article/articleList", 302)
}

//删除文章
func (this *ArticleController) DeleteArticle() {
	//获取Id
	articleId, err := this.GetInt("id")
	if err != nil {
		errmsg := "请求路径错误"
		this.Redirect("/article/articleList?errmsg="+errmsg, 302)
		return
	}
	//删除操作
	o := orm.NewOrm()
	var article models.Article
	article.Id = articleId

	//o.Read(&article)
	//fileName:=article.Image
	//删除数据库
	_, err = o.Delete(&article)
	if err != nil {
		errmsg := "删除失败！！"
		this.Redirect("/article/articleList?errmsg="+errmsg, 302)
		return
	}
	//删除文件
	//os.Remove(fileName)

	this.Redirect("/article/articleList", 302)
}

//展示添加类型界面
func (this *ArticleController) ShowAddType() {
	//用户名获取
	userName:=this.GetSession("userName")
	this.Data["userName"] = userName.(string)

	//获取所有类型
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	qs := o.QueryTable("ArticleType")
	qs.All(&articleTypes)

	this.Data["articleTypes"] = articleTypes

	errmsg := this.GetString("errmsg")
	if errmsg != "" {
		this.Data["errmsg"] = errmsg
	}

	this.Layout = "layout.html"
	this.TplName = "addType.html"
}

//处理添加类型
func (this *ArticleController) HandleAddType() {
	//获取数据（类型名）
	typeName := this.GetString("typeName")
	//校验数据
	if typeName == "" {
		errmsg := "类型名不能为空"
		this.Redirect("/article/addType?errmsg="+errmsg, 302)
		return
	}
	//插入数据库
	o := orm.NewOrm()
	var acticleType models.ArticleType
	acticleType.TypeName = typeName
	_, err := o.Insert(&acticleType)
	if err != nil {
		errmsg := "添加类型不成功"
		this.Redirect("/article/addType?errmsg="+errmsg, 302)
		return
	}

	this.Redirect("/article/addType", 302)
}

//删除类型
func (this *ArticleController) DeleteType() {
	typeId, err := this.GetInt("id")
	if err != nil {
		errmsg := "类型获取失败"
		this.Redirect("/article/addType?errmsg="+errmsg, 302)
		return
	}

	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.Id = typeId
	_, err = o.Delete(&articleType)
	if err != nil {
		errmsg := "类型删除失败"
		this.Redirect("/article/addType?errmsg="+errmsg, 302)
		return
	}

	this.Redirect("/article/addType", 302)
}

//退出登陆
func (this *ArticleController)Logout()  {
	//删除Session
	this.DelSession("userName")
	//跳转页面
	this.Redirect("/login",302)
}

//文件接收并校验函数	返回：文件储存地址
func UploadFile(this *ArticleController, filePath string) string {
	file, head, err := this.GetFile(filePath)
	if err != nil {
		this.Data["errmsg"] = "获取文件失败"
		this.TplName = "add.html"
		return ""
	}
	defer file.Close()
	//1、判断文件大小
	if head.Size > 1024*1000*5 {
		this.Data["errmsg"] = "图片大于5M，上传失败，"
		this.TplName = "add.html"
		return ""
	}
	//2、判断图片格式
	fileExt := path.Ext(head.Filename)
	if fileExt != ".jpg" && fileExt != ".png" && fileExt != ".ico" {
		this.Data["errmsg"] = "图片格式不正确，请重新上传。"
		this.TplName = "add.html"
		return ""
	}
	//3、文件名重复
	fileName := time.Now().Format("2006:01:02-15:04:05") + fileExt
	return "/static/image/" + fileName
}
