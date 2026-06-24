package main

import (
	"github.com/gaucho-racing/vault/vault/api"
	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
)

func main() {
	logger.Init(config.IsProduction())
	defer logger.Logger.Sync()

	config.Verify()
	config.PrintStartupBanner()

	api.Run()
}
