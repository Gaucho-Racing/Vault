package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"gorm.io/gorm"
)

var ErrSecretKeyRequired = errors.New("secret key is required")

func GetSecretsForAccount(accountID string) ([]model.Secret, error) {
	secrets := []model.Secret{}
	if err := database.DB.
		Where("account_id = ? AND archived_at IS NULL", accountID).
		Order("key ASC").
		Find(&secrets).Error; err != nil {
		return []model.Secret{}, err
	}
	return secrets, nil
}

func GetSecretByID(id string) (model.Secret, error) {
	var secret model.Secret
	if err := database.DB.Where("id = ? AND archived_at IS NULL", id).First(&secret).Error; err != nil {
		return model.Secret{}, err
	}
	return secret, nil
}

func GetSecretForAccount(accountID string, secretID string) (model.Secret, error) {
	var secret model.Secret
	if err := database.DB.Where("id = ? AND account_id = ? AND archived_at IS NULL", secretID, accountID).First(&secret).Error; err != nil {
		return model.Secret{}, err
	}
	return secret, nil
}

func CreateSecret(secret model.Secret) (model.Secret, error) {
	if _, err := GetAccountByID(secret.AccountID); err != nil {
		return model.Secret{}, err
	}
	if secret.ID == "" {
		secret.ID = ulid.Make().Prefixed("sec")
	}
	if err := normalizeSecret(&secret); err != nil {
		return model.Secret{}, err
	}
	if err := database.DB.Create(&secret).Error; err != nil {
		return model.Secret{}, err
	}
	return secret, nil
}

func UpdateSecret(secret model.Secret) (model.Secret, error) {
	if _, err := GetAccountByID(secret.AccountID); err != nil {
		return model.Secret{}, err
	}
	if err := normalizeSecret(&secret); err != nil {
		return model.Secret{}, err
	}
	if err := database.DB.Save(&secret).Error; err != nil {
		return model.Secret{}, err
	}
	return secret, nil
}

func ArchiveSecret(accountID string, secretID string, entityID string) error {
	now := time.Now()
	result := database.DB.
		Model(&model.Secret{}).
		Where("id = ? AND account_id = ?", secretID, accountID).
		Updates(map[string]interface{}{
			"archived_at":          &now,
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

func RevealSecret(secret model.Secret) (string, error) {
	if secret.Sensitive {
		return decryptSecretValue(secret)
	}
	return secret.PlainValue, nil
}

func normalizeSecret(secret *model.Secret) error {
	secret.Key = strings.TrimSpace(secret.Key)
	secret.Label = strings.TrimSpace(secret.Label)
	secret.Type = strings.TrimSpace(secret.Type)
	if secret.Key == "" {
		return ErrSecretKeyRequired
	}
	if secret.Sensitive {
		if secret.PlainValue != "" {
			return encryptSecretValue(secret)
		}
		if !hasEncryptedSecretValue(*secret) {
			return ErrSensitiveSecretValueRequired
		}
		secret.PlainValue = ""
		return nil
	}
	clearEncryptedSecretValue(secret)
	return nil
}
