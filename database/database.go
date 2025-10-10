package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"fmt"
	"log"
	"os"
)

func InitDatabase() *gorm.DB {
	dbPssw := os.Getenv("MYSQL_ROOT_PASSWORD")

	dsn := fmt.Sprintf("root:%s@tcp(user_database:3306)/user_api?charset=utf8&parseTime=true", dbPssw)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("Failed to connect to database: %s", err)
	}

	return db
}
