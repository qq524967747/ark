package db

import (
	"gorm.io/driver/sqlite" // Sqlite driver based on GGO
	"gorm.io/gorm"
	"qq-bot/utils"
)

var (
	sqliteClient *gorm.DB
)

func InitSqlite() {
	// github.com/mattn/go-sqlite3
	db, err := gorm.Open(sqlite.Open("D:\\data\\ark-data\\arkDB.db"), &gorm.Config{})
	utils.PanicNotNil(err)
	dbClient = db
}

func SqlClient() *gorm.DB {
	return sqliteClient
}
