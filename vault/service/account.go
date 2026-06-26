package service

import (
	"errors"
	"strings"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"gorm.io/gorm"
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
	if err := database.DB.Where("id = ?", id).First(&account).Error; err != nil {
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
		Where("account_id IN ?", accountIDs).
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
	return createAccount(database.DB, account)
}

func CreateAccountWithAudit(account model.Account, auditLog model.AuditLog) (model.Account, error) {
	created := model.Account{}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		created, err = createAccount(tx, account)
		if err != nil {
			return err
		}
		auditLog.AccountID = created.ID
		auditLog.AccountName = created.Name
		return recordAuditLog(tx, auditLog)
	})
	return created, err
}

func createAccount(db *gorm.DB, account model.Account) (model.Account, error) {
	if account.ID == "" {
		account.ID = ulid.Make().Prefixed("acct")
	}
	normalizeAccount(&account)
	if account.Name == "" {
		return model.Account{}, ErrAccountNameRequired
	}
	if err := db.Create(&account).Error; err != nil {
		return model.Account{}, err
	}
	return account, nil
}

func UpdateAccount(account model.Account) (model.Account, error) {
	return updateAccount(database.DB, account)
}

func UpdateAccountWithAudit(account model.Account, auditLog model.AuditLog) (model.Account, error) {
	updated := model.Account{}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		updated, err = updateAccount(tx, account)
		if err != nil {
			return err
		}
		auditLog.AccountID = updated.ID
		auditLog.AccountName = updated.Name
		return recordAuditLog(tx, auditLog)
	})
	return updated, err
}

func updateAccount(db *gorm.DB, account model.Account) (model.Account, error) {
	normalizeAccount(&account)
	if account.Name == "" {
		return model.Account{}, ErrAccountNameRequired
	}
	if err := db.Save(&account).Error; err != nil {
		return model.Account{}, err
	}
	return account, nil
}

func DeleteAccount(id string) error {
	return deleteAccount(database.DB, id)
}

func DeleteAccountWithAudit(account model.Account, auditLog model.AuditLog) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := deleteAccount(tx, account.ID); err != nil {
			return err
		}
		auditLog.AccountID = account.ID
		auditLog.AccountName = account.Name
		return recordAuditLog(tx, auditLog)
	})
}

func deleteAccount(db *gorm.DB, id string) error {
	if err := db.Where("account_id = ?", id).Delete(&model.Secret{}).Error; err != nil {
		return err
	}
	result := db.Where("id = ?", id).Delete(&model.Account{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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
