package config

import "github.com/gaucho-racing/vault/vault/pkg/logger"

func Verify() {
	if Env == "" {
		Env = "PROD"
		logger.SugarLogger.Infof("ENV is not set, defaulting to %s", Env)
	}
	if Port == "" {
		Port = "9994"
		logger.SugarLogger.Infof("PORT is not set, defaulting to %s", Port)
	}
}
