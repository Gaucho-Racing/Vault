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

func ListAccounts(c *gin.Context) {
	accounts, err := service.GetAllAccounts()
	if err != nil {
		if errors.Is(err, service.ErrAccountNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func GetAccount(c *gin.Context) {
	account, err := service.GetAccountWithSecrets(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, account)
}

func CreateAccount(c *gin.Context) {
	var req accountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	account, err := service.CreateAccount(model.Account{
		Name:              req.Name,
		Description:       req.Description,
		URL:               req.URL,
		AccessGroupNames:  req.AccessGroupNames,
		CreatedByEntityID: GetRequestEntityID(c),
		UpdatedByEntityID: GetRequestEntityID(c),
	})
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
	existing, err := service.GetAccountByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	updated, err := service.UpdateAccount(existing)
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

func ArchiveAccount(c *gin.Context) {
	id := c.Param("id")
	if _, err := service.GetAccountByID(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := service.ArchiveAccount(id, GetRequestEntityID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account archived"})
}
