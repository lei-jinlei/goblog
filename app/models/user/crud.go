package user

import (
	"goblog/pck/logger"
	"goblog/pck/model"
)

func (user *User) Create() (err error) {
	if err = model.DB.Create(&user).Error; err != nil {
		logger.LogError(err)
		return err
	}
	return nil
}
