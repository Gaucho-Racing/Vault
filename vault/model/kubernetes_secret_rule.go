package model

import "time"

type KubernetesSecretRule struct {
	ID                     string    `json:"id" gorm:"primaryKey"`
	Name                   string    `json:"name" gorm:"uniqueIndex"`
	ClusterPatterns        []string  `json:"cluster_patterns" gorm:"type:jsonb;serializer:json"`
	NamespacePatterns      []string  `json:"namespace_patterns" gorm:"type:jsonb;serializer:json"`
	ServiceAccountPatterns []string  `json:"service_account_patterns" gorm:"type:jsonb;serializer:json"`
	SecretSelectors        []string  `json:"secret_selectors" gorm:"type:jsonb;serializer:json"`
	Enabled                bool      `json:"enabled"`
	CreatedByEntityID      string    `json:"created_by_entity_id" gorm:"index"`
	UpdatedByEntityID      string    `json:"updated_by_entity_id" gorm:"index"`
	CreatedAt              time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (KubernetesSecretRule) TableName() string {
	return "vault_kubernetes_secret_rule"
}
