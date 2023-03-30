package service

import (
	"context"
	"gin-login/docs"
	"gin-login/internal/handler/auth/login"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/redis"
	"gin-login/tools"
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
	tools.InitSentry(
		"local",
		"https://80c4f993222946e4b2fa01f5db4e327f@o4504920735612928.ingest.sentry.io/4504920736530432",
		0.1,
		"1.0",
	)

	// mysql 연동
	migrate.ConnectDB()
	//migrate.DB.Select("id,")

	// redis 연동
	redis.Connect()

	// gin framework
	r := gin.New()
	r.Use(gin.Logger())

	// production 환경일시
	//gin.SetMode(gin.ReleaseMode)

	//swagger
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Use(middleware.CorsMiddleware)

	rAPI := r.Group("/api")

	rAPI.Use(middleware.WrapMiddleware)

	rAuth := rAPI.Group("/auth")
	{
		rAuth.POST("/register", login.Register)
		rAuth.POST("/login", login.Login)
		rAuth.PATCH("/reset-password/:num", login.ResetPassword)
		rAuth.POST("/logout", login.Logout)
		rAuth.DELETE("/delete", login.Delete)
		//rAuth.DELETE(fmt.Sprintf("/leave/:%s", "10"),auth.Leave)
		rAuth.POST("/refresh-token", login.RefreshAccessToken)
		rAuth.GET("info", login.Info)
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
