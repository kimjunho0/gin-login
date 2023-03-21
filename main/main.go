package main

import (
	"gin-login/internal/handler/auth"
	"gin-login/migrate"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func main() {
	migrate.ConnectDB()
	r := gin.New()

	r.Use(gin.Logger())
	gin.SetMode(gin.ReleaseMode)
	rAPI := r.Group("/api")

	rAuth := rAPI.Group("/auth")
	{
		rAuth.POST("/register", auth.Register)
	}

	//서버 시작
	srv := &http.Server{
		Handler:      r,
		Addr:         ":5050",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
