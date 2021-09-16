package controllers

import (
	"goblog/pck/view"
	"net/http"
)

// AuthController 处理静态页面
type AuthController struct {
}

func (*AuthController) Register(w http.ResponseWriter, r *http.Request) {
	view.RenderSimple(w, view.D{}, "auth.register")
}

func (*AuthController) DoRegister(w http.ResponseWriter, r *http.Request) {

}