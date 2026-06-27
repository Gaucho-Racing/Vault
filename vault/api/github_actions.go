package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/githubactions"
	"github.com/gaucho-racing/vault/vault/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type githubActionsAppSecretsRequest struct {
	Token   string   `json:"token" binding:"required"`
	Secrets []string `json:"secrets" binding:"required"`
}

var githubActionsVerifier = githubactions.NewVerifier(config.GitHubActionsOIDCIssuer, config.GitHubActionsOIDCAudience, nil)

type githubActionsRuleRequest struct {
	Name               string   `json:"name" binding:"required"`
	RepositoryPatterns []string `json:"repository_patterns"`
	RefPatterns        []string `json:"ref_patterns"`
	SecretSelectors    []string `json:"secret_selectors"`
	Enabled            bool     `json:"enabled"`
}

func ExportGitHubActionsEnv(c *gin.Context) {
	var req githubActionsAppSecretsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	envFile, err := service.BuildGitHubActionsEnvFile(req.Secrets, claims)
	if err != nil {
		handleGitHubActionsExportError(c, err)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(envFile))
}

func ListGitHubActionsRules(c *gin.Context) {
	Require(c, RequestTokenCanManageGitHubActionsRules(c))
	rules, err := service.GetAllGitHubActionsRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func CreateGitHubActionsRule(c *gin.Context) {
	Require(c, RequestTokenCanManageGitHubActionsRules(c))
	var req githubActionsRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := service.CreateGitHubActionsRule(modelGitHubActionsRule(req, GetRequestEntityID(c), GetRequestEntityID(c)))
	if err != nil {
		handleGitHubActionsRuleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

func UpdateGitHubActionsRule(c *gin.Context) {
	Require(c, RequestTokenCanManageGitHubActionsRules(c))
	rule, err := service.GetGitHubActionsRuleByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "github actions rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req githubActionsRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.Name = req.Name
	rule.RepositoryPatterns = req.RepositoryPatterns
	rule.RefPatterns = req.RefPatterns
	rule.SecretSelectors = req.SecretSelectors
	rule.Enabled = req.Enabled
	rule.UpdatedByEntityID = GetRequestEntityID(c)

	updated, err := service.UpdateGitHubActionsRule(rule)
	if err != nil {
		handleGitHubActionsRuleError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteGitHubActionsRule(c *gin.Context) {
	Require(c, RequestTokenCanManageGitHubActionsRules(c))
	if err := service.DeleteGitHubActionsRule(c.Param("id")); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "github actions rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "github actions rule deleted"})
}

func handleGitHubActionsExportError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "app secret not found"})
		return
	}
	if errors.Is(err, service.ErrGitHubActionsSecretSelectorNotAllowed) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, service.ErrGitHubActionsSecretSelectorRequired) ||
		errors.Is(err, service.ErrGitHubActionsSecretSelectorInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, service.ErrGitHubActionsEnvNameCollision) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func handleGitHubActionsRuleError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrGitHubActionsRuleNameRequired) ||
		errors.Is(err, service.ErrGitHubActionsRuleNameInvalid) ||
		errors.Is(err, service.ErrGitHubActionsRepositoryPatternRequired) ||
		errors.Is(err, service.ErrGitHubActionsRepositoryPatternInvalid) ||
		errors.Is(err, service.ErrGitHubActionsRefPatternRequired) ||
		errors.Is(err, service.ErrGitHubActionsRefPatternInvalid) ||
		errors.Is(err, service.ErrGitHubActionsSecretSelectorRequired) ||
		errors.Is(err, service.ErrGitHubActionsSecretSelectorInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"error": "github actions rule name already exists"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func modelGitHubActionsRule(req githubActionsRuleRequest, createdBy string, updatedBy string) model.GitHubActionsRule {
	return model.GitHubActionsRule{
		Name:               req.Name,
		RepositoryPatterns: req.RepositoryPatterns,
		RefPatterns:        req.RefPatterns,
		SecretSelectors:    req.SecretSelectors,
		Enabled:            req.Enabled,
		CreatedByEntityID:  createdBy,
		UpdatedByEntityID:  updatedBy,
	}
}
