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

type AccountWithSecretCount struct {
	model.Account
	SecretCount int64 `json:"secret_count"`
}

type accountSecretCount struct {
	AccountID   string
	SecretCount int64
}

func GetAllAccounts() ([]AccountWithSecretCount, error) {
	accounts := []model.Account{}
	if err := database.DB.
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&accounts).Error; err != nil {
		return []AccountWithSecretCount{}, err
	}

	accountIDs := make([]string, 0, len(accounts))
	for _, account := range accounts {
		accountIDs = append(accountIDs, account.ID)
	}

	countsByAccountID, err := getSecretCountsByAccountID(accountIDs)
	if err != nil {
		return []AccountWithSecretCount{}, err
	}

	result := make([]AccountWithSecretCount, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, AccountWithSecretCount{
			Account:     account,
			SecretCount: countsByAccountID[account.ID],
		})
	}
	return result, nil
}

func GetAccountByID(id string) (model.Account, error) {
	var account model.Account
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(&account).Error; err != nil {
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

func getSecretCountsByAccountID(accountIDs []string) (map[string]int64, error) {
	countsByAccountID := make(map[string]int64, len(accountIDs))
	if len(accountIDs) == 0 {
		return countsByAccountID, nil
	}

	counts := []accountSecretCount{}
	if err := database.DB.
		Model(&model.Secret{}).
		Select("account_id, count(*) as secret_count").
		Where("account_id IN ? AND deleted_at IS NULL", accountIDs).
		Group("account_id").
		Scan(&counts).Error; err != nil {
		return map[string]int64{}, err
	}

	for _, count := range counts {
		countsByAccountID[count.AccountID] = count.SecretCount
	}
	return countsByAccountID, nil
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

func DeleteAccount(id string, entityID string) error {
	now := time.Now()
	return database.DB.
		Model(&model.Account{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at":           &now,
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
