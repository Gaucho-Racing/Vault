package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type applicationRequest struct {
	Name                      string   `json:"name" binding:"required"`
	AccessGroupNames          []string `json:"access_group_names"`
	GitHubActionsRepositories []string `json:"github_actions_repositories"`
	GitHubActionsRefs         []string `json:"github_actions_refs"`
}

type appSecretRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

type applicationListItem struct {
	service.ApplicationWithSecretCount
	CanAccess bool `json:"can_access"`
}

func ListApplications(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	applications, err := service.GetAllApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]applicationListItem, 0, len(applications))
	for _, application := range applications {
		response = append(response, applicationListItem{
			ApplicationWithSecretCount: application,
			CanAccess:                  RequestTokenCanAccessApplication(c, application.Application),
		})
	}
	c.JSON(http.StatusOK, response)
}

func GetApplication(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationWithSecrets(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application.Application))
	c.JSON(http.StatusOK, application)
}

func CreateApplication(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	var req applicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	application, err := service.CreateApplication(model.Application{
		Name:                      req.Name,
		AccessGroupNames:          req.AccessGroupNames,
		GitHubActionsRepositories: req.GitHubActionsRepositories,
		GitHubActionsRefs:         req.GitHubActionsRefs,
		CreatedByEntityID:         GetRequestEntityID(c),
		UpdatedByEntityID:         GetRequestEntityID(c),
	})
	if err != nil {
		handleApplicationError(c, err)
		return
	}
	c.JSON(http.StatusOK, application)
}

func UpdateApplication(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application))

	var req applicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	application.Name = req.Name
	application.AccessGroupNames = req.AccessGroupNames
	application.GitHubActionsRepositories = req.GitHubActionsRepositories
	application.GitHubActionsRefs = req.GitHubActionsRefs
	application.UpdatedByEntityID = GetRequestEntityID(c)

	updated, err := service.UpdateApplication(application)
	if err != nil {
		handleApplicationError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteApplication(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application))
	if err := service.DeleteApplication(application); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "application deleted"})
}

func CreateApplicationSecret(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application))

	var req appSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	secret, err := service.CreateAppSecret(model.AppSecret{
		ApplicationID:     application.ID,
		Key:               req.Key,
		PlainValue:        req.Value,
		CreatedByEntityID: GetRequestEntityID(c),
		UpdatedByEntityID: GetRequestEntityID(c),
	})
	if err != nil {
		handleAppSecretError(c, err)
		return
	}
	c.JSON(http.StatusOK, secret)
}

func UpdateApplicationSecret(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application))

	secret, err := service.GetAppSecretForApplication(application.ID, c.Param("secretID"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "app secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req appSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	secret.Key = req.Key
	secret.PlainValue = req.Value
	secret.UpdatedByEntityID = GetRequestEntityID(c)

	updated, err := service.UpdateAppSecret(secret)
	if err != nil {
		handleAppSecretError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteApplicationSecret(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application))
	if err := service.DeleteAppSecret(application.ID, c.Param("secretID")); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "app secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "app secret deleted"})
}

func RevealApplicationSecret(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application))

	secret, err := service.GetAppSecretForApplication(application.ID, c.Param("secretID"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "app secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	value, err := service.RevealAppSecret(secret)
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

func DownloadApplicationEnvFile(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	application, err := service.GetApplicationByID(c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	Require(c, RequestTokenCanAccessApplication(c, application))

	envFile, err := service.BuildApplicationEnvFile(application.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+application.Name+`.env"`)
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(envFile))
}

func handleApplicationError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrApplicationNameRequired) ||
		errors.Is(err, service.ErrApplicationNameInvalid) ||
		errors.Is(err, service.ErrGitHubActionsRepositoryRequired) ||
		errors.Is(err, service.ErrGitHubActionsRefRequired) ||
		errors.Is(err, service.ErrGitHubActionsRepositoryInvalid) ||
		errors.Is(err, service.ErrGitHubActionsRefInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"error": "application name already exists"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func handleAppSecretError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrAppSecretKeyRequired) ||
		errors.Is(err, service.ErrAppSecretKeyInvalid) ||
		errors.Is(err, service.ErrAppSecretValueRequired) ||
		errors.Is(err, service.ErrSensitiveSecretValueRequired) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.JSON(http.StatusConflict, gin.H{"error": "app secret key already exists for application"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
