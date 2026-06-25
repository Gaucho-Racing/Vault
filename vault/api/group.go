package api

import (
	"errors"
	"net/http"

	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"github.com/gaucho-racing/vault/vault/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

func ListSentinelGroups(c *gin.Context) {
	Require(c, RequestTokenExists(c))

	groups, err := sentinel.GetGroups(GetRequestToken(c))
	if err != nil {
		logger.SugarLogger.Errorln("Failed to list Sentinel groups: " + err.Error())
		var sentinelErr sentinel.Error
		if errors.As(err, &sentinelErr) && sentinelErr.Code >= http.StatusBadRequest && sentinelErr.Code < http.StatusInternalServerError {
			c.JSON(sentinelErr.Code, gin.H{"error": sentinelErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}
