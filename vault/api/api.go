package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"github.com/gaucho-racing/vault/vault/pkg/sentinel"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run() {
	api := InitializeRouter()
	InitializeRoutes(api)
	err := api.Run(":" + config.Port)
	if err != nil {
		logger.SugarLogger.Fatalf("Failed to start server: %v", err)
	}
}

func InitializeRouter() *gin.Engine {
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: false,
	}))
	r.Use(AuthChecker())
	r.Use(UnauthorizedPanicHandler())
	return r
}

func InitializeRoutes(router *gin.Engine) {
	router.GET("/ping", Ping)

	router.POST("/integrations/github/actions/env", ExportGitHubActionsEnv)

	router.POST("/auth/login", LoginWithSentinel)
	router.GET("/auth/session", GetSession)
	router.POST("/auth/refresh", RefreshSession)
	router.POST("/auth/logout", Logout)
	router.GET("/users/@me", GetCurrentUser)
	router.GET("/groups", ListSentinelGroups)

	router.GET("/integrations/github/actions/rules", ListGitHubActionsRules)
	router.POST("/integrations/github/actions/rules", CreateGitHubActionsRule)
	router.PUT("/integrations/github/actions/rules/:id", UpdateGitHubActionsRule)
	router.DELETE("/integrations/github/actions/rules/:id", DeleteGitHubActionsRule)

	router.POST("/secrets/totp/qr", DecodeTOTPRegistrationQRCode)

	router.GET("/app-secrets", ListApplications)
	router.POST("/app-secrets", CreateApplication)
	router.GET("/app-secrets/:id", GetApplication)
	router.PUT("/app-secrets/:id", UpdateApplication)
	router.DELETE("/app-secrets/:id", DeleteApplication)
	router.GET("/app-secrets/:id/env", DownloadApplicationEnvFile)
	router.POST("/app-secrets/:id/secrets", CreateApplicationSecret)
	router.PUT("/app-secrets/:id/secrets/:secretID", UpdateApplicationSecret)
	router.DELETE("/app-secrets/:id/secrets/:secretID", DeleteApplicationSecret)
	router.POST("/app-secrets/:id/secrets/:secretID/reveal", RevealApplicationSecret)

	router.GET("/accounts", ListAccounts)
	router.POST("/accounts", CreateAccount)
	router.GET("/accounts/:id", GetAccount)
	router.GET("/accounts/:id/audit-logs", ListAccountAuditLogs)
	router.PUT("/accounts/:id", UpdateAccount)
	router.DELETE("/accounts/:id", DeleteAccount)

	router.GET("/accounts/:id/secrets", ListSecrets)
	router.POST("/accounts/:id/secrets", CreateSecret)
	router.GET("/accounts/:id/secrets/:secretID", GetSecret)
	router.PUT("/accounts/:id/secrets/:secretID", UpdateSecret)
	router.DELETE("/accounts/:id/secrets/:secretID", DeleteSecret)
	router.POST("/accounts/:id/secrets/:secretID/reveal", RevealSecret)
	router.POST("/accounts/:id/secrets/:secretID/totp", GenerateTOTPCode)
}

func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authRouteSkipsTokenValidation(c.Request.URL.Path) {
			c.Next()
			return
		}

		token := ""
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if token != "" {
			claims, err := sentinel.ValidateToken(token)
			if err != nil {
				logger.SugarLogger.Errorln("Failed to validate token: " + err.Error())
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			setAuthContext(c, token, claims)
			logger.SugarLogger.Infof("Decoded token: entity=%s audience=%s scope=%s", GetRequestTokenEntityID(c), GetRequestTokenAudience(c), GetRequestTokenScopes(c))
		}
		c.Next()
	}
}

func authRouteSkipsTokenValidation(path string) bool {
	return path == "/auth/login" ||
		path == "/auth/refresh" ||
		path == "/auth/logout" ||
		path == "/integrations/github/actions/env"
}

func UnauthorizedPanicHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if err == "Unauthorized" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "you are not authorized to access this resource"})
					return
				}
				logger.SugarLogger.Errorf("Unexpected panic: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprint(err)})
			}
		}()
		c.Next()
	}
}

func Require(c *gin.Context, condition bool) {
	if !condition {
		panic("Unauthorized")
	}
}

func Any(conditions ...bool) bool {
	for _, condition := range conditions {
		if condition {
			return true
		}
	}
	return false
}

func All(conditions ...bool) bool {
	for _, condition := range conditions {
		if !condition {
			return false
		}
	}
	return true
}

func RequestTokenExists(c *gin.Context) bool {
	_, exists := c.Get("Auth-Token")
	return exists
}

func RequestTokenHasScope(c *gin.Context, scope string) bool {
	for _, tokenScope := range strings.Fields(GetRequestTokenScopes(c)) {
		if tokenScope == scope {
			return true
		}
	}
	return false
}

func RequestTokenHasAudience(c *gin.Context, audience string) bool {
	return GetRequestTokenAudience(c) == audience
}

func RequestTokenHasEntityID(c *gin.Context, entityID string) bool {
	return GetRequestTokenEntityID(c) == entityID
}

func RequestTokenHasGroupName(c *gin.Context, groupName string) bool {
	for _, tokenGroup := range GetRequestTokenGroupNames(c) {
		if tokenGroup == groupName {
			return true
		}
	}
	return false
}

func RequestTokenHasAnyGroupName(c *gin.Context, groupNames []string) bool {
	for _, groupName := range groupNames {
		if RequestTokenHasGroupName(c, groupName) {
			return true
		}
	}
	return false
}

func RequestTokenCanAccessAccount(c *gin.Context, account model.Account) bool {
	if !RequestTokenExists(c) {
		return false
	}
	if RequestTokenHasScope(c, "sentinel:all") {
		return true
	}
	if RequestTokenHasGroupName(c, "Admins") {
		return true
	}
	if len(account.AccessGroupNames) == 0 {
		return true
	}
	return RequestTokenHasAnyGroupName(c, account.AccessGroupNames)
}

func RequestTokenCanAccessApplication(c *gin.Context, application model.Application) bool {
	if !RequestTokenExists(c) {
		return false
	}
	if RequestTokenHasScope(c, "sentinel:all") {
		return true
	}
	if RequestTokenHasGroupName(c, "Admins") {
		return true
	}
	if len(application.AccessGroupNames) == 0 {
		return true
	}
	return RequestTokenHasAnyGroupName(c, application.AccessGroupNames)
}

func RequestTokenCanViewAuditLogs(c *gin.Context) bool {
	return RequestTokenHasScope(c, "sentinel:all") || RequestTokenHasGroupName(c, "Admins")
}

func RequestTokenCanManageSettings(c *gin.Context) bool {
	return RequestTokenHasScope(c, "sentinel:all") || RequestTokenHasGroupName(c, "Admins")
}

func RequestTokenCanManageGitHubActionsRules(c *gin.Context) bool {
	return RequestTokenCanManageSettings(c) || RequestTokenHasGroupName(c, "DevopsMembers")
}

func GetRequestToken(c *gin.Context) string {
	token, _ := c.Get("Auth-Token")
	return contextString(token)
}

func GetRequestTokenScopes(c *gin.Context) string {
	scopes, _ := c.Get("Auth-Scope")
	return contextString(scopes)
}

func GetRequestTokenAudience(c *gin.Context) string {
	audience, _ := c.Get("Auth-Audience")
	return contextString(audience)
}

func GetRequestTokenClaims(c *gin.Context) map[string]interface{} {
	claims, exists := c.Get("Auth-Claims")
	if !exists {
		return nil
	}
	value, ok := claims.(map[string]interface{})
	if !ok {
		return nil
	}
	return value
}

func GetRequestTokenEntityID(c *gin.Context) string {
	entityID, _ := c.Get("Auth-EntityID")
	return contextString(entityID)
}

func GetRequestEntityID(c *gin.Context) string {
	return GetRequestTokenEntityID(c)
}

func GetRequestTokenUserID(c *gin.Context) string {
	return claimString(GetRequestTokenClaims(c), "user_id")
}

func GetRequestTokenGroupNames(c *gin.Context) []string {
	return claimStringSlice(GetRequestTokenClaims(c), "groups")
}

func setAuthContext(c *gin.Context, token string, claims map[string]interface{}) {
	c.Set("Auth-Token", token)
	c.Set("Auth-Claims", claims)
	c.Set("Auth-EntityID", claimString(claims, "sub"))
	c.Set("Auth-Scope", claimString(claims, "scope"))
	c.Set("Auth-UserID", claimString(claims, "user_id"))
	audiences := claimStringSlice(claims, "aud")
	if len(audiences) > 0 {
		c.Set("Auth-Audience", audiences[0])
	}
}

func contextString(value interface{}) string {
	if value == nil {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func claimString(claims map[string]interface{}, key string) string {
	if claims == nil {
		return ""
	}
	value, ok := claims[key].(string)
	if !ok {
		return ""
	}
	return value
}

func claimStringSlice(claims map[string]interface{}, key string) []string {
	if claims == nil {
		return []string{}
	}
	switch value := claims[key].(type) {
	case []string:
		return value
	case []interface{}:
		result := make([]string, 0, len(value))
		for _, item := range value {
			if str, ok := item.(string); ok && str != "" {
				result = append(result, str)
			}
		}
		return result
	case string:
		if value == "" {
			return []string{}
		}
		return []string{value}
	default:
		return []string{}
	}
}
