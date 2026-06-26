package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"github.com/gaucho-racing/vault/vault/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

func GetCurrentUser(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	userID := GetRequestTokenUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current token is not a user token"})
		return
	}

	user, err := sentinel.GetCurrentUser(GetRequestToken(c), userID)
	if err != nil {
		logger.SugarLogger.Errorln("Failed to get current Sentinel user: " + err.Error())
		var sentinelErr sentinel.Error
		if errors.As(err, &sentinelErr) && sentinelErr.Code >= http.StatusBadRequest && sentinelErr.Code < http.StatusInternalServerError {
			c.JSON(sentinelErr.Code, gin.H{"error": sentinelErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
