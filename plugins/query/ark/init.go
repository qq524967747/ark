package ark

import (
	"qq-bot/database/db"
)

// 初始化数据库

func InitDB() {
	if err := db.Client().AutoMigrate(&ArkShopPlayer{}); err != nil {
		panic(err)
	}
	// todo 更新权限

}
