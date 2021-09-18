package controllers

import (
	"fmt"
	"goblog/app/models/user"
	"goblog/app/requests"
	"goblog/pck/auth"
	"goblog/pck/flash"
	"goblog/pck/view"
	"net/http"
)

// AuthController 处理静态页面
type AuthController struct {
}

type userForm struct {
	Name            string `valid:"name"`
	Email           string `valid:"email"`
	Password        string `valid:"password"`
	PasswordConfirm string `valid:"password_confirm"`
}

func (*AuthController) Register(w http.ResponseWriter, r *http.Request) {
	view.RenderSimple(w, view.D{}, "auth.register")
}

func (*AuthController) DoRegister(w http.ResponseWriter, r *http.Request) {

	_user := user.User{
		Name:            r.PostFormValue("name"),
		Email:           r.PostFormValue("email"),
		Password:        r.PostFormValue("password"),
		PasswordConfirm: r.PostFormValue("password_confirm"),
	}

	// 表单规则
	errs := requests.ValidateRegistrationForm(_user)

	if len(errs) > 0 {
		view.RenderSimple(w, view.D{
			"Errors":	errs,
			"User":		_user,
		}, "auth.register")
	} else {
		// 验证成功
		_user.Create()

		if _user.ID > 0 {
			flash.Success("恭喜您注册成功！")
			auth.Login(_user)
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "注册失败，请联系管理员")
		}
	}
}

func (*AuthController) Login(w http.ResponseWriter, r *http.Request) {
	view.RenderSimple(w, view.D{}, "auth.login")
}

func (*AuthController) DoLogin(w http.ResponseWriter, r *http.Request) {

	// 初始化表单
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	if err := auth.Attempt(email, password); err == nil {
		flash.Success("欢迎回来！")
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		view.RenderSimple(w, view.D{
			"Error": 	err.Error(),
			"Email":	email,
			"Password":	password,
		}, "auth.login")
	}
}

func (*AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	auth.Logout()
	flash.Success("您已成功退出登录")
	http.Redirect(w, r, "/", http.StatusFound)
}
