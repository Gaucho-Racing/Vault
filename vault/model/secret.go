package model

import "time"

type Secret struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	AccountID         string    `json:"account_id" gorm:"index;uniqueIndex:idx_secret_account_key"`
	Key               string    `json:"key" gorm:"uniqueIndex:idx_secret_account_key"`
	Label             string    `json:"label"`
	Type              string    `json:"type" gorm:"index"`
	Sensitive         bool      `json:"sensitive" gorm:"index"`
	PlainValue        string    `json:"plain_value"`
	Ciphertext        []byte    `json:"-" gorm:"type:bytea"`
	Nonce             []byte    `json:"-" gorm:"type:bytea"`
	EncryptedDataKey  []byte    `json:"-" gorm:"type:bytea"`
	KeyID             string    `json:"key_id" gorm:"index"`
	Algorithm         string    `json:"algorithm"`
	CreatedByEntityID string    `json:"created_by_entity_id" gorm:"index"`
	UpdatedByEntityID string    `json:"updated_by_entity_id" gorm:"index"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Secret) TableName() string {
	return "vault_account_secret"
}
