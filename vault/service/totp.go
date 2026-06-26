package service

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const SecretTypeTOTPSeed = "totp_seed"

var ErrSecretNotTOTP = errors.New("secret is not a TOTP seed")
var ErrTOTPSeedRequired = errors.New("TOTP seed is required")
var ErrTOTPSeedInvalid = errors.New("TOTP seed is invalid")
var ErrTOTPQRCodeNotFound = errors.New("no QR code found")
var ErrTOTPRegistrationInvalid = errors.New("QR code is not a valid TOTP registration")

type TOTPCode struct {
	Code             string    `json:"code"`
	Period           uint      `json:"period"`
	Digits           int       `json:"digits"`
	Algorithm        string    `json:"algorithm"`
	SecondsRemaining int       `json:"seconds_remaining"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type TOTPRegistration struct {
	Value          string `json:"value"`
	Issuer         string `json:"issuer"`
	AccountName    string `json:"account_name"`
	SuggestedKey   string `json:"suggested_key"`
	SuggestedLabel string `json:"suggested_label"`
}

type totpSeedConfig struct {
	Secret    string
	Period    uint
	Digits    otp.Digits
	Algorithm otp.Algorithm
	Encoder   otp.Encoder
}

func IsTOTPSecret(secret model.Secret) bool {
	return strings.EqualFold(strings.TrimSpace(secret.Type), SecretTypeTOTPSeed)
}

func DecodeTOTPRegistrationQRCode(reader io.Reader) (TOTPRegistration, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return TOTPRegistration{}, fmt.Errorf("%w: %v", ErrTOTPQRCodeNotFound, err)
	}
	bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return TOTPRegistration{}, fmt.Errorf("%w: %v", ErrTOTPQRCodeNotFound, err)
	}
	result, err := qrcode.NewQRCodeReader().Decode(bitmap, map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER:    true,
		gozxing.DecodeHintType_ALSO_INVERTED: true,
	})
	if err != nil {
		return TOTPRegistration{}, fmt.Errorf("%w: %v", ErrTOTPQRCodeNotFound, err)
	}
	return ParseTOTPRegistrationURL(result.GetText())
}

func ParseTOTPRegistrationURL(value string) (TOTPRegistration, error) {
	rawValue := strings.TrimSpace(value)
	key, err := parseTOTPKeyFromURL(rawValue)
	if err != nil {
		return TOTPRegistration{}, err
	}
	if strings.TrimSpace(key.Secret()) == "" {
		return TOTPRegistration{}, ErrTOTPSeedRequired
	}
	issuer := strings.TrimSpace(key.Issuer())
	accountName := strings.TrimSpace(key.AccountName())
	return TOTPRegistration{
		Value:          rawValue,
		Issuer:         issuer,
		AccountName:    accountName,
		SuggestedKey:   "totp_" + slugifyTOTPRegistrationKey(firstNonEmpty(accountName, issuer, "code")),
		SuggestedLabel: strings.Join(nonEmptyStrings(issuer, accountName), " "),
	}, nil
}

func GenerateTOTPCode(secret model.Secret, now time.Time) (TOTPCode, error) {
	if !IsTOTPSecret(secret) {
		return TOTPCode{}, ErrSecretNotTOTP
	}
	value, err := RevealSecret(secret)
	if err != nil {
		return TOTPCode{}, err
	}
	config, err := parseTOTPSeedConfig(value)
	if err != nil {
		return TOTPCode{}, err
	}

	code, err := totp.GenerateCodeCustom(config.Secret, now.UTC(), totp.ValidateOpts{
		Period:    config.Period,
		Digits:    config.Digits,
		Algorithm: config.Algorithm,
		Encoder:   config.Encoder,
	})
	if err != nil {
		return TOTPCode{}, fmt.Errorf("%w: %v", ErrTOTPSeedInvalid, err)
	}

	expiresAt := nextTOTPExpiration(now, config.Period)
	return TOTPCode{
		Code:             code,
		Period:           config.Period,
		Digits:           config.Digits.Length(),
		Algorithm:        config.Algorithm.String(),
		SecondsRemaining: int(expiresAt.Unix() - now.UTC().Unix()),
		ExpiresAt:        expiresAt,
	}, nil
}

func parseTOTPSeedConfig(value string) (totpSeedConfig, error) {
	seed := strings.TrimSpace(value)
	if seed == "" {
		return totpSeedConfig{}, ErrTOTPSeedRequired
	}
	if strings.HasPrefix(strings.ToLower(seed), "otpauth://") {
		key, err := parseTOTPKeyFromURL(seed)
		if err != nil {
			return totpSeedConfig{}, err
		}
		if strings.TrimSpace(key.Secret()) == "" {
			return totpSeedConfig{}, ErrTOTPSeedRequired
		}
		period := uint(key.Period())
		if period == 0 {
			period = 30
		}
		return totpSeedConfig{
			Secret:    key.Secret(),
			Period:    period,
			Digits:    key.Digits(),
			Algorithm: key.Algorithm(),
			Encoder:   key.Encoder(),
		}, nil
	}
	return totpSeedConfig{
		Secret:    strings.Join(strings.Fields(seed), ""),
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
		Encoder:   otp.EncoderDefault,
	}, nil
}

func parseTOTPKeyFromURL(value string) (*otp.Key, error) {
	rawValue := strings.TrimSpace(value)
	parsedURL, err := url.Parse(rawValue)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTOTPRegistrationInvalid, err)
	}
	if parsedURL.Scheme != "otpauth" || parsedURL.Host != "totp" {
		return nil, ErrSecretNotTOTP
	}
	key, err := otp.NewKeyFromURL(rawValue)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTOTPRegistrationInvalid, err)
	}
	if key.Type() != "totp" {
		return nil, ErrSecretNotTOTP
	}
	return key, nil
}

func slugifyTOTPRegistrationKey(value string) string {
	parts := []rune{}
	lastWasUnderscore := false
	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			parts = append(parts, char)
			lastWasUnderscore = false
			continue
		}
		if !lastWasUnderscore {
			parts = append(parts, '_')
			lastWasUnderscore = true
		}
	}
	slug := strings.Trim(string(parts), "_")
	if slug == "" {
		return "code"
	}
	return slug
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func nextTOTPExpiration(now time.Time, period uint) time.Time {
	if period == 0 {
		period = 30
	}
	now = now.UTC()
	periodSeconds := int64(period)
	return time.Unix(((now.Unix()/periodSeconds)+1)*periodSeconds, 0).UTC()
}
