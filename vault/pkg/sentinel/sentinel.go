package sentinel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gaucho-racing/vault/vault/config"
)

type Error struct {
	Code    int
	Message string `json:"error"`
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

func ValidateToken(token string) (map[string]interface{}, error) {
	if strings.TrimSpace(config.SentinelURL) == "" {
		return nil, fmt.Errorf("SENTINEL_URL is not configured")
	}

	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(config.SentinelURL, "/")+"/api/core/token/validate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		var sentinelErr Error
		if err := json.Unmarshal(respBody, &sentinelErr); err != nil {
			return nil, err
		}
		sentinelErr.Code = resp.StatusCode
		return nil, fmt.Errorf("sentinel error: [%d] %s", sentinelErr.Code, sentinelErr.Message)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(respBody, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}
