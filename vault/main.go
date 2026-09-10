package main

import (
	"github.com/gaucho-racing/vault/vault/api"
	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"github.com/gaucho-racing/vault/vault/pkg/sentinel"
	"github.com/gaucho-racing/vault/vault/service"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()
	if err := sentinel.InitializeSigningKeys(); err != nil {
		logger.SugarLogger.Warnf("initialize Sentinel signing keys: %v", err)
	}
	database.Init()
	service.InitializeVaultKeys()

	api.Run()
}
