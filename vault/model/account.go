package model

import "time"

type Account struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	Name              string     `json:"name" gorm:"index"`
	Description       string     `json:"description"`
	URL               string     `json:"url"`
	AccessGroupNames  []string   `json:"access_group_names" gorm:"type:jsonb;serializer:json"`
	CreatedByEntityID string     `json:"created_by_entity_id" gorm:"index"`
	UpdatedByEntityID string     `json:"updated_by_entity_id" gorm:"index"`
	DeletedAt         *time.Time `json:"deleted_at" gorm:"index"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Account) TableName() string {
	return "vault_account"
}
