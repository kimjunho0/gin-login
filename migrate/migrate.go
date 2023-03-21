package migrate

import (
	"gin-login/models"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
)

var DB *gorm.DB
var err error

func ConnectDB() {
	dsn := "ginlogin:@(localhost)/ginlogin?parseTime=True&loc=Asia%2FSeoul"
	DB, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		log.Fatal("DB Connect Failed")
	}

	createTables(DB)

}
func createTables(DB *gorm.DB) {
	tables := []interface{}{
		(*models.User)(nil),
	}

	if err := DB.AutoMigrate(tables...); err != nil {
		panic(http.StatusBadRequest)
	}
}
