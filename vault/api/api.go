package api

import (
	"time"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run() {
	api := InitializeRouter()
	err := api.Run(":" + config.Port)
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to start server: %v", err)
	}
}

func InitializeRouter() *gin.Engine {
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))
	InitializeRoutes(r)
	return r
}

func InitializeRoutes(router *gin.Engine) {
	router.GET("/vault/ping", Ping)

	router.GET("/vault/accounts", ListAccounts)
	router.POST("/vault/accounts", CreateAccount)
	router.GET("/vault/accounts/:id", GetAccount)
	router.PUT("/vault/accounts/:id", UpdateAccount)
	router.DELETE("/vault/accounts/:id", ArchiveAccount)

	router.GET("/vault/accounts/:id/secrets", ListSecrets)
	router.POST("/vault/accounts/:id/secrets", CreateSecret)
	router.GET("/vault/accounts/:id/secrets/:secretID", GetSecret)
	router.PUT("/vault/accounts/:id/secrets/:secretID", UpdateSecret)
	router.DELETE("/vault/accounts/:id/secrets/:secretID", ArchiveSecret)
	router.POST("/vault/accounts/:id/secrets/:secretID/reveal", RevealSecret)
}
