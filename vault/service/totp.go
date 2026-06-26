package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gaucho-racing/vault/vault/model"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const SecretTypeTOTPSeed = "totp_seed"

var ErrSecretNotTOTP = errors.New("secret is not a TOTP seed")
var ErrTOTPSeedRequired = errors.New("TOTP seed is required")
var ErrTOTPSeedInvalid = errors.New("TOTP seed is invalid")

type TOTPCode struct {
	Code             string    `json:"code"`
	Period           uint      `json:"period"`
	Digits           int       `json:"digits"`
	Algorithm        string    `json:"algorithm"`
	SecondsRemaining int       `json:"seconds_remaining"`
	ExpiresAt        time.Time `json:"expires_at"`
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
		key, err := otp.NewKeyFromURL(seed)
		if err != nil {
			return totpSeedConfig{}, fmt.Errorf("%w: %v", ErrTOTPSeedInvalid, err)
		}
		if key.Type() != "totp" {
			return totpSeedConfig{}, ErrSecretNotTOTP
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

func nextTOTPExpiration(now time.Time, period uint) time.Time {
	if period == 0 {
		period = 30
	}
	now = now.UTC()
	periodSeconds := int64(period)
	return time.Unix(((now.Unix()/periodSeconds)+1)*periodSeconds, 0).UTC()
}
