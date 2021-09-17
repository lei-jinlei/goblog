package user

import (
	"goblog/pck/logger"
	"goblog/pck/model"
	"goblog/pck/types"
)

func (user *User) Create() (err error) {
	if err = model.DB.Create(&user).Error; err != nil {
		logger.LogError(err)
		return err
	}
	return nil
}

func GetByEmail(email string) (_user User,err error) {
	if err := model.DB.Where("email = ?", email).
		First(&_user).Error; err != nil {
		return _user, err
	}
	return _user, nil
}

func Get(uid string) (_user User, err error) {
	id := types.StringToInt(uid)
	if err := model.DB.First(&_user, id).Error; err != nil {
		return _user, err
	}
	return _user, nil
}
