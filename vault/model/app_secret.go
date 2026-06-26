package model

import "time"

type AppSecret struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	ApplicationID     string     `json:"application_id" gorm:"index;index:idx_app_secret_application_key_live,unique,where:deleted_at IS NULL"`
	Key               string     `json:"key" gorm:"index:idx_app_secret_application_key_live,unique,where:deleted_at IS NULL"`
	Ciphertext        []byte     `json:"-" gorm:"type:bytea"`
	Nonce             []byte     `json:"-" gorm:"type:bytea"`
	EncryptedDataKey  []byte     `json:"-" gorm:"type:bytea"`
	KeyID             string     `json:"key_id" gorm:"index"`
	Algorithm         string     `json:"algorithm"`
	CreatedByEntityID string     `json:"created_by_entity_id" gorm:"index"`
	UpdatedByEntityID string     `json:"updated_by_entity_id" gorm:"index"`
	DeletedAt         *time.Time `json:"deleted_at" gorm:"index"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	PlainValue        string     `json:"-" gorm:"-"`
}

func (AppSecret) TableName() string {
	return "vault_app_secret"
}
