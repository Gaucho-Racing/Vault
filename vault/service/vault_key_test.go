package service

import (
	"strings"
	"testing"

	"github.com/gaucho-racing/vault/vault/config"
)

func TestBuildVaultKeyGeneratesKeyWithoutConfig(t *testing.T) {
	restoreVaultKeyConfig(t)
	config.VaultMasterKey = ""
	config.VaultMasterKeyID = ""

	key, err := buildVaultKey()
	if err != nil {
		t.Fatalf("buildVaultKey() error = %v", err)
	}
	if !strings.HasPrefix(key.ID, "vkey_") {
		t.Fatalf("buildVaultKey() id = %q", key.ID)
	}
	if key.Algorithm != vaultMasterKeyAlgorithm {
		t.Fatalf("buildVaultKey() algorithm = %q", key.Algorithm)
	}
	if !key.Active {
		t.Fatalf("buildVaultKey() active = false")
	}
	keyMaterial, err := decodeVaultMasterKey(key.KeyMaterial)
	if err != nil {
		t.Fatalf("decodeVaultMasterKey() error = %v", err)
	}
	defer zeroBytes(keyMaterial)
	if len(keyMaterial) != secretDataKeySize {
		t.Fatalf("decoded key length = %d", len(keyMaterial))
	}
}

func TestBuildVaultKeyImportsConfiguredKey(t *testing.T) {
	restoreVaultKeyConfig(t)
	config.VaultMasterKey = " MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY= "
	config.VaultMasterKeyID = "local-dev"

	key, err := buildVaultKey()
	if err != nil {
		t.Fatalf("buildVaultKey() error = %v", err)
	}
	if key.ID != "local-dev" {
		t.Fatalf("buildVaultKey() id = %q", key.ID)
	}
	if key.KeyMaterial != "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=" {
		t.Fatalf("buildVaultKey() key material = %q", key.KeyMaterial)
	}
	if err := validateVaultKey(key); err != nil {
		t.Fatalf("validateVaultKey() error = %v", err)
	}
}

func TestBuildVaultKeyRejectsInvalidConfiguredKey(t *testing.T) {
	restoreVaultKeyConfig(t)
	config.VaultMasterKey = "not-base64"
	config.VaultMasterKeyID = "local-dev"

	if _, err := buildVaultKey(); err == nil {
		t.Fatalf("buildVaultKey() error = nil")
	}
}

func restoreVaultKeyConfig(t *testing.T) {
	t.Helper()
	oldMasterKey := config.VaultMasterKey
	oldMasterKeyID := config.VaultMasterKeyID
	t.Cleanup(func() {
		config.VaultMasterKey = oldMasterKey
		config.VaultMasterKeyID = oldMasterKeyID
	})
}
