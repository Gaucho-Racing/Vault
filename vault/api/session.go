package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gaucho-racing/vault/vault/config"
	"github.com/gaucho-racing/vault/vault/pkg/logger"
	"github.com/gaucho-racing/vault/vault/pkg/sentinel"
	"github.com/gin-gonic/gin"
)

const oauthScope = "user:read groups:read"

func LoginWithSentinel(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	returnTo := sanitizeReturnTo(c.Query("return_to"))
	redirectURI := callbackURL(c)

	setCookie(c, oauthStateCookieName, state, int((10 * time.Minute).Seconds()))
	setCookie(c, oauthReturnToCookieName, returnTo, int((10 * time.Minute).Seconds()))

	params := url.Values{}
	params.Set("client_id", config.SentinelClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", oauthScope)
	params.Set("state", state)

	c.Redirect(http.StatusFound, strings.TrimRight(config.SentinelURL, "/")+"/oauth/authorize?"+params.Encode())
}

func SentinelCallback(c *gin.Context) {
	if errMessage := c.Query("error"); errMessage != "" {
		clearOAuthCookies(c)
		c.Redirect(http.StatusFound, "/auth/login?error="+url.QueryEscape(errMessage))
		return
	}
	expectedState, err := c.Cookie(oauthStateCookieName)
	if err != nil || expectedState == "" || c.Query("state") != expectedState {
		clearOAuthCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid oauth state"})
		return
	}
	code := c.Query("code")
	if code == "" {
		clearOAuthCookies(c)
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	token, err := sentinel.ExchangeAuthorizationCode(code, callbackURL(c))
	if err != nil {
		clearOAuthCookies(c)
		logger.SugarLogger.Errorln("Failed to exchange Sentinel authorization code: " + err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	returnTo := "/accounts"
	if value, err := c.Cookie(oauthReturnToCookieName); err == nil {
		returnTo = sanitizeReturnTo(value)
	}
	setSessionCookies(c, token)
	clearOAuthCookies(c)
	c.Redirect(http.StatusFound, returnTo)
}

func GetSession(c *gin.Context) {
	Require(c, RequestTokenExists(c))
	c.JSON(http.StatusOK, gin.H{
		"entity_id": GetRequestTokenEntityID(c),
		"user_id":   GetRequestTokenUserID(c),
		"scope":     GetRequestTokenScopes(c),
		"groups":    GetRequestTokenGroupNames(c),
	})
}

func RefreshSession(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil || refreshToken == "" {
		clearSessionCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}
	token, err := sentinel.RefreshToken(refreshToken)
	if err != nil {
		clearSessionCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	setSessionCookies(c, token)
	c.JSON(http.StatusOK, gin.H{"message": "session refreshed"})
}

func Logout(c *gin.Context) {
	clearSessionCookies(c)
	clearOAuthCookies(c)
	c.Status(http.StatusNoContent)
}

func setSessionCookies(c *gin.Context, token sentinel.TokenResponse) {
	maxAge := token.ExpiresIn
	if maxAge <= 0 {
		maxAge = int((15 * time.Minute).Seconds())
	}
	setCookie(c, accessTokenCookieName, token.AccessToken, maxAge)
	if token.RefreshToken != "" {
		setCookie(c, refreshTokenCookieName, token.RefreshToken, int((30 * 24 * time.Hour).Seconds()))
	}
}

func clearSessionCookies(c *gin.Context) {
	clearCookie(c, accessTokenCookieName)
	clearCookie(c, refreshTokenCookieName)
}

func clearOAuthCookies(c *gin.Context) {
	clearCookie(c, oauthStateCookieName)
	clearCookie(c, oauthReturnToCookieName)
}

func setCookie(c *gin.Context, name string, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   config.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookie(c *gin.Context, name string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	})
}

func callbackURL(c *gin.Context) string {
	if config.SentinelRedirectURI != "" {
		return config.SentinelRedirectURI
	}
	return requestBaseURL(c) + "/api/auth/callback"
}

func requestBaseURL(c *gin.Context) string {
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return fmt.Sprintf("%s://%s", proto, host)
}

func randomState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func sanitizeReturnTo(value string) string {
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || strings.HasPrefix(value, "/api/") {
		return "/accounts"
	}
	return value
}
