package my

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Migrate program
func Migrate() {
	db, er := gorm.Open(sqlite.Open("data.sqlite3"), &gorm.Config{})
	if er != nil {
		fmt.Println(er)
		return
	}

	db.AutoMigrate(&User{}, &Group{}, &Post{}, &Comment{})

}
