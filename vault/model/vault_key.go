package model

import "time"

type VaultKey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Algorithm   string    `json:"algorithm"`
	KeyMaterial string    `json:"-"`
	Active      bool      `json:"active" gorm:"index;uniqueIndex:idx_vault_key_active,where:active = true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (VaultKey) TableName() string {
	return "vault_key"
}
