package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"gorm.io/gorm"
)

const vaultMasterKeyAlgorithm = "AES-256-GCM"

func InitializeVaultKeys() {
	key, created, err := loadOrCreateVaultMasterKey()
	if err != nil {
		logger.SugarLogger.Fatalf("failed to initialize vault key: %v", err)
		return
	}
	config.VaultMasterKey = key.KeyMaterial
	config.VaultMasterKeyID = key.ID
	if created {
		logger.SugarLogger.Infof("Persisted new vault key %s", key.ID)
		return
	}
	logger.SugarLogger.Infof("Loaded vault key %s from db", key.ID)
}

func loadOrCreateVaultMasterKey() (model.VaultKey, bool, error) {
	key, err := getActiveVaultKey()
	if err == nil {
		return key, false, validateVaultKey(key)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.VaultKey{}, false, err
	}

	key, err = buildVaultKey()
	if err != nil {
		return model.VaultKey{}, false, err
	}
	if err := database.DB.Create(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			key, err = getActiveVaultKey()
			if err != nil {
				return model.VaultKey{}, false, err
			}
			return key, false, validateVaultKey(key)
		}
		return model.VaultKey{}, false, err
	}
	return key, true, nil
}

func getActiveVaultKey() (model.VaultKey, error) {
	var key model.VaultKey
	result := database.DB.Where("active = ?", true).Limit(1).Find(&key)
	if result.Error != nil {
		return model.VaultKey{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.VaultKey{}, gorm.ErrRecordNotFound
	}
	return key, nil
}

func buildVaultKey() (model.VaultKey, error) {
	if strings.TrimSpace(config.VaultMasterKey) != "" {
		return buildConfiguredVaultKey()
	}

	keyMaterial, err := randomBytes(secretDataKeySize)
	if err != nil {
		return model.VaultKey{}, err
	}
	defer zeroBytes(keyMaterial)

	return model.VaultKey{
		ID:          ulid.Make().Prefixed("vkey"),
		Algorithm:   vaultMasterKeyAlgorithm,
		KeyMaterial: base64.StdEncoding.EncodeToString(keyMaterial),
		Active:      true,
	}, nil
}

func buildConfiguredVaultKey() (model.VaultKey, error) {
	keyMaterial, err := decodeVaultMasterKey(config.VaultMasterKey)
	if err != nil {
		return model.VaultKey{}, err
	}
	defer zeroBytes(keyMaterial)

	keyID := strings.TrimSpace(config.VaultMasterKeyID)
	if keyID == "" {
		keyID = ulid.Make().Prefixed("vkey")
	}

	return model.VaultKey{
		ID:          keyID,
		Algorithm:   vaultMasterKeyAlgorithm,
		KeyMaterial: base64.StdEncoding.EncodeToString(keyMaterial),
		Active:      true,
	}, nil
}

func validateVaultKey(key model.VaultKey) error {
	if strings.TrimSpace(key.ID) == "" {
		return fmt.Errorf("%w: key id is required", ErrVaultMasterKeyInvalid)
	}
	if key.Algorithm != vaultMasterKeyAlgorithm {
		return fmt.Errorf("%w: %s", ErrUnsupportedSecretEncryptionAlgorithm, key.Algorithm)
	}
	keyMaterial, err := decodeVaultMasterKey(key.KeyMaterial)
	if err != nil {
		return err
	}
	zeroBytes(keyMaterial)
	return nil
}
