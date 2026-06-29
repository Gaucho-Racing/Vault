package api

import (
	"errors"
	"net/http"

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

type kubernetesClusterRequest struct {
	Name     string `json:"name" binding:"required"`
	Issuer   string `json:"issuer" binding:"required"`
	Audience string `json:"audience"`
	Enabled  bool   `json:"enabled"`
}

type kubernetesSecretRuleRequest struct {
	Name                   string   `json:"name" binding:"required"`
	ClusterIDs             []string `json:"cluster_ids"`
	NamespacePatterns      []string `json:"namespace_patterns"`
	ServiceAccountPatterns []string `json:"service_account_patterns"`
	SecretSelectors        []string `json:"secret_selectors"`
	Enabled                bool     `json:"enabled"`
}

func ExportKubernetesSecrets(c *gin.Context) {
	var req kubernetesAppSecretsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	identity, err := service.VerifyKubernetesToken(c.Request.Context(), req.Token)
	if err != nil {
		if errors.Is(err, kubernetes.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrKubernetesClusterNotTrusted) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	secrets, err := service.BuildKubernetesSecretData(service.KubernetesExportRequest{
		Secrets: req.Secrets,
		Claims:  identity.Claims,
		Cluster: identity.Cluster,
	})
	if err != nil {
		handleKubernetesSecretExportError(c, err)
		return
	}
	c.JSON(http.StatusOK, kubernetesAppSecretsResponse{Secrets: secrets})
}

func ListKubernetesClusters(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	clusters, err := service.GetAllKubernetesClusters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, clusters)
}

func CreateKubernetesCluster(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	var req kubernetesClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cluster, err := service.CreateKubernetesCluster(modelKubernetesCluster(req, GetRequestEntityID(c), GetRequestEntityID(c)))
	if err != nil {
		handleKubernetesClusterError(c, err)
		return
	}
	c.JSON(http.StatusOK, cluster)
}

func UpdateKubernetesCluster(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	cluster, err := service.GetKubernetesClusterByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "kubernetes cluster not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req kubernetesClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cluster.Name = req.Name
	cluster.Issuer = req.Issuer
	cluster.Audience = req.Audience
	cluster.Enabled = req.Enabled
	cluster.UpdatedByEntityID = GetRequestEntityID(c)

	updated, err := service.UpdateKubernetesCluster(cluster)
	if err != nil {
		handleKubernetesClusterError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteKubernetesCluster(c *gin.Context) {
	Require(c, RequestTokenCanManageKubernetesSecretRules(c))
	if err := service.DeleteKubernetesCluster(c.Param("id")); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "kubernetes cluster not found"})
			return
		}
		if errors.Is(err, service.ErrKubernetesClusterInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "kubernetes cluster deleted"})
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
	rule.ClusterIDs = req.ClusterIDs
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
		errors.Is(err, service.ErrKubernetesClusterRequired) ||
		errors.Is(err, service.ErrKubernetesClusterInvalid) ||
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

func handleKubernetesClusterError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrKubernetesClusterNameRequired) ||
		errors.Is(err, service.ErrKubernetesClusterNameInvalid) ||
		errors.Is(err, service.ErrKubernetesClusterIssuerRequired) ||
		errors.Is(err, service.ErrKubernetesClusterIssuerInvalid) ||
		errors.Is(err, service.ErrKubernetesClusterAudienceInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"error": "kubernetes cluster name or issuer already exists"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func modelKubernetesCluster(req kubernetesClusterRequest, createdBy string, updatedBy string) model.KubernetesCluster {
	return model.KubernetesCluster{
		Name:              req.Name,
		Issuer:            req.Issuer,
		Audience:          req.Audience,
		Enabled:           req.Enabled,
		CreatedByEntityID: createdBy,
		UpdatedByEntityID: updatedBy,
	}
}

func modelKubernetesSecretRule(req kubernetesSecretRuleRequest, createdBy string, updatedBy string) model.KubernetesSecretRule {
	return model.KubernetesSecretRule{
		Name:                   req.Name,
		ClusterIDs:             req.ClusterIDs,
		NamespacePatterns:      req.NamespacePatterns,
		ServiceAccountPatterns: req.ServiceAccountPatterns,
		SecretSelectors:        req.SecretSelectors,
		Enabled:                req.Enabled,
		CreatedByEntityID:      createdBy,
		UpdatedByEntityID:      updatedBy,
	}
}
