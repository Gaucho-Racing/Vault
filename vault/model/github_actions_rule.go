package model

import "time"

type GitHubActionsRule struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	Name               string    `json:"name" gorm:"uniqueIndex"`
	RepositoryPatterns []string  `json:"repository_patterns" gorm:"type:jsonb;serializer:json"`
	RefPatterns        []string  `json:"ref_patterns" gorm:"type:jsonb;serializer:json"`
	SecretSelectors    []string  `json:"secret_selectors" gorm:"type:jsonb;serializer:json"`
	Enabled            bool      `json:"enabled"`
	CreatedByEntityID  string    `json:"created_by_entity_id" gorm:"index"`
	UpdatedByEntityID  string    `json:"updated_by_entity_id" gorm:"index"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (GitHubActionsRule) TableName() string {
	return "vault_github_actions_rule"
}
