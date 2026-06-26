package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/gaucho-racing/vault/vault/pkg/githubactions"
)

func TestNormalizeApplicationGitHubActions(t *testing.T) {
	application := model.Application{
		Name: "Mapache-Prod",
		GitHubActionsRepositories: []string{
			" Gaucho-Racing/Mapache ",
			"gaucho-racing/mapache",
		},
		GitHubActionsRefs: []string{"refs/heads/main", " refs/tags/v* "},
	}

	if err := normalizeApplication(&application); err != nil {
		t.Fatalf("normalizeApplication returned error: %v", err)
	}
	if application.Name != "mapache-prod" {
		t.Fatalf("name = %q", application.Name)
	}
	if got := strings.Join(application.GitHubActionsRepositories, ","); got != "gaucho-racing/mapache" {
		t.Fatalf("repositories = %q", got)
	}
	if got := strings.Join(application.GitHubActionsRefs, ","); got != "refs/heads/main,refs/tags/v*" {
		t.Fatalf("refs = %q", got)
	}
}

func TestNormalizeApplicationRequiresRefsWhenGitHubRepositoryConfigured(t *testing.T) {
	application := model.Application{
		Name:                      "mapache",
		GitHubActionsRepositories: []string{"gaucho-racing/mapache"},
	}

	err := normalizeApplication(&application)
	if !errors.Is(err, ErrGitHubActionsRefRequired) {
		t.Fatalf("error = %v, want ErrGitHubActionsRefRequired", err)
	}
}

func TestNormalizeApplicationRejectsInvalidGitHubRef(t *testing.T) {
	application := model.Application{
		Name:                      "mapache",
		GitHubActionsRepositories: []string{"gaucho-racing/mapache"},
		GitHubActionsRefs:         []string{"main"},
	}

	err := normalizeApplication(&application)
	if !errors.Is(err, ErrGitHubActionsRefInvalid) {
		t.Fatalf("error = %v, want ErrGitHubActionsRefInvalid", err)
	}
}

func TestAuthorizeGitHubActionsApplicationAccess(t *testing.T) {
	application := model.Application{
		GitHubActionsRepositories: []string{"gaucho-racing/mapache"},
		GitHubActionsRefs:         []string{"refs/heads/main", "refs/tags/v*"},
	}

	err := authorizeGitHubActionsApplicationAccess(application, githubactions.Claims{
		Repository: "Gaucho-Racing/Mapache",
		Ref:        "refs/tags/v1.2.3",
	})
	if err != nil {
		t.Fatalf("authorizeGitHubActionsApplicationAccess returned error: %v", err)
	}
}

func TestAuthorizeGitHubActionsApplicationAccessRejectsRef(t *testing.T) {
	application := model.Application{
		GitHubActionsRepositories: []string{"gaucho-racing/mapache"},
		GitHubActionsRefs:         []string{"refs/heads/main"},
	}

	err := authorizeGitHubActionsApplicationAccess(application, githubactions.Claims{
		Repository: "gaucho-racing/mapache",
		Ref:        "refs/heads/feature",
	})
	if !errors.Is(err, ErrGitHubActionsRefNotAllowed) {
		t.Fatalf("error = %v, want ErrGitHubActionsRefNotAllowed", err)
	}
}

func TestGitHubActionsEnvName(t *testing.T) {
	tests := map[string]string{
		"api-key":      "API_KEY",
		"api_key":      "API_KEY",
		"1password":    "_1PASSWORD",
		"client.token": "CLIENT_TOKEN",
	}

	for key, expected := range tests {
		if actual := githubActionsEnvName(key); actual != expected {
			t.Fatalf("githubActionsEnvName(%q) = %q, want %q", key, actual, expected)
		}
	}
}

func TestFormatGitHubActionsEnvAssignment(t *testing.T) {
	assignment := formatGitHubActionsEnvAssignment("API_KEY", "line one\nline two")
	if !strings.HasPrefix(assignment, "API_KEY<<VAULT_") {
		t.Fatalf("assignment prefix = %q", assignment)
	}
	if !strings.Contains(assignment, "\nline one\nline two\n") {
		t.Fatalf("assignment does not include exact multiline value: %q", assignment)
	}
}

func TestWildcardMatchesAcrossSlashBoundaries(t *testing.T) {
	if !wildcardMatches("refs/heads/*", "refs/heads/feature/github-actions") {
		t.Fatal("wildcard did not match branch path")
	}
	if wildcardMatches("refs/tags/v*", "refs/heads/main") {
		t.Fatal("wildcard matched the wrong namespace")
	}
}
