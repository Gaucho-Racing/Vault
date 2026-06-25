package config

import (
	"encoding/base64"

	"github.com/gaucho-racing/vault/vault/pkg/logger"
)

const developmentVaultMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func Verify() {
	if Env == "" {
		Env = "PROD"
		logger.SugarLogger.Infof("ENV is not set, defaulting to %s", Env)
	}
	if Port == "" {
		Port = "9994"
		logger.SugarLogger.Infof("PORT is not set, defaulting to %s", Port)
	}
	if DatabaseHost == "" {
		DatabaseHost = "localhost"
		logger.SugarLogger.Infof("DATABASE_HOST is not set, defaulting to %s", DatabaseHost)
	}
	if DatabasePort == "" {
		DatabasePort = "5432"
		logger.SugarLogger.Infof("DATABASE_PORT is not set, defaulting to %s", DatabasePort)
	}
	if DatabaseUser == "" {
		DatabaseUser = "postgres"
		logger.SugarLogger.Infof("DATABASE_USER is not set, defaulting to %s", DatabaseUser)
	}
	if DatabasePassword == "" {
		DatabasePassword = "password"
		logger.SugarLogger.Infof("DATABASE_PASSWORD is not set, defaulting to %s", DatabasePassword)
	}
	if DatabaseName == "" {
		DatabaseName = "vault"
		logger.SugarLogger.Infof("DATABASE_NAME is not set, defaulting to %s", DatabaseName)
	}
	if VaultMasterKey == "" {
		if IsProduction() {
			logger.SugarLogger.Fatal("VAULT_MASTER_KEY is required in production")
		}
		VaultMasterKey = developmentVaultMasterKey
		logger.SugarLogger.Infof("VAULT_MASTER_KEY is not set, defaulting to development key")
	}
	decodedKey, err := base64.StdEncoding.DecodeString(VaultMasterKey)
	if err != nil {
		logger.SugarLogger.Fatalf("VAULT_MASTER_KEY must be base64 encoded: %v", err)
	}
	if len(decodedKey) != 32 {
		logger.SugarLogger.Fatalf("VAULT_MASTER_KEY must decode to 32 bytes, got %d", len(decodedKey))
	}
	if VaultMasterKeyID == "" {
		VaultMasterKeyID = "local"
		logger.SugarLogger.Infof("VAULT_MASTER_KEY_ID is not set, defaulting to %s", VaultMasterKeyID)
	}
}
