package main

import (
	"database/sql"
	"fmt"
	"goblog/bootstrap"
	"goblog/pck/database"
	"goblog/pck/logger"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var router *mux.Router
var db *sql.DB

// ArticlesFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

// Article 对应的一条文章数据
type Article struct {
	Title, Body string
	ID          int64
}

func forceHTMLMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		h.ServeHTTP(w, r)
	})
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 除首页以外，移除所有请求路径后面的斜杠
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}

		// 将请求传递下去
		next.ServeHTTP(w, r)
	})
}

func articlesEditHandler(w http.ResponseWriter, r *http.Request)  {

	// 获取 URL 参数
	id := getRouteVariable("id", r)

	// 读取相应的文章数据
	article, err := getArticleByID(id)

	// 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
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
		updateURL, _ := router.Get("articles.update").URL("id", id)
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

func articlesUpdateHandler(w http.ResponseWriter, r *http.Request)  {
	
	id := getRouteVariable("id", r)

	_, err := getArticleByID(id)

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
			query := "UPDATE articles SET title = ?, body = ? where id = ?"
			rs, err := db.Exec(query, title, body, id)

			if err != nil {
				logger.LogError(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "500 服务器内部错误")
			}

			// 更新成功，跳转到文章详情页
			if n, _ := rs.RowsAffected(); n > 0 {
				showURL, _ := router.Get("articles.show").URL("id", id)
				http.Redirect(w, r, showURL.String(), http.StatusFound)
			} else {
				fmt.Fprintf(w, "您没有做任何更改！")
			}
		} else {
			// 表单验证不通过
			updateURL, _ := router.Get("articles.update").URL("id", id)
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

func articlesDeleteHandler(w http.ResponseWriter, r *http.Request)  {

	// 获取URL参数
	id := getRouteVariable("id", r)

	// 读取对应的文章数据
	article, err := getArticleByID(id)

	// 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
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
		rowsAffected, err := article.Delete()

		// 发生错误
		if err != nil {
			// 应该是 sql 报错了
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器错误")
		} else {
			// 未发生错误
			if rowsAffected > 0 {
				// 重定向到文章列表
				indexURL, _ := router.Get("articles.index").URL()
				http.Redirect(w, r, indexURL.String(), http.StatusFound)
			} else {
				// 文章未找到
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "404 文章未找到")
			}
		}
	}


}

func getArticleByID(id string) (Article, error)  {
	article := Article{}
	query := "SELECT * FROM articles WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	return article, err
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

func (a Article) Delete() (rowsAffected int64, err error)  {
	rs, err := db.Exec("DELETE FROM articles WHERE id = " + strconv.FormatInt(a.ID, 10))

	if err != nil {
		return 0, err
	}

	// 删除成功，跳转到文章详情页
	if n, _ := rs.RowsAffected(); n > 0 {
		return n, nil
	}

	return 0, err
}

func getRouteVariable(parameterName string, r *http.Request) string {
	vars := mux.Vars(r)
	return vars[parameterName]
}

func main() {
	database.Initialize()
	db = database.DB

	bootstrap.SetupDB()
	router = bootstrap.SetupRoute()

	router.HandleFunc("/articles/{id:[0-9]+}/edit",
		articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}",
		articlesUpdateHandler).Methods("POST").Name("articles.update")
	router.HandleFunc("/articles/{id:[0-9]+}/delete",
		articlesDeleteHandler).Methods("POST").Name("articles.delete")

	// 中间件：强制内容类型为 HTML
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe(":3000", removeTrailingSlash(router))
}
