package controllers

import (
	"database/sql"
	"fmt"
	"goblog/app/models/article"
	"goblog/pck/logger"
	"goblog/pck/route"
	"goblog/pck/view"
	"gorm.io/gorm"
	"net/http"
	"unicode/utf8"
)

// ArticlesController 处理静态页面
type ArticlesController struct {
}

func (*ArticlesController) Show(w http.ResponseWriter, r *http.Request) {
	// 获取 URL 参数
	id := route.GetRouteVariable("id", r)

	// 读取相应的文章数据
	article, err := article.Get(id)

	// 如果出现错误
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 数据找不到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 文章未找到")
		} else {
			// 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误")
		}
	} else {
		// 读取成功
		view.Render(w, view.D{
			"Article": article,
		}, "articles.show")
	}
}

func (*ArticlesController) Index(w http.ResponseWriter, r *http.Request) {

	// 获取结果集
	articles, err := article.GetAll()

	if err != nil {
		// 数据库错误
		logger.LogError(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500服务器错误")
	} else {
		// 读取成功
		view.Render(w, view.D{
			"Articles": articles,
		},"articles.index")
	}
}

func (*ArticlesController) Store(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	errors := validateArticleFormData(title, body)

	// 检查是否有错误
	if len(errors) == 0 {
		_article := article.Article{
			Title: title,
			Body:  body,
		}
		_article.Create()
		if _article.ID > 0 {
			fmt.Fprintf(w, "插入成功，ID为"+_article.GetStringID())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误")
		}
	} else {
		view.Render(w, view.D{
			"Title":  title,
			"Body":   body,
			"Errors": errors,
		}, "articles.create", "articles._form_field")
	}
}

func (*ArticlesController) Create(w http.ResponseWriter, r *http.Request) {
	view.Render(w, view.D{}, "articles.create", "articles._form_field")
}

func (*ArticlesController) Edit(w http.ResponseWriter, r *http.Request) {
	// 获取 URL 参数
	id := route.GetRouteVariable("id", r)

	// 读取相应的文章数据
	article, err := article.Get(id)

	// 如果出现错误
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 文章未找到")
		} else {
			// 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误")
		}
	} else {
		view.Render(w, view.D{
			"Title":  article.Title,
			"Body":   article.Body,
			"Article":  article,
			"Errors": nil,
		}, "articles.edit", "articles._form_field")
	}
}

func (*ArticlesController) Update(w http.ResponseWriter, r *http.Request) {

	id := route.GetRouteVariable("id", r)

	_article, err := article.Get(id)

	if err != nil {
		if err == sql.ErrNoRows {
			// 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 数据未找到")
		} else {
			// 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器错误")
		}
	} else {
		// 未出现错误

		// 表单验证
		title := r.PostFormValue("title")
		body := r.PostFormValue("body")

		errors := validateArticleFormData(title, body)

		if len(errors) == 0 {
			// 表单验证通过，更新数据
			_article.Title = title
			_article.Body = body

			rowsAffected, err := _article.Update()

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "500 服务器内部错误")
			}

			// 更新成功，跳转到文章详情页
			if rowsAffected > 0 {
				showURL := route.Name2URL("articles.show", "id", id)
				http.Redirect(w, r, showURL, http.StatusFound)
			} else {
				fmt.Fprintf(w, "您没有做任何更改！")
			}
		} else {
			view.Render(w, view.D{
				"Title":  title,
				"Body":   body,
				"Article":    _article,
				"Errors": errors,
			}, "articles.edit", "articles._form_field")
		}
	}
}

func (*ArticlesController) Delete(w http.ResponseWriter, r *http.Request) {
	// 获取URL参数
	id := route.GetRouteVariable("id", r)

	// 读取对应的文章数据
	_article, err := article.Get(id)

	// 如果出现错误
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 数据未找到")
		} else {
			// 数据错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器错误")
		}
	} else {
		// 未出现错误
		rowsAffected, err := _article.Delete()

		// 发生错误
		if err != nil {
			// 应该是 sql 报错了
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器错误")
		} else {
			// 未发生错误
			if rowsAffected > 0 {
				// 重定向到文章列表
				indexURL := route.Name2URL("articles.index")
				http.Redirect(w, r, indexURL, http.StatusFound)
			} else {
				// 文章未找到
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "404 文章未找到")
			}
		}
	}
}

func validateArticleFormData(title string, body string) map[string]string {
	errors := make(map[string]string)

	// 验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于 3-40"
	}

	// 验证内容
	if body == "" {
		errors["body"] = "内容不存在"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容需大于等于10个字符"
	}

	return errors
}
