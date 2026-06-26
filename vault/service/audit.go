package service

import (
	"errors"
	"strings"
	"time"

	"github.com/gaucho-racing/ulid-go"
	"github.com/gaucho-racing/vault/vault/database"
	"github.com/gaucho-racing/vault/vault/model"
	"gorm.io/gorm"
)

const (
	AuditActionAccountCreated = "account.created"
	AuditActionAccountUpdated = "account.updated"
	AuditActionAccountDeleted = "account.deleted"
	AuditActionAccountViewed  = "account.viewed"
	AuditActionSecretViewed   = "secret.viewed"
)

var ErrAuditActionRequired = errors.New("audit action is required")
var ErrAuditActorRequired = errors.New("audit actor is required")
var ErrAuditAccountRequired = errors.New("audit account is required")

func RecordAuditLog(auditLog model.AuditLog) error {
	return recordAuditLog(database.DB, auditLog)
}

func RecordAccountViewAuditLog(auditLog model.AuditLog, window time.Duration) error {
	return recordViewAuditLog(database.DB, auditLog, window)
}

func RecordViewAuditLog(auditLog model.AuditLog, window time.Duration) error {
	return recordViewAuditLog(database.DB, auditLog, window)
}

func GetAuditLogsForAccount(accountID string, limit int) ([]model.AuditLog, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return []model.AuditLog{}, ErrAuditAccountRequired
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	auditLogs := []model.AuditLog{}
	if err := database.DB.
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Limit(limit).
		Find(&auditLogs).Error; err != nil {
		return []model.AuditLog{}, err
	}
	return auditLogs, nil
}

func recordViewAuditLog(db *gorm.DB, auditLog model.AuditLog, window time.Duration) error {
	if window <= 0 {
		return recordAuditLog(db, auditLog)
	}
	normalizeAuditLog(&auditLog)
	if auditLog.Action == "" {
		return ErrAuditActionRequired
	}
	if auditLog.ActorEntityID == "" {
		return ErrAuditActorRequired
	}
	if auditLog.AccountID == "" {
		return ErrAuditAccountRequired
	}

	var count int64
	if err := recentViewAuditLogQuery(db, auditLog, time.Now().Add(-window)).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return recordAuditLog(db, auditLog)
}

func recentViewAuditLogQuery(db *gorm.DB, auditLog model.AuditLog, since time.Time) *gorm.DB {
	return db.Model(&model.AuditLog{}).
		Where(
			"action = ? AND account_id = ? AND actor_entity_id = ? AND created_at >= ?",
			auditLog.Action,
			auditLog.AccountID,
			auditLog.ActorEntityID,
			since,
		).Where("secret_id = ?", auditLog.SecretID)
}

func recordAuditLog(db *gorm.DB, auditLog model.AuditLog) error {
	if auditLog.ID == "" {
		auditLog.ID = ulid.Make().Prefixed("audit")
	}
	normalizeAuditLog(&auditLog)
	if auditLog.Action == "" {
		return ErrAuditActionRequired
	}
	if auditLog.ActorEntityID == "" {
		return ErrAuditActorRequired
	}
	if auditLog.AccountID == "" {
		return ErrAuditAccountRequired
	}
	return db.Create(&auditLog).Error
}

func normalizeAuditLog(auditLog *model.AuditLog) {
	auditLog.Action = strings.TrimSpace(auditLog.Action)
	auditLog.ActorEntityID = strings.TrimSpace(auditLog.ActorEntityID)
	auditLog.ActorUserID = strings.TrimSpace(auditLog.ActorUserID)
	auditLog.ActorGroupNames = normalizeStringSlice(auditLog.ActorGroupNames)
	auditLog.AccountID = strings.TrimSpace(auditLog.AccountID)
	auditLog.AccountName = strings.TrimSpace(auditLog.AccountName)
	auditLog.SecretID = strings.TrimSpace(auditLog.SecretID)
	auditLog.SecretKey = strings.TrimSpace(auditLog.SecretKey)
	auditLog.SecretLabel = strings.TrimSpace(auditLog.SecretLabel)
	auditLog.RequestMethod = strings.TrimSpace(auditLog.RequestMethod)
	auditLog.RequestPath = strings.TrimSpace(auditLog.RequestPath)
	auditLog.IPAddress = strings.TrimSpace(auditLog.IPAddress)
	auditLog.UserAgent = strings.TrimSpace(auditLog.UserAgent)
}
