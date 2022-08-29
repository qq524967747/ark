package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"qq-bot/config"
	"qq-bot/utils"
)

var (
	dbClient *gorm.DB
)

func InitMysql() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Config.Database.Mysql.User,
		config.Config.Database.Mysql.Pwd,
		config.Config.Database.Mysql.Hostname,
		config.Config.Database.Mysql.Port,
		config.Config.Database.Mysql.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	utils.PanicNotNil(err)
	dbClient = db
}

func Client() *gorm.DB {
	return dbClient
}
