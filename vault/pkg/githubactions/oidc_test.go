package githubactions

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerifierAcceptsValidGitHubOIDCToken(t *testing.T) {
	key := generateTestKey(t)
	now := time.Unix(1893456000, 0)
	server := newOIDCTestServer(t, key)
	defer server.Close()

	verifier := NewVerifier(server.URL, "vault-test", server.Client())
	verifier.now = func() time.Time { return now }

	token := signTestToken(t, key, map[string]any{
		"iss":        server.URL,
		"sub":        "repo:Gaucho-Racing/Mapache:ref:refs/heads/main",
		"aud":        "vault-test",
		"repository": "Gaucho-Racing/Mapache",
		"ref":        "refs/heads/main",
		"workflow":   "deploy",
		"actor":      "bk1031",
		"run_id":     "123",
		"exp":        now.Add(5 * time.Minute).Unix(),
		"nbf":        now.Add(-time.Minute).Unix(),
		"iat":        now.Unix(),
	})

	claims, err := verifier.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if claims.Repository != "gaucho-racing/mapache" {
		t.Fatalf("repository = %q, want normalized repository", claims.Repository)
	}
	if claims.Ref != "refs/heads/main" {
		t.Fatalf("ref = %q", claims.Ref)
	}
}

func TestVerifierRejectsWrongAudience(t *testing.T) {
	key := generateTestKey(t)
	now := time.Unix(1893456000, 0)
	server := newOIDCTestServer(t, key)
	defer server.Close()

	verifier := NewVerifier(server.URL, "vault-test", server.Client())
	verifier.now = func() time.Time { return now }

	token := signTestToken(t, key, map[string]any{
		"iss":        server.URL,
		"sub":        "repo:Gaucho-Racing/Mapache:ref:refs/heads/main",
		"aud":        "other-audience",
		"repository": "Gaucho-Racing/Mapache",
		"ref":        "refs/heads/main",
		"exp":        now.Add(5 * time.Minute).Unix(),
	})

	if _, err := verifier.Verify(context.Background(), token); err == nil {
		t.Fatal("Verify succeeded with wrong audience")
	}
}

func TestVerifierRejectsExpiredToken(t *testing.T) {
	key := generateTestKey(t)
	now := time.Unix(1893456000, 0)
	server := newOIDCTestServer(t, key)
	defer server.Close()

	verifier := NewVerifier(server.URL, "vault-test", server.Client())
	verifier.now = func() time.Time { return now }

	token := signTestToken(t, key, map[string]any{
		"iss":        server.URL,
		"sub":        "repo:Gaucho-Racing/Mapache:ref:refs/heads/main",
		"aud":        "vault-test",
		"repository": "Gaucho-Racing/Mapache",
		"ref":        "refs/heads/main",
		"exp":        now.Add(-2 * time.Minute).Unix(),
	})

	if _, err := verifier.Verify(context.Background(), token); err == nil {
		t.Fatal("Verify succeeded with expired token")
	}
}

func newOIDCTestServer(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	var server *httptest.Server
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{"jwks_uri": server.URL + "/keys"})
	})
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{"keys": []any{publicJWK(key)}})
	})
	server = httptest.NewServer(mux)
	return server
}

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}
	return key
}

func signTestToken(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()

	header := map[string]any{"alg": "RS256", "kid": "test-key"}
	signingInput := encodeJWTJSON(t, header) + "." + encodeJWTJSON(t, claims)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("SignPKCS1v15 returned error: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func publicJWK(key *rsa.PrivateKey) map[string]string {
	exponent := big.NewInt(int64(key.PublicKey.E)).Bytes()
	return map[string]string{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": "test-key",
		"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(exponent),
	}
}

func encodeJWTJSON(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
}
