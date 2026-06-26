package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"gorm.io/gorm"
)

var ErrApplicationNameRequired = errors.New("application name is required")
var ErrApplicationNameInvalid = errors.New("application name must contain only lowercase letters, numbers, hyphens, and underscores")
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

func GetAllApplications() ([]ApplicationWithSecretCount, error) {
	applications := []model.Application{}
	if err := database.DB.
		Where("deleted_at IS NULL").
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
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(&application).Error; err != nil {
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

func DeleteApplication(application model.Application, entityID string) error {
	now := time.Now()
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&model.AppSecret{}).
			Where("application_id = ? AND deleted_at IS NULL", application.ID).
			Updates(map[string]interface{}{
				"deleted_at":           &now,
				"updated_by_entity_id": entityID,
			}).Error; err != nil {
			return err
		}

		result := tx.
			Model(&model.Application{}).
			Where("id = ? AND deleted_at IS NULL", application.ID).
			Updates(map[string]interface{}{
				"deleted_at":           &now,
				"updated_by_entity_id": entityID,
			})
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
		Where("application_id = ? AND deleted_at IS NULL", applicationID).
		Order("key ASC").
		Find(&secrets).Error; err != nil {
		return []model.AppSecret{}, err
	}
	return secrets, nil
}

func GetAppSecretForApplication(applicationID string, secretID string) (model.AppSecret, error) {
	var secret model.AppSecret
	if err := database.DB.
		Where("id = ? AND application_id = ? AND deleted_at IS NULL", secretID, applicationID).
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

func DeleteAppSecret(applicationID string, secretID string, entityID string) error {
	now := time.Now()
	result := database.DB.
		Model(&model.AppSecret{}).
		Where("id = ? AND application_id = ? AND deleted_at IS NULL", secretID, applicationID).
		Updates(map[string]interface{}{
			"deleted_at":           &now,
			"updated_by_entity_id": entityID,
		})
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

func getAppSecretCountsByApplicationID(applicationIDs []string) (map[string]int64, error) {
	countsByApplicationID := make(map[string]int64, len(applicationIDs))
	if len(applicationIDs) == 0 {
		return countsByApplicationID, nil
	}

	counts := []applicationSecretCount{}
	if err := database.DB.
		Model(&model.AppSecret{}).
		Select("application_id, count(*) as secret_count").
		Where("application_id IN ? AND deleted_at IS NULL", applicationIDs).
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
	if application.Name == "" {
		return ErrApplicationNameRequired
	}
	if !appSecretIdentifierPattern.MatchString(application.Name) {
		return ErrApplicationNameInvalid
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
