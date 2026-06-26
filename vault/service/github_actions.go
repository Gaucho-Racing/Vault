package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/githubactions"
	"gorm.io/gorm"
)

var ErrGitHubActionsRuleNameRequired = errors.New("github actions rule name is required")
var ErrGitHubActionsRuleNameInvalid = errors.New("github actions rule name must contain only lowercase letters, numbers, hyphens, and underscores")
var ErrGitHubActionsRepositoryPatternRequired = errors.New("github actions repository pattern is required")
var ErrGitHubActionsRepositoryPatternInvalid = errors.New("github actions repository pattern must use owner/repo format and may include wildcards")
var ErrGitHubActionsRefPatternRequired = errors.New("github actions ref pattern is required")
var ErrGitHubActionsRefPatternInvalid = errors.New("github actions ref pattern must start with refs/ and may include wildcards")
var ErrGitHubActionsSecretSelectorRequired = errors.New("github actions secret selector is required")
var ErrGitHubActionsSecretSelectorInvalid = errors.New("github actions secret selector must use application.secret format and may include wildcards")
var ErrGitHubActionsSecretSelectorNotAllowed = errors.New("github actions secret selector is not allowed")
var ErrGitHubActionsEnvNameCollision = errors.New("github actions environment variable name collision")

type GitHubActionsExportSecret struct {
	Selector string
	Secret   model.AppSecret
}

var githubRepositoryPattern = regexp.MustCompile(`^[a-z0-9*][a-z0-9_.*-]*/[a-z0-9*][a-z0-9_.*-]*$`)
var githubSecretSelectorPattern = regexp.MustCompile(`^[a-z0-9*][a-z0-9_*-]*\.[a-z0-9*][a-z0-9_*-]*$`)
var whitespacePattern = regexp.MustCompile(`\s`)

func GetAllGitHubActionsRules() ([]model.GitHubActionsRule, error) {
	rules := []model.GitHubActionsRule{}
	if err := database.DB.Order("name ASC").Find(&rules).Error; err != nil {
		return []model.GitHubActionsRule{}, err
	}
	return rules, nil
}

func GetGitHubActionsRuleByID(id string) (model.GitHubActionsRule, error) {
	var rule model.GitHubActionsRule
	if err := database.DB.Where("id = ?", id).First(&rule).Error; err != nil {
		return model.GitHubActionsRule{}, err
	}
	return rule, nil
}

func CreateGitHubActionsRule(rule model.GitHubActionsRule) (model.GitHubActionsRule, error) {
	if rule.ID == "" {
		rule.ID = ulid.Make().Prefixed("gharule")
	}
	if err := normalizeGitHubActionsRule(&rule); err != nil {
		return model.GitHubActionsRule{}, err
	}
	if err := database.DB.Create(&rule).Error; err != nil {
		return model.GitHubActionsRule{}, err
	}
	return rule, nil
}

func UpdateGitHubActionsRule(rule model.GitHubActionsRule) (model.GitHubActionsRule, error) {
	if err := normalizeGitHubActionsRule(&rule); err != nil {
		return model.GitHubActionsRule{}, err
	}
	if err := database.DB.Save(&rule).Error; err != nil {
		return model.GitHubActionsRule{}, err
	}
	return rule, nil
}

func DeleteGitHubActionsRule(id string) error {
	result := database.DB.Where("id = ?", id).Delete(&model.GitHubActionsRule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func BuildGitHubActionsEnvFile(selectors []string, claims githubactions.Claims) (string, error) {
	exportSecrets, err := ResolveGitHubActionsExportSecrets(selectors, claims)
	if err != nil {
		return "", err
	}

	envNamesBySelector := map[string]string{}
	var builder strings.Builder
	for _, exportSecret := range exportSecrets {
		name := githubActionsEnvName(exportSecret.Secret.Key)
		if existingSelector, exists := envNamesBySelector[name]; exists {
			return "", fmt.Errorf("%w: %s and %s both map to %s", ErrGitHubActionsEnvNameCollision, existingSelector, exportSecret.Selector, name)
		}
		envNamesBySelector[name] = exportSecret.Selector

		value, err := RevealAppSecret(exportSecret.Secret)
		if err != nil {
			return "", fmt.Errorf("reveal app secret %s: %w", exportSecret.Selector, err)
		}
		builder.WriteString(formatGitHubActionsEnvAssignment(name, value))
	}
	return builder.String(), nil
}

func ResolveGitHubActionsExportSecrets(selectors []string, claims githubactions.Claims) ([]GitHubActionsExportSecret, error) {
	normalizedSelectors, err := normalizeRequestedGitHubActionsSecretSelectors(selectors)
	if err != nil {
		return []GitHubActionsExportSecret{}, err
	}

	rules := []model.GitHubActionsRule{}
	if err := database.DB.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return []GitHubActionsExportSecret{}, err
	}

	matchingRules := make([]model.GitHubActionsRule, 0, len(rules))
	for _, rule := range rules {
		if githubActionsRuleMatchesClaims(rule, claims) {
			matchingRules = append(matchingRules, rule)
		}
	}

	exportSecrets := make([]GitHubActionsExportSecret, 0, len(normalizedSelectors))
	for _, selector := range normalizedSelectors {
		if !githubActionsSelectorAllowed(matchingRules, selector) {
			return []GitHubActionsExportSecret{}, fmt.Errorf("%w: %s", ErrGitHubActionsSecretSelectorNotAllowed, selector)
		}
		secret, err := getAppSecretBySelector(selector)
		if err != nil {
			return []GitHubActionsExportSecret{}, err
		}
		exportSecrets = append(exportSecrets, GitHubActionsExportSecret{
			Selector: selector,
			Secret:   secret,
		})
	}
	return exportSecrets, nil
}

func normalizeGitHubActionsRule(rule *model.GitHubActionsRule) error {
	rule.Name = strings.ToLower(strings.TrimSpace(rule.Name))
	rule.RepositoryPatterns = normalizeGitHubRepositoryPatterns(rule.RepositoryPatterns)
	rule.RefPatterns = normalizeStringSlice(rule.RefPatterns)
	rule.SecretSelectors = normalizeGitHubActionsSecretSelectors(rule.SecretSelectors)
	if rule.Name == "" {
		return ErrGitHubActionsRuleNameRequired
	}
	if !appSecretIdentifierPattern.MatchString(rule.Name) {
		return ErrGitHubActionsRuleNameInvalid
	}
	if len(rule.RepositoryPatterns) == 0 {
		return ErrGitHubActionsRepositoryPatternRequired
	}
	if len(rule.RefPatterns) == 0 {
		return ErrGitHubActionsRefPatternRequired
	}
	if len(rule.SecretSelectors) == 0 {
		return ErrGitHubActionsSecretSelectorRequired
	}
	for _, pattern := range rule.RepositoryPatterns {
		if !githubRepositoryPattern.MatchString(pattern) {
			return ErrGitHubActionsRepositoryPatternInvalid
		}
	}
	for _, pattern := range rule.RefPatterns {
		if !strings.HasPrefix(pattern, "refs/") || whitespacePattern.MatchString(pattern) {
			return ErrGitHubActionsRefPatternInvalid
		}
	}
	for _, selector := range rule.SecretSelectors {
		if !githubSecretSelectorPattern.MatchString(selector) {
			return ErrGitHubActionsSecretSelectorInvalid
		}
	}
	return nil
}

func normalizeRequestedGitHubActionsSecretSelectors(selectors []string) ([]string, error) {
	normalized := normalizeGitHubActionsSecretSelectors(selectors)
	if len(normalized) == 0 {
		return []string{}, ErrGitHubActionsSecretSelectorRequired
	}
	for _, selector := range normalized {
		if strings.Contains(selector, "*") || !githubSecretSelectorPattern.MatchString(selector) {
			return []string{}, ErrGitHubActionsSecretSelectorInvalid
		}
	}
	return normalized, nil
}

func normalizeGitHubRepositoryPatterns(patterns []string) []string {
	normalized := make([]string, 0, len(patterns))
	seen := map[string]bool{}
	for _, pattern := range patterns {
		value := strings.ToLower(strings.TrimSpace(pattern))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeGitHubActionsSecretSelectors(selectors []string) []string {
	normalized := make([]string, 0, len(selectors))
	seen := map[string]bool{}
	for _, selector := range selectors {
		value := strings.ToLower(strings.TrimSpace(selector))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	return normalized
}

func githubActionsRuleMatchesClaims(rule model.GitHubActionsRule, claims githubactions.Claims) bool {
	if !anyWildcardMatches(rule.RepositoryPatterns, strings.ToLower(claims.Repository)) {
		return false
	}
	return anyWildcardMatches(rule.RefPatterns, claims.Ref)
}

func githubActionsSelectorAllowed(rules []model.GitHubActionsRule, selector string) bool {
	for _, rule := range rules {
		if anyWildcardMatches(rule.SecretSelectors, selector) {
			return true
		}
	}
	return false
}

func getAppSecretBySelector(selector string) (model.AppSecret, error) {
	applicationName, secretKey, ok := strings.Cut(selector, ".")
	if !ok || applicationName == "" || secretKey == "" {
		return model.AppSecret{}, ErrGitHubActionsSecretSelectorInvalid
	}
	application, err := GetApplicationByName(applicationName)
	if err != nil {
		return model.AppSecret{}, err
	}
	return GetAppSecretForApplicationByKey(application.ID, secretKey)
}

func anyWildcardMatches(patterns []string, value string) bool {
	for _, pattern := range patterns {
		if wildcardMatches(pattern, value) {
			return true
		}
	}
	return false
}

func wildcardMatches(pattern string, value string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == value
	}

	position := 0
	if parts[0] != "" {
		if !strings.HasPrefix(value, parts[0]) {
			return false
		}
		position = len(parts[0])
	}

	for _, part := range parts[1 : len(parts)-1] {
		index := strings.Index(value[position:], part)
		if index < 0 {
			return false
		}
		position += index + len(part)
	}

	last := parts[len(parts)-1]
	return last == "" || strings.HasSuffix(value[position:], last)
}

func githubActionsEnvName(key string) string {
	var builder strings.Builder
	for _, value := range strings.ToUpper(key) {
		if (value >= 'A' && value <= 'Z') || (value >= '0' && value <= '9') || value == '_' {
			builder.WriteRune(value)
		} else {
			builder.WriteByte('_')
		}
	}
	name := builder.String()
	if name == "" {
		return "SECRET"
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "_" + name
	}
	return name
}

func formatGitHubActionsEnvAssignment(name string, value string) string {
	delimiter := githubActionsDelimiter(name, value)
	return name + "<<" + delimiter + "\n" + value + "\n" + delimiter + "\n"
}

func githubActionsDelimiter(name string, value string) string {
	hash := sha256.Sum256([]byte(name + "\x00" + value))
	delimiter := "VAULT_" + strings.ToUpper(hex.EncodeToString(hash[:8]))
	for suffix := 1; strings.Contains(value, delimiter); suffix++ {
		delimiter = fmt.Sprintf("VAULT_%s_%d", strings.ToUpper(hex.EncodeToString(hash[:8])), suffix)
	}
	return delimiter
}
