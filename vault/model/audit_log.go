package model

import "time"

type AuditLog struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Action          string    `json:"action" gorm:"index"`
	ActorEntityID   string    `json:"actor_entity_id" gorm:"index"`
	ActorUserID     string    `json:"actor_user_id" gorm:"index"`
	ActorGroupNames []string  `json:"actor_group_names" gorm:"type:jsonb;serializer:json"`
	AccountID       string    `json:"account_id" gorm:"index"`
	AccountName     string    `json:"account_name"`
	SecretID        string    `json:"secret_id" gorm:"index"`
	SecretKey       string    `json:"secret_key"`
	SecretLabel     string    `json:"secret_label"`
	RequestMethod   string    `json:"request_method"`
	RequestPath     string    `json:"request_path"`
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

func (AuditLog) TableName() string {
	return "vault_audit_log"
}
