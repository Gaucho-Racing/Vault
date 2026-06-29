package kubernetes

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const DefaultAudience = "gaucho-racing-vault"

var ErrInvalidToken = errors.New("invalid kubernetes service account token")
var ErrNotConfigured = errors.New("kubernetes oidc issuer is not configured")

type Claims struct {
	Issuer             string
	Subject            string
	Audience           []string
	Namespace          string
	ServiceAccountName string
	ServiceAccountUID  string
	ExpiresAt          time.Time
	NotBefore          time.Time
	IssuedAt           time.Time
}

type Verifier struct {
	issuer   string
	audience string
	client   *http.Client
	now      func() time.Time

	mu            sync.Mutex
	jwksURI       string
	keys          map[string]*rsa.PublicKey
	keysExpiresAt time.Time
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type tokenClaims struct {
	Issuer     string           `json:"iss"`
	Subject    string           `json:"sub"`
	Audience   audienceClaim    `json:"aud"`
	Kubernetes kubernetesClaims `json:"kubernetes.io"`
	ExpiresAt  int64            `json:"exp"`
	NotBefore  int64            `json:"nbf"`
	IssuedAt   int64            `json:"iat"`
}

type kubernetesClaims struct {
	Namespace      string              `json:"namespace"`
	ServiceAccount serviceAccountClaim `json:"serviceaccount"`
}

type serviceAccountClaim struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type audienceClaim []string

type discoveryDocument struct {
	JWKSURI string `json:"jwks_uri"`
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	N         string `json:"n"`
	E         string `json:"e"`
}

func NewVerifier(issuer string, audience string, client *http.Client) *Verifier {
	if strings.TrimSpace(audience) == "" {
		audience = DefaultAudience
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &Verifier{
		issuer:   strings.TrimRight(strings.TrimSpace(issuer), "/"),
		audience: strings.TrimSpace(audience),
		client:   client,
		now:      time.Now,
	}
}

func (v *Verifier) Verify(ctx context.Context, token string) (Claims, error) {
	if v.issuer == "" {
		return Claims{}, ErrNotConfigured
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return Claims{}, invalidToken("token is required")
	}
	if len(token) > 64*1024 {
		return Claims{}, invalidToken("token is too large")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, invalidToken("token must have three segments")
	}

	var header tokenHeader
	if err := decodeSegment(parts[0], &header); err != nil {
		return Claims{}, invalidToken("decode header: %v", err)
	}
	if header.Algorithm != "RS256" {
		return Claims{}, invalidToken("unsupported signing algorithm")
	}
	if header.KeyID == "" {
		return Claims{}, invalidToken("missing key id")
	}

	key, err := v.publicKey(ctx, header.KeyID)
	if err != nil {
		return Claims{}, err
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Claims{}, invalidToken("decode signature: %v", err)
	}
	digest := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return Claims{}, invalidToken("signature verification failed")
	}

	var rawClaims tokenClaims
	if err := decodeSegment(parts[1], &rawClaims); err != nil {
		return Claims{}, invalidToken("decode claims: %v", err)
	}
	if err := v.validateClaims(rawClaims); err != nil {
		return Claims{}, err
	}

	namespace := rawClaims.Kubernetes.Namespace
	serviceAccountName := rawClaims.Kubernetes.ServiceAccount.Name
	if namespace == "" || serviceAccountName == "" {
		namespace, serviceAccountName = serviceAccountFromSubject(rawClaims.Subject)
	}

	return Claims{
		Issuer:             rawClaims.Issuer,
		Subject:            rawClaims.Subject,
		Audience:           []string(rawClaims.Audience),
		Namespace:          strings.ToLower(namespace),
		ServiceAccountName: strings.ToLower(serviceAccountName),
		ServiceAccountUID:  rawClaims.Kubernetes.ServiceAccount.UID,
		ExpiresAt:          unixTime(rawClaims.ExpiresAt),
		NotBefore:          unixTime(rawClaims.NotBefore),
		IssuedAt:           unixTime(rawClaims.IssuedAt),
	}, nil
}

func (v *Verifier) publicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	if v.keysExpiresAt.After(v.now()) {
		key := v.keys[keyID]
		v.mu.Unlock()
		if key != nil {
			return key, nil
		}
	} else {
		v.mu.Unlock()
	}

	if err := v.refreshKeys(ctx); err != nil {
		return nil, err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	key := v.keys[keyID]
	if key == nil {
		return nil, invalidToken("signing key is not trusted")
	}
	return key, nil
}

func (v *Verifier) refreshKeys(ctx context.Context) error {
	jwksURI, err := v.discoveryJWKSURI(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURI, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kubernetes oidc jwks returned status %d", resp.StatusCode)
	}

	var document jwksDocument
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&document); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, rawKey := range document.Keys {
		key, err := rawKey.publicKey()
		if err != nil || key == nil {
			continue
		}
		keys[rawKey.KeyID] = key
	}
	if len(keys) == 0 {
		return errors.New("kubernetes oidc jwks did not include usable rsa keys")
	}

	v.mu.Lock()
	v.keys = keys
	v.keysExpiresAt = v.now().Add(5 * time.Minute)
	v.mu.Unlock()
	return nil
}

func (v *Verifier) discoveryJWKSURI(ctx context.Context) (string, error) {
	v.mu.Lock()
	jwksURI := v.jwksURI
	v.mu.Unlock()
	if jwksURI != "" {
		return jwksURI, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.issuer+"/.well-known/openid-configuration", nil)
	if err != nil {
		return "", err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kubernetes oidc discovery returned status %d", resp.StatusCode)
	}

	var document discoveryDocument
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&document); err != nil {
		return "", err
	}
	if document.JWKSURI == "" {
		return "", errors.New("kubernetes oidc discovery did not include jwks_uri")
	}

	v.mu.Lock()
	v.jwksURI = document.JWKSURI
	v.mu.Unlock()
	return document.JWKSURI, nil
}

func (v *Verifier) validateClaims(claims tokenClaims) error {
	now := v.now()
	if claims.Issuer != v.issuer {
		return invalidToken("issuer is not trusted")
	}
	if claims.Subject == "" {
		return invalidToken("subject is required")
	}
	if !claims.Audience.contains(v.audience) {
		return invalidToken("audience is not trusted")
	}
	if claims.ExpiresAt == 0 {
		return invalidToken("expiration is required")
	}
	if now.Add(-1*time.Minute).Unix() > claims.ExpiresAt {
		return invalidToken("token is expired")
	}
	if claims.NotBefore != 0 && now.Add(time.Minute).Unix() < claims.NotBefore {
		return invalidToken("token is not valid yet")
	}
	if claims.IssuedAt != 0 && now.Add(time.Minute).Unix() < claims.IssuedAt {
		return invalidToken("token issued-at is in the future")
	}

	namespace := claims.Kubernetes.Namespace
	serviceAccountName := claims.Kubernetes.ServiceAccount.Name
	if namespace == "" || serviceAccountName == "" {
		namespace, serviceAccountName = serviceAccountFromSubject(claims.Subject)
	}
	if namespace == "" {
		return invalidToken("namespace claim is required")
	}
	if serviceAccountName == "" {
		return invalidToken("service account claim is required")
	}
	return nil
}

func serviceAccountFromSubject(subject string) (string, string) {
	parts := strings.Split(subject, ":")
	if len(parts) != 4 || parts[0] != "system" || parts[1] != "serviceaccount" {
		return "", ""
	}
	return parts[2], parts[3]
}

func (a *audienceClaim) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = []string{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

func (a audienceClaim) contains(expected string) bool {
	for _, audience := range a {
		if audience == expected {
			return true
		}
	}
	return false
}

func (k jwk) publicKey() (*rsa.PublicKey, error) {
	if k.KeyID == "" || k.KeyType != "RSA" || k.N == "" || k.E == "" {
		return nil, nil
	}
	if k.Algorithm != "" && k.Algorithm != "RS256" {
		return nil, nil
	}
	if k.Use != "" && k.Use != "sig" {
		return nil, nil
	}

	modulusBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}

	exponent := 0
	for _, value := range exponentBytes {
		exponent = exponent<<8 + int(value)
	}
	if exponent == 0 {
		return nil, errors.New("rsa exponent is empty")
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulusBytes),
		E: exponent,
	}, nil
}

func decodeSegment(segment string, destination any) error {
	data, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, destination)
}

func invalidToken(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidToken, fmt.Sprintf(format, args...))
}

func unixTime(value int64) time.Time {
	if value == 0 {
		return time.Time{}
	}
	return time.Unix(value, 0)
}
