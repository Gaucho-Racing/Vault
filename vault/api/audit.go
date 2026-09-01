package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"github.com/gaucho-racing/vault/vault/pkg/sentinel"
	"github.com/gaucho-racing/vault/vault/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const auditViewDebounceWindow = 2 * time.Minute

type auditLogResponse struct {
	model.AuditLog
	Actor *sentinel.IdentitySummary `json:"actor,omitempty"`
}

func ListAccountAuditLogs(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	account, err := service.GetAccountByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanViewAuditLogs(c))

	limit := 50
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 || parsedLimit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return
		}
		limit = parsedLimit
	}

	auditLogs, err := service.GetAuditLogsForAccount(account.ID, limit)
	if err != nil {
		if errors.Is(err, service.ErrAuditAccountRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buildAuditLogResponses(c, auditLogs))
}

func newAccountAuditLog(c *gin.Context, action string, account model.Account) model.AuditLog {
	return model.AuditLog{
		Action:          action,
		ActorEntityID:   GetRequestEntityID(c),
		ActorUserID:     GetRequestTokenUserID(c),
		ActorGroupNames: GetRequestTokenGroupNames(c),
		AccountID:       account.ID,
		AccountName:     account.Name,
		RequestMethod:   c.Request.Method,
		RequestPath:     c.Request.URL.Path,
		IPAddress:       c.ClientIP(),
		UserAgent:       c.Request.UserAgent(),
	}
}

func newSecretAuditLog(c *gin.Context, action string, account model.Account, secret model.Secret) model.AuditLog {
	auditLog := newAccountAuditLog(c, action, account)
	auditLog.SecretID = secret.ID
	auditLog.SecretKey = secret.Key
	auditLog.SecretLabel = secret.Label
	return auditLog
}

func buildAuditLogResponses(c *gin.Context, auditLogs []model.AuditLog) []auditLogResponse {
	entityIDs := make([]string, 0, len(auditLogs))
	seenEntityIDs := make(map[string]struct{}, len(auditLogs))
	for _, auditLog := range auditLogs {
		if auditLog.ActorEntityID == "" {
			continue
		}
		if _, exists := seenEntityIDs[auditLog.ActorEntityID]; exists {
			continue
		}
		seenEntityIDs[auditLog.ActorEntityID] = struct{}{}
		entityIDs = append(entityIDs, auditLog.ActorEntityID)
	}

	actorsByEntityID := make(map[string]*sentinel.IdentitySummary, len(entityIDs))
	actors, err := sentinel.ResolveIdentities(c.Request.Context(), GetRequestToken(c), entityIDs)
	if err != nil {
		logger.SugarLogger.Warnf("failed to hydrate audit actors: %v", err)
	} else {
		for index := range actors {
			actorsByEntityID[actors[index].ID] = &actors[index]
		}
	}

	responses := make([]auditLogResponse, 0, len(auditLogs))
	for _, auditLog := range auditLogs {
		responses = append(responses, auditLogResponse{
			AuditLog: auditLog,
			Actor:    actorsByEntityID[auditLog.ActorEntityID],
		})
	}
	return responses
}
