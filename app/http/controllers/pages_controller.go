package controllers

import (
	"fmt"
	"goblog/pck/view"
	"net/http"
)

// PagesController 处理静态页面
type PagesController struct {
}

// Home 首页
func (*PagesController) Home(w http.ResponseWriter, r *http.Request)  {
	view.Render(w, nil, "pages.index")
}

// About 关于我们页面
func (*PagesController) About(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprint(w, "此博客是用以记录编程笔记，如您有反馈或建议")
}

// notFound 404 页面
func (*PagesController) NotFound(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1><p>如有疑惑，请联系我们。</p>")
}