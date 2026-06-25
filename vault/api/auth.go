package api

import "github.com/gin-gonic/gin"

func GetRequestEntityID(c *gin.Context) string {
	entityID, exists := c.Get("Auth-EntityID")
	if !exists {
		return ""
	}
	value, ok := entityID.(string)
	if !ok {
		return ""
	}
	return value
}
