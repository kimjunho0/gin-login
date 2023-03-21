package migrate

import (
	"gin-login/models"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"net/http"
)

var DB *gorm.DB
var err error

func ConnectDB() {
	dsn := "ginlogin:@(localhost)/ginlogin?parseTime=True&loc=Asia%2FSeoul"

	dbConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	DB, err = gorm.Open(mysql.Open(dsn), dbConfig)
	if err != nil {
		log.Fatal("DB Connect Failed")
	}
	//replica 생성
	readDsn := "ginlogin:@(localhost)/ginlogin?parseTime=True"

	if readDnsConnectionErr := DB.Use(dbresolver.Register(dbresolver.Config{
		Replicas: []gorm.Dialector{mysql.Open(readDsn)},
	})); readDnsConnectionErr != nil {
		panic("replica error")
	}
	//커넥션풀 생성

	sqlDB, connPoolErr := DB.DB()

	if connPoolErr != nil {
		panic("connection pool error")
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)

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
