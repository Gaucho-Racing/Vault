package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/model"
)

const secretEncryptionAlgorithm = "AES-256-GCM-Envelope-v1"
const secretDataKeySize = 32

var ErrSensitiveSecretValueRequired = errors.New("sensitive secret value is required")
var ErrSensitiveSecretValueMissing = errors.New("sensitive secret has no encrypted value")
var ErrVaultMasterKeyInvalid = errors.New("vault master key is invalid")
var ErrUnsupportedSecretEncryptionAlgorithm = errors.New("unsupported secret encryption algorithm")

func encryptSecretValue(secret *model.Secret) error {
	if secret.PlainValue == "" {
		return ErrSensitiveSecretValueRequired
	}
	masterKey, err := loadVaultMasterKey()
	if err != nil {
		return err
	}
	defer zeroBytes(masterKey)

	dataKey := make([]byte, secretDataKeySize)
	defer zeroBytes(dataKey)
	if _, err := io.ReadFull(rand.Reader, dataKey); err != nil {
		return fmt.Errorf("generate secret data key: %w", err)
	}

	aead, err := newAESGCM(dataKey)
	if err != nil {
		return err
	}
	nonce, err := randomBytes(aead.NonceSize())
	if err != nil {
		return err
	}
	wrappedDataKey, err := wrapDataKey(masterKey, dataKey, config.VaultMasterKeyID)
	if err != nil {
		return err
	}

	secret.Ciphertext = aead.Seal(nil, nonce, []byte(secret.PlainValue), secretAssociatedData(*secret))
	secret.Nonce = nonce
	secret.EncryptedDataKey = wrappedDataKey
	secret.KeyID = config.VaultMasterKeyID
	secret.Algorithm = secretEncryptionAlgorithm
	secret.PlainValue = ""
	return nil
}

func decryptSecretValue(secret model.Secret) (string, error) {
	if !hasEncryptedSecretValue(secret) {
		return "", ErrSensitiveSecretValueMissing
	}
	if secret.Algorithm != secretEncryptionAlgorithm {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedSecretEncryptionAlgorithm, secret.Algorithm)
	}
	masterKey, err := loadVaultMasterKey()
	if err != nil {
		return "", err
	}
	defer zeroBytes(masterKey)
	dataKey, err := unwrapDataKey(masterKey, secret.EncryptedDataKey, secret.KeyID)
	if err != nil {
		return "", err
	}
	defer zeroBytes(dataKey)

	aead, err := newAESGCM(dataKey)
	if err != nil {
		return "", err
	}
	plaintext, err := aead.Open(nil, secret.Nonce, secret.Ciphertext, secretAssociatedData(secret))
	if err != nil {
		return "", fmt.Errorf("decrypt secret value: %w", err)
	}
	return string(plaintext), nil
}

func clearEncryptedSecretValue(secret *model.Secret) {
	secret.Ciphertext = nil
	secret.Nonce = nil
	secret.EncryptedDataKey = nil
	secret.KeyID = ""
	secret.Algorithm = ""
}

func hasEncryptedSecretValue(secret model.Secret) bool {
	return len(secret.Ciphertext) > 0 && len(secret.Nonce) > 0 && len(secret.EncryptedDataKey) > 0
}

func loadVaultMasterKey() ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(config.VaultMasterKey)
	if err != nil {
		return nil, fmt.Errorf("%w: must be base64 encoded", ErrVaultMasterKeyInvalid)
	}
	if len(key) != secretDataKeySize {
		return nil, fmt.Errorf("%w: must decode to %d bytes", ErrVaultMasterKeyInvalid, secretDataKeySize)
	}
	return key, nil
}

func wrapDataKey(masterKey []byte, dataKey []byte, keyID string) ([]byte, error) {
	aead, err := newAESGCM(masterKey)
	if err != nil {
		return nil, err
	}
	nonce, err := randomBytes(aead.NonceSize())
	if err != nil {
		return nil, err
	}
	wrapped := aead.Seal(nil, nonce, dataKey, []byte(keyID))
	return append(nonce, wrapped...), nil
}

func unwrapDataKey(masterKey []byte, encryptedDataKey []byte, keyID string) ([]byte, error) {
	aead, err := newAESGCM(masterKey)
	if err != nil {
		return nil, err
	}
	if len(encryptedDataKey) <= aead.NonceSize() {
		return nil, ErrSensitiveSecretValueMissing
	}
	nonce := encryptedDataKey[:aead.NonceSize()]
	wrapped := encryptedDataKey[aead.NonceSize():]
	dataKey, err := aead.Open(nil, nonce, wrapped, []byte(keyID))
	if err != nil {
		return nil, fmt.Errorf("unwrap secret data key: %w", err)
	}
	return dataKey, nil
}

func newAESGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create aes-gcm cipher: %w", err)
	}
	return aead, nil
}

func randomBytes(size int) ([]byte, error) {
	value := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, value); err != nil {
		return nil, fmt.Errorf("generate random bytes: %w", err)
	}
	return value, nil
}

func secretAssociatedData(secret model.Secret) []byte {
	return []byte(secret.AccountID + "\x00" + secret.ID)
}

func zeroBytes(value []byte) {
	for i := range value {
		value[i] = 0
	}
}
