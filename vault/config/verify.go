package config

import (
	"github.com/gaucho-racing/vault/vault/pkg/githubactions"
	"github.com/gaucho-racing/vault/vault/pkg/kubernetes"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
)

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
	if SentinelURL == "" {
		logger.SugarLogger.Fatal("SENTINEL_URL is required")
	}
	if SentinelClientID == "" {
		logger.SugarLogger.Fatal("SENTINEL_CLIENT_ID is required")
	}
	if SentinelClientSecret == "" {
		logger.SugarLogger.Fatal("SENTINEL_CLIENT_SECRET is required")
	}
	if GitHubActionsOIDCIssuer == "" {
		GitHubActionsOIDCIssuer = githubactions.DefaultIssuer
		logger.SugarLogger.Infof("GITHUB_ACTIONS_OIDC_ISSUER is not set, defaulting to %s", GitHubActionsOIDCIssuer)
	}
	if GitHubActionsOIDCAudience == "" {
		GitHubActionsOIDCAudience = githubactions.DefaultAudience
		logger.SugarLogger.Infof("GITHUB_ACTIONS_OIDC_AUDIENCE is not set, defaulting to %s", GitHubActionsOIDCAudience)
	}
	if KubernetesOIDCAudience == "" {
		KubernetesOIDCAudience = kubernetes.DefaultAudience
		logger.SugarLogger.Infof("KUBERNETES_OIDC_AUDIENCE is not set, defaulting to %s", KubernetesOIDCAudience)
	}
	if KubernetesClusterID == "" {
		KubernetesClusterID = "default"
		logger.SugarLogger.Infof("KUBERNETES_CLUSTER_ID is not set, defaulting to %s", KubernetesClusterID)
	}
}
