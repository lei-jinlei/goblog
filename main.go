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

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {

	// 执行查询语句，返回一个结果集
	rows, err := db.Query("select * from articles")
	logger.LogError(err)
	defer rows.Close()

	var articles []Article

	// 循环读取结果
	for rows.Next() {
		var article Article
		// 扫描每一行的结果并赋值到一个 article 对象中
		err := rows.Scan(&article.ID, &article.Title, &article.Body)
		logger.LogError(err)
		// 将 article 追加到 articles 的这个数组中
		articles = append(articles, article)
	}

	// 检测遍历时是否发生错误
	err = rows.Err()
	logger.LogError(err)

	// 加载模板
	tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
	logger.LogError(err)

	tmpl.Execute(w, articles)
}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {

	title := r.FormValue("title")
	body := r.FormValue("body")

	errors := validateArticleFormData(title, body)

	// 检查是否有错误
	if len(errors) == 0 {
		fmt.Fprint(w, "验证通过！<br>")
		fmt.Fprintf(w, "title 的值为：%v <br>", title)
		fmt.Fprintf(w, "title 的长度为：%v <br>", utf8.RuneCountInString(title))
		fmt.Fprintf(w, "body 的值为：%v <br>", body)
		fmt.Fprintf(w, "body 的长度为：%v <br>", utf8.RuneCountInString(body))
		lastInsertID, err := saveArticleToDB(title, body)
		if lastInsertID > 0 {
			fmt.Fprintf(w, "插入成功，ID为"+strconv.FormatInt(lastInsertID, 10))
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 服务器内部错误")
		}
	} else {
		storeURL, _ := router.Get("articles.store").URL()

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

func saveArticleToDB(title string, body string) (int64, error) {

	// 变量初始化
	var (
		id   int64
		err  error
		rs   sql.Result
		stmt *sql.Stmt
	)

	// 1.获取一个 prepare 声明语句
	stmt, err = db.Prepare("INSERT INTO articles (title, body) VALUES(?,?)")
	// 例行错误检查
	if err != nil {
		return 0, err
	}

	// 2.在此函数运行结束后关闭此语句，防止占用 SQL 连接
	defer stmt.Close()

	// 3.执行请求，传参进入绑定内容
	rs, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}

	// 4.插入成功的话，会返回自增 ID
	if id, err = rs.LastInsertId(); id > 0 {
		return id, nil
	}

	return 0, err
}

func articlesCreateHandler(w http.ResponseWriter, r *http.Request) {
	storeURL, _ := router.Get("articles.store").URL()
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

func (a Article) Link() string  {
	showURL, err := router.Get("articles.show").URL("id", strconv.FormatInt(a.ID, 10))
	if err != nil {
		logger.LogError(err)
		return ""
	}
	return showURL.String()
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

	router.HandleFunc("/articles",
		articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles",
		articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create",
		articlesCreateHandler).Methods("GET").Name("articles.create")
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
