package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/pkg/githubactions"
	"github.com/gaucho-racing/vault/vault/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type githubActionsAppSecretsRequest struct {
	Token string `json:"token" binding:"required"`
}

var githubActionsVerifier = githubactions.NewVerifier(config.GitHubActionsOIDCIssuer, config.GitHubActionsOIDCAudience, nil)

func ExportGitHubActionsApplicationEnv(c *gin.Context) {
	var req githubActionsAppSecretsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	claims, err := githubActionsVerifier.Verify(c.Request.Context(), req.Token)
	if err != nil {
		if errors.Is(err, githubactions.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	envFile, err := service.BuildApplicationGitHubActionsEnvFile(c.Param("name"), claims)
	if err != nil {
		handleGitHubActionsExportError(c, err)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(envFile))
}

func handleGitHubActionsExportError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}
	if errors.Is(err, service.ErrGitHubActionsRepositoryNotAllowed) ||
		errors.Is(err, service.ErrGitHubActionsRefNotAllowed) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, service.ErrGitHubActionsEnvNameCollision) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
