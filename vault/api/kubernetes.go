package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/kubernetes"
	"github.com/gaucho-racing/vault/vault/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type kubernetesAppSecretsRequest struct {
	Token   string            `json:"token" binding:"required"`
	Secrets map[string]string `json:"secrets" binding:"required"`
}

type kubernetesAppSecretsResponse struct {
	Secrets map[string]string `json:"secrets"`
}

type kubernetesSecretRuleRequest struct {
	Name                   string   `json:"name" binding:"required"`
	ClusterPatterns        []string `json:"cluster_patterns"`
	NamespacePatterns      []string `json:"namespace_patterns"`
	ServiceAccountPatterns []string `json:"service_account_patterns"`
	SecretSelectors        []string `json:"secret_selectors"`
	Enabled                bool     `json:"enabled"`
}

var kubernetesVerifier = kubernetes.NewVerifier(config.KubernetesOIDCIssuer, config.KubernetesOIDCAudience, nil)

func ExportKubernetesSecrets(c *gin.Context) {
	var req kubernetesAppSecretsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := kubernetesVerifier.Verify(c.Request.Context(), req.Token)
	if err != nil {
		if errors.Is(err, kubernetes.ErrNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, kubernetes.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	secrets, err := service.BuildKubernetesSecretData(service.KubernetesExportRequest{
		Secrets: req.Secrets,
		Claims:  claims,
	})
	if err != nil {
		handleKubernetesSecretExportError(c, err)
		return
	}
	c.JSON(http.StatusOK, kubernetesAppSecretsResponse{Secrets: secrets})
}

func ListKubernetesSecretRules(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	rules, err := service.GetAllKubernetesSecretRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func CreateKubernetesSecretRule(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	var req kubernetesSecretRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := service.CreateKubernetesSecretRule(modelKubernetesSecretRule(req, GetRequestEntityID(c), GetRequestEntityID(c)))
	if err != nil {
		handleKubernetesSecretRuleError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

func UpdateKubernetesSecretRule(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	rule, err := service.GetKubernetesSecretRuleByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "kubernetes secret rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req kubernetesSecretRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule.Name = req.Name
	rule.ClusterPatterns = req.ClusterPatterns
	rule.NamespacePatterns = req.NamespacePatterns
	rule.ServiceAccountPatterns = req.ServiceAccountPatterns
	rule.SecretSelectors = req.SecretSelectors
	rule.Enabled = req.Enabled
	rule.UpdatedByEntityID = GetRequestEntityID(c)

	updated, err := service.UpdateKubernetesSecretRule(rule)
	if err != nil {
		handleKubernetesSecretRuleError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteKubernetesSecretRule(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	if err := service.DeleteKubernetesSecretRule(c.Param("id")); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "kubernetes secret rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "kubernetes secret rule deleted"})
}

func handleKubernetesSecretExportError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "app secret not found"})
		return
	}
	if errors.Is(err, service.ErrKubernetesSecretSelectorNotAllowed) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, service.ErrKubernetesSecretSelectorRequired) ||
		errors.Is(err, service.ErrKubernetesRequestedSecretSelectorInvalid) ||
		errors.Is(err, service.ErrKubernetesSecretKeyRequired) ||
		errors.Is(err, service.ErrKubernetesSecretKeyInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func handleKubernetesSecretRuleError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrKubernetesSecretRuleNameRequired) ||
		errors.Is(err, service.ErrKubernetesSecretRuleNameInvalid) ||
		errors.Is(err, service.ErrKubernetesClusterPatternRequired) ||
		errors.Is(err, service.ErrKubernetesClusterPatternInvalid) ||
		errors.Is(err, service.ErrKubernetesNamespacePatternRequired) ||
		errors.Is(err, service.ErrKubernetesNamespacePatternInvalid) ||
		errors.Is(err, service.ErrKubernetesServiceAccountPatternRequired) ||
		errors.Is(err, service.ErrKubernetesServiceAccountPatternInvalid) ||
		errors.Is(err, service.ErrKubernetesSecretSelectorRequired) ||
		errors.Is(err, service.ErrKubernetesSecretSelectorInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"error": "kubernetes secret rule name already exists"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func modelKubernetesSecretRule(req kubernetesSecretRuleRequest, createdBy string, updatedBy string) model.KubernetesSecretRule {
	return model.KubernetesSecretRule{
		Name:                   req.Name,
		ClusterPatterns:        req.ClusterPatterns,
		NamespacePatterns:      req.NamespacePatterns,
		ServiceAccountPatterns: req.ServiceAccountPatterns,
		SecretSelectors:        req.SecretSelectors,
		Enabled:                req.Enabled,
		CreatedByEntityID:      createdBy,
		UpdatedByEntityID:      updatedBy,
	}
}
