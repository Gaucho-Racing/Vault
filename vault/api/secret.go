package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type secretRequest struct {
	Key        string `json:"key" binding:"required"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	Sensitive  bool   `json:"sensitive"`
	PlainValue string `json:"plain_value"`
}

func ListSecrets(c *gin.Context) {
	accountID := c.Param("id")
	if _, err := service.GetAccountByID(accountID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	secrets, err := service.GetSecretsForAccount(accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, secrets)
}

func CreateSecret(c *gin.Context) {
	accountID := c.Param("id")
	var req secretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	secret, err := service.CreateSecret(model.Secret{
		AccountID:         accountID,
		Key:               req.Key,
		Label:             req.Label,
		Type:              req.Type,
		Sensitive:         req.Sensitive,
		PlainValue:        req.PlainValue,
		CreatedByEntityID: GetRequestEntityID(c),
		UpdatedByEntityID: GetRequestEntityID(c),
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		if errors.Is(err, service.ErrSecretKeyRequired) || errors.Is(err, service.ErrSensitiveSecretValueRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "secret key already exists for account"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, secret)
}

func GetSecret(c *gin.Context) {
	secret, err := service.GetSecretForAccount(c.Param("id"), c.Param("secretID"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, secret)
}

func UpdateSecret(c *gin.Context) {
	existing, err := service.GetSecretForAccount(c.Param("id"), c.Param("secretID"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req secretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing.Key = req.Key
	existing.Label = req.Label
	existing.Type = req.Type
	existing.Sensitive = req.Sensitive
	existing.PlainValue = req.PlainValue
	existing.UpdatedByEntityID = GetRequestEntityID(c)

	updated, err := service.UpdateSecret(existing)
	if err != nil {
		if errors.Is(err, service.ErrSecretKeyRequired) || errors.Is(err, service.ErrSensitiveSecretValueRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "secret key already exists for account"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func ArchiveSecret(c *gin.Context) {
	if err := service.ArchiveSecret(c.Param("id"), c.Param("secretID"), GetRequestEntityID(c)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "secret archived"})
}

func RevealSecret(c *gin.Context) {
	secret, err := service.GetSecretForAccount(c.Param("id"), c.Param("secretID"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	value, err := service.RevealSecret(secret)
	if err != nil {
		if errors.Is(err, service.ErrSensitiveSecretValueMissing) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"value": value})
}
