package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gin-gonic/gin"
)

func TestRequestTokenCanAccessAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		claims  map[string]interface{}
		account model.Account
		want    bool
	}{
		{
			name: "no token",
			account: model.Account{
				AccessGroupNames: []string{"Admins"},
			},
			want: false,
		},
		{
			name: "sentinel all",
			claims: map[string]interface{}{
				"sub":   "ent_1",
				"scope": "sentinel:all",
			},
			account: model.Account{
				AccessGroupNames: []string{"Admins"},
			},
			want: true,
		},
		{
			name: "matching group name",
			claims: map[string]interface{}{
				"sub":    "ent_1",
				"scope":  "groups:read",
				"groups": []interface{}{"SocialManagers"},
			},
			account: model.Account{
				AccessGroupNames: []string{"Admins", "SocialManagers"},
			},
			want: true,
		},
		{
			name: "missing group name",
			claims: map[string]interface{}{
				"sub":    "ent_1",
				"scope":  "groups:read",
				"groups": []interface{}{"Fabrication"},
			},
			account: model.Account{
				AccessGroupNames: []string{"Admins", "SocialManagers"},
			},
			want: false,
		},
		{
			name: "open account",
			claims: map[string]interface{}{
				"sub":    "ent_1",
				"scope":  "groups:read",
				"groups": []interface{}{"Fabrication"},
			},
			account: model.Account{},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.claims != nil {
				setAuthContext(c, "token", tt.claims)
			}

			got := RequestTokenCanAccessAccount(c, tt.account)
			if got != tt.want {
				t.Fatalf("RequestTokenCanAccessAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
