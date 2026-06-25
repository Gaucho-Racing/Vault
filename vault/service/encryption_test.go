package service

import (
	"testing"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/model"
)

func TestSecretEncryptionRoundTripPreservesValue(t *testing.T) {
	config.VaultMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
	config.VaultMasterKeyID = "test-key"

	secret := model.Secret{
		ID:         "sec_test",
		AccountID:  "acct_test",
		Sensitive:  true,
		PlainValue: "  super-secret value  ",
	}

	if err := encryptSecretValue(&secret); err != nil {
		t.Fatalf("encryptSecretValue() error = %v", err)
	}
	if secret.PlainValue != "" {
		t.Fatalf("encryptSecretValue() left plaintext on model")
	}
	if len(secret.Ciphertext) == 0 || len(secret.Nonce) == 0 || len(secret.EncryptedDataKey) == 0 {
		t.Fatalf("encryptSecretValue() did not populate encrypted fields")
	}

	value, err := decryptSecretValue(secret)
	if err != nil {
		t.Fatalf("decryptSecretValue() error = %v", err)
	}
	if value != "  super-secret value  " {
		t.Fatalf("decryptSecretValue() = %q", value)
	}
}

func TestSecretEncryptionRejectsAssociatedDataMismatch(t *testing.T) {
	config.VaultMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
	config.VaultMasterKeyID = "test-key"

	secret := model.Secret{
		ID:         "sec_test",
		AccountID:  "acct_test",
		Sensitive:  true,
		PlainValue: "super-secret value",
	}

	if err := encryptSecretValue(&secret); err != nil {
		t.Fatalf("encryptSecretValue() error = %v", err)
	}

	secret.AccountID = "acct_other"
	if _, err := decryptSecretValue(secret); err == nil {
		t.Fatalf("decryptSecretValue() error = nil")
	}
}
