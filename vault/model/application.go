package model

import "time"

type Application struct {
	ID                        string      `json:"id" gorm:"primaryKey"`
	Name                      string      `json:"name" gorm:"uniqueIndex"`
	AccessGroupNames          []string    `json:"access_group_names" gorm:"type:jsonb;serializer:json"`
	GitHubActionsRepositories []string    `json:"github_actions_repositories" gorm:"type:jsonb;serializer:json"`
	GitHubActionsRefs         []string    `json:"github_actions_refs" gorm:"type:jsonb;serializer:json"`
	CreatedByEntityID         string      `json:"created_by_entity_id" gorm:"index"`
	UpdatedByEntityID         string      `json:"updated_by_entity_id" gorm:"index"`
	CreatedAt                 time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	Secrets                   []AppSecret `json:"secrets,omitempty" gorm:"foreignKey:ApplicationID"`
}

func (Application) TableName() string {
	return "vault_application"
}
