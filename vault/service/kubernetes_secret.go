package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/kubernetes"
	"gorm.io/gorm"
)

var ErrKubernetesSecretRuleNameRequired = errors.New("kubernetes secret rule name is required")
var ErrKubernetesSecretRuleNameInvalid = errors.New("kubernetes secret rule name must contain only lowercase letters, numbers, hyphens, and underscores")
var ErrKubernetesClusterRequired = errors.New("kubernetes cluster is required")
var ErrKubernetesClusterInvalid = errors.New("kubernetes cluster is invalid")
var ErrKubernetesNamespacePatternRequired = errors.New("kubernetes namespace pattern is required")
var ErrKubernetesNamespacePatternInvalid = errors.New("kubernetes namespace pattern must contain only lowercase letters, numbers, hyphens, and wildcards")
var ErrKubernetesServiceAccountPatternRequired = errors.New("kubernetes service account pattern is required")
var ErrKubernetesServiceAccountPatternInvalid = errors.New("kubernetes service account pattern must contain only lowercase letters, numbers, hyphens, and wildcards")
var ErrKubernetesSecretSelectorRequired = errors.New("kubernetes secret selector is required")
var ErrKubernetesSecretSelectorInvalid = errors.New("kubernetes secret selector must use application.secret format and may include wildcards")
var ErrKubernetesRequestedSecretSelectorInvalid = errors.New("kubernetes requested secret selector must use explicit application.secret format")
var ErrKubernetesSecretSelectorNotAllowed = errors.New("kubernetes secret selector is not allowed")
var ErrKubernetesSecretKeyRequired = errors.New("kubernetes secret key is required")
var ErrKubernetesSecretKeyInvalid = errors.New("kubernetes secret key must contain only letters, numbers, hyphens, underscores, and dots")

type KubernetesExportRequest struct {
	Secrets map[string]string
	Claims  kubernetes.Claims
	Cluster model.KubernetesCluster
}

var kubernetesNamespacePattern = regexp.MustCompile(`^[a-z0-9*][a-z0-9*-]*$`)
var kubernetesServiceAccountPattern = regexp.MustCompile(`^[a-z0-9*][a-z0-9*-]*$`)
var kubernetesSecretKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func GetAllKubernetesSecretRules() ([]model.KubernetesSecretRule, error) {
	rules := []model.KubernetesSecretRule{}
	if err := database.DB.Order("name ASC").Find(&rules).Error; err != nil {
		return []model.KubernetesSecretRule{}, err
	}
	return rules, nil
}

func GetKubernetesSecretRuleByID(id string) (model.KubernetesSecretRule, error) {
	var rule model.KubernetesSecretRule
	if err := database.DB.Where("id = ?", id).First(&rule).Error; err != nil {
		return model.KubernetesSecretRule{}, err
	}
	return rule, nil
}

func CreateKubernetesSecretRule(rule model.KubernetesSecretRule) (model.KubernetesSecretRule, error) {
	if rule.ID == "" {
		rule.ID = ulid.Make().Prefixed("k8srule")
	}
	if err := normalizeKubernetesSecretRule(&rule); err != nil {
		return model.KubernetesSecretRule{}, err
	}
	if err := database.DB.Create(&rule).Error; err != nil {
		return model.KubernetesSecretRule{}, err
	}
	return rule, nil
}

func UpdateKubernetesSecretRule(rule model.KubernetesSecretRule) (model.KubernetesSecretRule, error) {
	if err := normalizeKubernetesSecretRule(&rule); err != nil {
		return model.KubernetesSecretRule{}, err
	}
	if err := database.DB.Save(&rule).Error; err != nil {
		return model.KubernetesSecretRule{}, err
	}
	return rule, nil
}

func DeleteKubernetesSecretRule(id string) error {
	result := database.DB.Where("id = ?", id).Delete(&model.KubernetesSecretRule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func BuildKubernetesSecretData(req KubernetesExportRequest) (map[string]string, error) {
	normalizedSecrets, err := normalizeRequestedKubernetesSecrets(req.Secrets)
	if err != nil {
		return map[string]string{}, err
	}

	rules := []model.KubernetesSecretRule{}
	if err := database.DB.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return map[string]string{}, err
	}

	matchingRules := make([]model.KubernetesSecretRule, 0, len(rules))
	for _, rule := range rules {
		if kubernetesSecretRuleMatchesClaims(rule, req.Cluster, req.Claims) {
			matchingRules = append(matchingRules, rule)
		}
	}

	data := make(map[string]string, len(normalizedSecrets))
	for key, selector := range normalizedSecrets {
		if !kubernetesSelectorAllowed(matchingRules, selector) {
			return map[string]string{}, fmt.Errorf("%w: %s", ErrKubernetesSecretSelectorNotAllowed, selector)
		}
		secret, err := getAppSecretBySelector(selector)
		if err != nil {
			return map[string]string{}, err
		}
		value, err := RevealAppSecret(secret)
		if err != nil {
			return map[string]string{}, fmt.Errorf("reveal app secret %s: %w", selector, err)
		}
		data[key] = value
	}
	return data, nil
}

func normalizeKubernetesSecretRule(rule *model.KubernetesSecretRule) error {
	rule.Name = strings.ToLower(strings.TrimSpace(rule.Name))
	rule.ClusterIDs = normalizeStringSlice(rule.ClusterIDs)
	rule.NamespacePatterns = normalizeLowerStringSlice(rule.NamespacePatterns)
	rule.ServiceAccountPatterns = normalizeLowerStringSlice(rule.ServiceAccountPatterns)
	rule.SecretSelectors = normalizeGitHubActionsSecretSelectors(rule.SecretSelectors)
	if rule.Name == "" {
		return ErrKubernetesSecretRuleNameRequired
	}
	if !appSecretIdentifierPattern.MatchString(rule.Name) {
		return ErrKubernetesSecretRuleNameInvalid
	}
	if len(rule.ClusterIDs) == 0 {
		return ErrKubernetesClusterRequired
	}
	if len(rule.NamespacePatterns) == 0 {
		return ErrKubernetesNamespacePatternRequired
	}
	if len(rule.ServiceAccountPatterns) == 0 {
		return ErrKubernetesServiceAccountPatternRequired
	}
	if len(rule.SecretSelectors) == 0 {
		return ErrKubernetesSecretSelectorRequired
	}
	for _, clusterID := range rule.ClusterIDs {
		if _, err := GetKubernetesClusterByID(clusterID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrKubernetesClusterInvalid
			}
			return err
		}
	}
	for _, pattern := range rule.NamespacePatterns {
		if !kubernetesNamespacePattern.MatchString(pattern) {
			return ErrKubernetesNamespacePatternInvalid
		}
	}
	for _, pattern := range rule.ServiceAccountPatterns {
		if !kubernetesServiceAccountPattern.MatchString(pattern) {
			return ErrKubernetesServiceAccountPatternInvalid
		}
	}
	for _, selector := range rule.SecretSelectors {
		if !githubSecretSelectorPattern.MatchString(selector) {
			return ErrKubernetesSecretSelectorInvalid
		}
	}
	return nil
}

func normalizeRequestedKubernetesSecrets(secrets map[string]string) (map[string]string, error) {
	if len(secrets) == 0 {
		return map[string]string{}, ErrKubernetesSecretSelectorRequired
	}
	normalized := make(map[string]string, len(secrets))
	for key, selector := range secrets {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" {
			return map[string]string{}, ErrKubernetesSecretKeyRequired
		}
		if !kubernetesSecretKeyPattern.MatchString(trimmedKey) {
			return map[string]string{}, ErrKubernetesSecretKeyInvalid
		}

		normalizedSelector := strings.ToLower(strings.TrimSpace(selector))
		if normalizedSelector == "" {
			return map[string]string{}, ErrKubernetesSecretSelectorRequired
		}
		if strings.Contains(normalizedSelector, "*") || !githubSecretSelectorPattern.MatchString(normalizedSelector) {
			return map[string]string{}, ErrKubernetesRequestedSecretSelectorInvalid
		}
		normalized[trimmedKey] = normalizedSelector
	}
	return normalized, nil
}

func kubernetesSecretRuleMatchesClaims(rule model.KubernetesSecretRule, cluster model.KubernetesCluster, claims kubernetes.Claims) bool {
	if !stringSliceContains(rule.ClusterIDs, cluster.ID) {
		return false
	}
	if !anyWildcardMatches(rule.NamespacePatterns, strings.ToLower(claims.Namespace)) {
		return false
	}
	return anyWildcardMatches(rule.ServiceAccountPatterns, strings.ToLower(claims.ServiceAccountName))
}

func kubernetesSelectorAllowed(rules []model.KubernetesSecretRule, selector string) bool {
	for _, rule := range rules {
		if anyWildcardMatches(rule.SecretSelectors, selector) {
			return true
		}
	}
	return false
}

func normalizeLowerStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		lower := strings.ToLower(strings.TrimSpace(value))
		if lower == "" || seen[lower] {
			continue
		}
		seen[lower] = true
		normalized = append(normalized, lower)
	}
	return normalized
}
