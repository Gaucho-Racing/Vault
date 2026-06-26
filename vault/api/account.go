package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type accountRequest struct {
	Name             string   `json:"name" binding:"required"`
	Description      string   `json:"description"`
	URL              string   `json:"url"`
	AccessGroupNames []string `json:"access_group_names"`
}

type accountListItem struct {
	service.AccountWithSecretCount
	CanAccess bool `json:"can_access"`
}

func ListAccounts(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	accounts, err := service.GetAllAccounts()
	if err != nil {
		if errors.Is(err, service.ErrAccountNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]accountListItem, 0, len(accounts))
	for _, account := range accounts {
		response = append(response, accountListItem{
			AccountWithSecretCount: account,
			CanAccess:              RequestTokenCanAccessAccount(c, account.Account),
		})
	}
	c.JSON(http.StatusOK, response)
}

func GetAccount(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	account, err := service.GetAccountWithSecrets(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessAccount(c, account.Account))
	if err := service.RecordAccountViewAuditLog(newAccountAuditLog(c, service.AuditActionAccountViewed, account.Account), auditViewDebounceWindow); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, account)
}

func CreateAccount(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	var req accountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	account, err := service.CreateAccountWithAudit(model.Account{
		Name:              req.Name,
		Description:       req.Description,
		URL:               req.URL,
		AccessGroupNames:  req.AccessGroupNames,
		CreatedByEntityID: GetRequestEntityID(c),
		UpdatedByEntityID: GetRequestEntityID(c),
	}, newAccountAuditLog(c, service.AuditActionAccountCreated, model.Account{}))
	if err != nil {
		if errors.Is(err, service.ErrAccountNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, account)
}

func UpdateAccount(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	existing, err := service.GetAccountByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessAccount(c, existing))

	var req accountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing.Name = req.Name
	existing.Description = req.Description
	existing.URL = req.URL
	existing.AccessGroupNames = req.AccessGroupNames
	existing.UpdatedByEntityID = GetRequestEntityID(c)

	updated, err := service.UpdateAccountWithAudit(existing, newAccountAuditLog(c, service.AuditActionAccountUpdated, existing))
	if err != nil {
		if errors.Is(err, service.ErrAccountNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteAccount(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	id := c.Param("id")
	account, err := service.GetAccountByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessAccount(c, account))
	if err := service.DeleteAccountWithAudit(account, newAccountAuditLog(c, service.AuditActionAccountDeleted, account)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
