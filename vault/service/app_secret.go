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

var ErrApplicationNameRequired = errors.New("application name is required")
var ErrApplicationNameInvalid = errors.New("application name must contain only lowercase letters, numbers, hyphens, and underscores")
var ErrGitHubActionsRepositoryRequired = errors.New("github actions repository is required when refs are configured")
var ErrGitHubActionsRefRequired = errors.New("github actions ref is required when repositories are configured")
var ErrGitHubActionsRepositoryInvalid = errors.New("github actions repository must use owner/repo format")
var ErrGitHubActionsRefInvalid = errors.New("github actions ref must start with refs/ and cannot contain whitespace")
var ErrGitHubActionsRepositoryNotAllowed = errors.New("github actions repository is not allowed for this application")
var ErrGitHubActionsRefNotAllowed = errors.New("github actions ref is not allowed for this application")
var ErrGitHubActionsEnvNameCollision = errors.New("github actions environment variable name collision")
var ErrAppSecretKeyRequired = errors.New("app secret key is required")
var ErrAppSecretKeyInvalid = errors.New("app secret key must contain only lowercase letters, numbers, hyphens, and underscores")
var ErrAppSecretValueRequired = errors.New("app secret value is required")

type ApplicationWithSecrets struct {
	model.Application
	Secrets []model.AppSecret `json:"secrets"`
}

type ApplicationWithSecretCount struct {
	model.Application
	SecretCount int64 `json:"secret_count"`
}

type applicationSecretCount struct {
	ApplicationID string
	SecretCount   int64
}

var appSecretIdentifierPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
var githubRepositoryPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*/[a-z0-9][a-z0-9_.-]*$`)
var whitespacePattern = regexp.MustCompile(`\s`)

func GetAllApplications() ([]ApplicationWithSecretCount, error) {
	applications := []model.Application{}
	if err := database.DB.
		Order("name ASC").
		Find(&applications).Error; err != nil {
		return []ApplicationWithSecretCount{}, err
	}

	applicationIDs := make([]string, 0, len(applications))
	for _, application := range applications {
		applicationIDs = append(applicationIDs, application.ID)
	}

	countsByApplicationID, err := getAppSecretCountsByApplicationID(applicationIDs)
	if err != nil {
		return []ApplicationWithSecretCount{}, err
	}

	result := make([]ApplicationWithSecretCount, 0, len(applications))
	for _, application := range applications {
		result = append(result, ApplicationWithSecretCount{
			Application: application,
			SecretCount: countsByApplicationID[application.ID],
		})
	}
	return result, nil
}

func GetApplicationByID(id string) (model.Application, error) {
	var application model.Application
	if err := database.DB.Where("id = ?", id).First(&application).Error; err != nil {
		return model.Application{}, err
	}
	return application, nil
}

func GetApplicationByName(name string) (model.Application, error) {
	var application model.Application
	if err := database.DB.Where("name = ?", strings.ToLower(strings.TrimSpace(name))).First(&application).Error; err != nil {
		return model.Application{}, err
	}
	return application, nil
}

func GetApplicationWithSecrets(id string) (ApplicationWithSecrets, error) {
	application, err := GetApplicationByID(id)
	if err != nil {
		return ApplicationWithSecrets{}, err
	}
	secrets, err := GetAppSecretsForApplication(id)
	if err != nil {
		return ApplicationWithSecrets{}, err
	}
	return ApplicationWithSecrets{Application: application, Secrets: secrets}, nil
}

func CreateApplication(application model.Application) (model.Application, error) {
	if application.ID == "" {
		application.ID = ulid.Make().Prefixed("app")
	}
	if err := normalizeApplication(&application); err != nil {
		return model.Application{}, err
	}
	if err := database.DB.Create(&application).Error; err != nil {
		return model.Application{}, err
	}
	return application, nil
}

func UpdateApplication(application model.Application) (model.Application, error) {
	if err := normalizeApplication(&application); err != nil {
		return model.Application{}, err
	}
	if err := database.DB.Save(&application).Error; err != nil {
		return model.Application{}, err
	}
	return application, nil
}

func DeleteApplication(application model.Application) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("application_id = ?", application.ID).Delete(&model.AppSecret{}).Error; err != nil {
			return err
		}
		result := tx.Where("id = ?", application.ID).Delete(&model.Application{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func GetAppSecretsForApplication(applicationID string) ([]model.AppSecret, error) {
	secrets := []model.AppSecret{}
	if err := database.DB.
		Where("application_id = ?", applicationID).
		Order("key ASC").
		Find(&secrets).Error; err != nil {
		return []model.AppSecret{}, err
	}
	return secrets, nil
}

func GetAppSecretForApplication(applicationID string, secretID string) (model.AppSecret, error) {
	var secret model.AppSecret
	if err := database.DB.
		Where("id = ? AND application_id = ?", secretID, applicationID).
		First(&secret).Error; err != nil {
		return model.AppSecret{}, err
	}
	return secret, nil
}

func CreateAppSecret(secret model.AppSecret) (model.AppSecret, error) {
	if _, err := GetApplicationByID(secret.ApplicationID); err != nil {
		return model.AppSecret{}, err
	}
	if secret.ID == "" {
		secret.ID = ulid.Make().Prefixed("appsec")
	}
	if err := normalizeAppSecret(&secret, true); err != nil {
		return model.AppSecret{}, err
	}
	if err := database.DB.Create(&secret).Error; err != nil {
		return model.AppSecret{}, err
	}
	return secret, nil
}

func UpdateAppSecret(secret model.AppSecret) (model.AppSecret, error) {
	if _, err := GetApplicationByID(secret.ApplicationID); err != nil {
		return model.AppSecret{}, err
	}
	if err := normalizeAppSecret(&secret, false); err != nil {
		return model.AppSecret{}, err
	}
	if err := database.DB.Save(&secret).Error; err != nil {
		return model.AppSecret{}, err
	}
	return secret, nil
}

func DeleteAppSecret(applicationID string, secretID string) error {
	result := database.DB.
		Where("id = ? AND application_id = ?", secretID, applicationID).
		Delete(&model.AppSecret{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func RevealAppSecret(secret model.AppSecret) (string, error) {
	return decryptPlainValue(encryptedAppSecretValue(secret), appSecretAssociatedData(secret))
}

func BuildApplicationEnvFile(applicationID string) (string, error) {
	secrets, err := GetAppSecretsForApplication(applicationID)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for _, secret := range secrets {
		value, err := RevealAppSecret(secret)
		if err != nil {
			return "", fmt.Errorf("reveal app secret %s: %w", secret.Key, err)
		}
		builder.WriteString(secret.Key)
		builder.WriteString("=")
		builder.WriteString(formatEnvValue(value))
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func BuildApplicationGitHubActionsEnvFile(applicationName string, claims githubactions.Claims) (string, error) {
	application, err := GetApplicationByName(applicationName)
	if err != nil {
		return "", err
	}
	if err := authorizeGitHubActionsApplicationAccess(application, claims); err != nil {
		return "", err
	}
	return BuildGitHubActionsEnvFile(application.ID)
}

func BuildGitHubActionsEnvFile(applicationID string) (string, error) {
	secrets, err := GetAppSecretsForApplication(applicationID)
	if err != nil {
		return "", err
	}

	secretKeysByEnvName := map[string]string{}
	var builder strings.Builder
	for _, secret := range secrets {
		name := githubActionsEnvName(secret.Key)
		if existingKey, exists := secretKeysByEnvName[name]; exists {
			return "", fmt.Errorf("%w: %s and %s both map to %s", ErrGitHubActionsEnvNameCollision, existingKey, secret.Key, name)
		}
		secretKeysByEnvName[name] = secret.Key

		value, err := RevealAppSecret(secret)
		if err != nil {
			return "", fmt.Errorf("reveal app secret %s: %w", secret.Key, err)
		}
		builder.WriteString(formatGitHubActionsEnvAssignment(name, value))
	}
	return builder.String(), nil
}

func getAppSecretCountsByApplicationID(applicationIDs []string) (map[string]int64, error) {
	countsByApplicationID := make(map[string]int64, len(applicationIDs))
	if len(applicationIDs) == 0 {
		return countsByApplicationID, nil
	}

	counts := []applicationSecretCount{}
	if err := database.DB.
		Model(&model.AppSecret{}).
		Select("application_id, count(*) as secret_count").
		Where("application_id IN ?", applicationIDs).
		Group("application_id").
		Scan(&counts).Error; err != nil {
		return map[string]int64{}, err
	}

	for _, count := range counts {
		countsByApplicationID[count.ApplicationID] = count.SecretCount
	}
	return countsByApplicationID, nil
}

func normalizeApplication(application *model.Application) error {
	application.Name = strings.ToLower(strings.TrimSpace(application.Name))
	application.AccessGroupNames = normalizeStringSlice(application.AccessGroupNames)
	application.GitHubActionsRepositories = normalizeGitHubRepositories(application.GitHubActionsRepositories)
	application.GitHubActionsRefs = normalizeStringSlice(application.GitHubActionsRefs)
	if application.Name == "" {
		return ErrApplicationNameRequired
	}
	if !appSecretIdentifierPattern.MatchString(application.Name) {
		return ErrApplicationNameInvalid
	}
	if len(application.GitHubActionsRefs) > 0 && len(application.GitHubActionsRepositories) == 0 {
		return ErrGitHubActionsRepositoryRequired
	}
	if len(application.GitHubActionsRepositories) > 0 && len(application.GitHubActionsRefs) == 0 {
		return ErrGitHubActionsRefRequired
	}
	for _, repository := range application.GitHubActionsRepositories {
		if !githubRepositoryPattern.MatchString(repository) {
			return ErrGitHubActionsRepositoryInvalid
		}
	}
	for _, ref := range application.GitHubActionsRefs {
		if !strings.HasPrefix(ref, "refs/") || whitespacePattern.MatchString(ref) {
			return ErrGitHubActionsRefInvalid
		}
	}
	return nil
}

func normalizeAppSecret(secret *model.AppSecret, requireValue bool) error {
	secret.Key = strings.ToLower(strings.TrimSpace(secret.Key))
	if secret.Key == "" {
		return ErrAppSecretKeyRequired
	}
	if !appSecretIdentifierPattern.MatchString(secret.Key) {
		return ErrAppSecretKeyInvalid
	}
	if secret.PlainValue != "" {
		encryptedValue, err := encryptPlainValue(secret.PlainValue, appSecretAssociatedData(*secret))
		if err != nil {
			return err
		}
		secret.Ciphertext = encryptedValue.Ciphertext
		secret.Nonce = encryptedValue.Nonce
		secret.EncryptedDataKey = encryptedValue.EncryptedDataKey
		secret.KeyID = encryptedValue.KeyID
		secret.Algorithm = encryptedValue.Algorithm
		secret.PlainValue = ""
		return nil
	}
	if requireValue || !hasEncryptedPlainValue(encryptedAppSecretValue(*secret)) {
		return ErrAppSecretValueRequired
	}
	return nil
}

func encryptedAppSecretValue(secret model.AppSecret) encryptedSecretValue {
	return encryptedSecretValue{
		Ciphertext:       secret.Ciphertext,
		Nonce:            secret.Nonce,
		EncryptedDataKey: secret.EncryptedDataKey,
		KeyID:            secret.KeyID,
		Algorithm:        secret.Algorithm,
	}
}

func appSecretAssociatedData(secret model.AppSecret) []byte {
	return []byte(secret.ApplicationID + "\x00" + secret.ID)
}

func formatEnvValue(value string) string {
	if value == "" {
		return `""`
	}
	if strings.TrimSpace(value) == value && !strings.ContainsAny(value, " \t\r\n#\"'\\=") {
		return value
	}
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		`"`, `\"`,
		"\r", `\r`,
		"\n", `\n`,
	)
	return `"` + replacer.Replace(value) + `"`
}

func authorizeGitHubActionsApplicationAccess(application model.Application, claims githubactions.Claims) error {
	if !stringSliceContainsFold(application.GitHubActionsRepositories, claims.Repository) {
		return ErrGitHubActionsRepositoryNotAllowed
	}
	for _, refPattern := range application.GitHubActionsRefs {
		if wildcardMatches(refPattern, claims.Ref) {
			return nil
		}
	}
	return ErrGitHubActionsRefNotAllowed
}

func normalizeGitHubRepositories(repositories []string) []string {
	normalized := make([]string, 0, len(repositories))
	seen := map[string]bool{}
	for _, repository := range repositories {
		value := strings.ToLower(strings.TrimSpace(repository))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	return normalized
}

func stringSliceContainsFold(values []string, candidate string) bool {
	for _, value := range values {
		if strings.EqualFold(value, candidate) {
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
