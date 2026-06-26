package service

import (
	"strings"
	"testing"
	"time"

	"github.com/gaucho-racing/vault/vault/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRecentViewAuditLogQueryScopesSecretViewsBySecretID(t *testing.T) {
	db := newDryRunDB(t)
	var count int64

	statement := recentViewAuditLogQuery(db, model.AuditLog{
		Action:        AuditActionSecretViewed,
		AccountID:     "acct_123",
		ActorEntityID: "ent_123",
		SecretID:      "sec_123",
	}, time.Unix(1893456000, 0)).Count(&count).Statement

	assertAuditDebounceQuery(t, statement, "sec_123")
}

func TestRecentViewAuditLogQueryScopesAccountViewsToEmptySecretID(t *testing.T) {
	db := newDryRunDB(t)
	var count int64

	statement := recentViewAuditLogQuery(db, model.AuditLog{
		Action:        AuditActionAccountViewed,
		AccountID:     "acct_123",
		ActorEntityID: "ent_123",
	}, time.Unix(1893456000, 0)).Count(&count).Statement

	assertAuditDebounceQuery(t, statement, "")
}

func assertAuditDebounceQuery(t *testing.T, statement *gorm.Statement, expectedSecretID string) {
	t.Helper()

	query := statement.SQL.String()
	for _, column := range []string{"action", "account_id", "actor_entity_id", "created_at", "secret_id"} {
		if !strings.Contains(query, column) {
			t.Fatalf("query does not include %s scope: %s", column, query)
		}
	}
	if len(statement.Vars) != 5 {
		t.Fatalf("vars = %v, want 5 vars", statement.Vars)
	}
	if actualSecretID := statement.Vars[4]; actualSecretID != expectedSecretID {
		t.Fatalf("secret id var = %v, want %q", actualSecretID, expectedSecretID)
	}
}

func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=localhost user=postgres password=postgres dbname=vault port=5432 sslmode=disable",
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("gorm.Open returned error: %v", err)
	}
	return db
}
