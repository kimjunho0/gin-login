package service

import (
	"gin-login/docs"
	"gin-login/internal/handler/auth"
	"gin-login/middleware"
	"gin-login/migrate"
	"gin-login/redis"
	"gin-login/tools"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
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
		rAuth.POST("/register", auth.Register)
		rAuth.POST("/login", auth.Login)
		rAuth.PATCH("/reset-password/:num", auth.ResetPassword)
		rAuth.POST("/logout", auth.Logout)
		rAuth.DELETE("/delete", auth.Delete)
		//rAuth.DELETE(fmt.Sprintf("/leave/:%s", "10"),auth.Leave)
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

	//Graceful shutdown
	//처리중이던 요청들이 모두 처리된 뒤에 종료가 되도록 하는 tool
	tools.WaitForShutdown(srv)
}
