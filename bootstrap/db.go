package bootstrap

import (
	"goblog/app/models/article"
	"goblog/app/models/user"
	"goblog/pck/config"
	"goblog/pck/model"
	"gorm.io/gorm"
	"time"
)

func SetupDB() {

	// 建立数据库连接池
	db := model.ConnectDB()

	// 命令行打印数据库请求信息
	sqlDB, _ := db.DB()

	// 设置最大连接数
	sqlDB.SetMaxOpenConns(config.GetInt("database.mysql.max_open_connections"))
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(config.GetInt("database.mysql.max_idle_connections"))
	// 设置每个链接的过期时间
	sqlDB.SetConnMaxLifetime(time.Duration(config.GetInt("database.mysql.max_life_seconds")) * time.Second)

	// 创建和维护数据表结构
	migration(db)
}

func migration(db *gorm.DB)  {

	// 自动迁移
	db.AutoMigrate(
		&user.User{},
		&article.Article{},
	)
}
