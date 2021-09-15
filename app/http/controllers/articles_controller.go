package controllers

import (
	"database/sql"
	"fmt"
	"goblog/app/models/article"
	"goblog/pck/logger"
	"goblog/pck/route"
	"goblog/pck/types"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"strconv"
	"unicode/utf8"
)

// ArticlesController 处理静态页面
type ArticlesController struct {
}

// ArticlesFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         string
	Errors      map[string]string
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
		// 4. 读取成功，显示文章
		tmpl, err := template.New("show.gohtml").
			Funcs(template.FuncMap{
				"RouteName2URL": route.Name2URL,
				"Int64ToString": types.Int64ToString,
			}).
			ParseFiles("resources/views/articles/show.gohtml")
		logger.LogError(err)
		tmpl.Execute(w, article)
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
		// 加载模板
		tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
		logger.LogError(err)

		// 渲染模板
		tmpl.Execute(w, articles)
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
			fmt.Fprintf(w, "插入成功，ID为"+strconv.FormatInt(_article.ID, 10))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误")
		}
	} else {
		storeURL := route.Name2URL("articles.store")

		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}

		tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
		if err != nil {
			panic(err)
		}

		tmpl.Execute(w, data)
	}

}

func (*ArticlesController) Create(w http.ResponseWriter, r *http.Request) {
	storeURL := route.Name2URL("articles.store")
	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}

	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}
	tmpl.Execute(w, data)
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
		// 读取成功，显示表单
		updateURL := route.Name2URL("articles.update", "id", id)
		data := ArticlesFormData{
			Title: article.Title,
			Body: article.Body,
			URL: updateURL,
			Errors: nil,
		}
		tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
		logger.LogError(err)

		tmpl.Execute(w, data)
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
			// 表单验证不通过
			updateURL := route.Name2URL("articles.update", "id", id)
			data := ArticlesFormData{
				Title: title,
				Body: body,
				URL: updateURL,
				Errors: errors,
			}
			tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
			logger.LogError(err)

			tmpl.Execute(w, data)
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

func validateArticleFormData(title string, body string) map[string]string  {
	errors := make(map[string]string)

	// 验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于 3-40"
	}

	// 验证内容
	if body == ""{
		errors["body"] = "内容不存在"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容需大于等于10个字符"
	}

	return errors
}
