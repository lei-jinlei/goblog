package controllers

import (
	"database/sql"
	"fmt"
	"goblog/pck/logger"
	"goblog/pck/route"
	"goblog/pck/types"
	"net/http"
)

// ArticlesController 处理静态页面
type ArticlesController struct {
}

func (*ArticlesController) show(w http.ResponseWriter, r *http.Request) {
	// 获取 URL 参数
	id := route.GetRouteVariable("id", r)

	// 读取相应的文章数据
	article, err := getArticleByID(id)

	// 如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
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
