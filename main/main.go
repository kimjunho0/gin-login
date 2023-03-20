package main

import (
	"gin-login/migrate"
	"github.com/gin-gonic/gin"
)

func main() {
	migrate.ConnectDB()
	r := gin.Default()

}
