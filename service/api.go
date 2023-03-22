package service

import (
	"context"
	"gin-login/docs"
	"gin-login/internal/handler/auth"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/redis"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title Swagger gin-login
// @version 1.0
// @description This is a sample server to dooluck
// @BasePath /

func Run() {
	migrate.ConnectDB()
	redis.Connect()
	r := gin.New()
	migrate.DB.Select("id,")
	r.Use(gin.Logger())
	gin.SetMode(gin.ReleaseMode)

	//swagger
	docs.SwaggerInfo.BasePath = "/"

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Use(middleware.CorsMiddleware)
	rAPI := r.Group("/api")

	rAuth := rAPI.Group("/auth")
	{
		rAuth.POST("/register", auth.Register)
		rAuth.POST("/login", auth.Login)
		rAuth.POST("/reset-password", auth.ResetPassword)
		rAuth.POST("/logout", auth.Logout)
		rAuth.POST("/leave", auth.Leave)
		rAuth.POST("/refresh-token", auth.RefreshAccessToken)
		rAuth.GET("info", auth.Info)
	}

	//서버 시작
	srv := &http.Server{
		Handler:      r,
		Addr:         ":5050",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		baseUrl := "http://localhost:5050"
		log.Printf("Server listen %s\n", baseUrl)
		log.Printf("Now you can check api docs %s/swagger/index.html", baseUrl)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	WaitForShutdown(srv)
}

func WaitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	//

	<-interruptChan

	// channel 서버 꺼질때까지 기다려주기
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		return
	}
	log.Println("Shutting down")
	os.Exit(0)

}
