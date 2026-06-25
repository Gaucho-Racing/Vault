package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
)

var ErrAccountNameRequired = errors.New("account name is required")

type AccountWithSecrets struct {
	model.Account
	Secrets []model.Secret `json:"secrets"`
}

func GetAllAccounts() ([]model.Account, error) {
	accounts := []model.Account{}
	if err := database.DB.
		Where("archived_at IS NULL").
		Order("name ASC").
		Find(&accounts).Error; err != nil {
		return []model.Account{}, err
	}
	return accounts, nil
}

func GetAccountByID(id string) (model.Account, error) {
	var account model.Account
	if err := database.DB.Where("id = ? AND archived_at IS NULL", id).First(&account).Error; err != nil {
		return model.Account{}, err
	}
	return account, nil
}

func GetAccountWithSecrets(id string) (AccountWithSecrets, error) {
	account, err := GetAccountByID(id)
	if err != nil {
		return AccountWithSecrets{}, err
	}
	secrets, err := GetSecretsForAccount(id)
	if err != nil {
		return AccountWithSecrets{}, err
	}
	return AccountWithSecrets{Account: account, Secrets: secrets}, nil
}

func CreateAccount(account model.Account) (model.Account, error) {
	if account.ID == "" {
		account.ID = ulid.Make().Prefixed("acct")
	}
	normalizeAccount(&account)
	if account.Name == "" {
		return model.Account{}, ErrAccountNameRequired
	}
	if err := database.DB.Create(&account).Error; err != nil {
		return model.Account{}, err
	}
	return account, nil
}

func UpdateAccount(account model.Account) (model.Account, error) {
	normalizeAccount(&account)
	if account.Name == "" {
		return model.Account{}, ErrAccountNameRequired
	}
	if err := database.DB.Save(&account).Error; err != nil {
		return model.Account{}, err
	}
	return account, nil
}

func ArchiveAccount(id string, entityID string) error {
	now := time.Now()
	return database.DB.
		Model(&model.Account{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"archived_at":          &now,
			"updated_by_entity_id": entityID,
		}).Error
}

func normalizeAccount(account *model.Account) {
	account.Name = strings.TrimSpace(account.Name)
	account.Description = strings.TrimSpace(account.Description)
	account.URL = strings.TrimSpace(account.URL)
	account.AccessGroupNames = normalizeStringSlice(account.AccessGroupNames)
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
