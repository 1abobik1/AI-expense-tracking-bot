package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

const TokenTTL = 7 * 24 * time.Hour

// Токен — срок и HMAC от него; ключ из пароля, поэтому смена пароля завершает все входы.
type AuthUseCase struct {
	password []byte
	key      []byte
	now      func() time.Time
}

func NewAuthUC(password string) *AuthUseCase {
	key := sha256.Sum256([]byte("expensebot-admin-token:" + password))
	return &AuthUseCase{password: []byte(password), key: key[:], now: time.Now}
}

func (a *AuthUseCase) CheckPassword(given string) bool {
	want, got := sha256.Sum256(a.password), sha256.Sum256([]byte(given))
	return subtle.ConstantTimeCompare(want[:], got[:]) == 1
}

func (a *AuthUseCase) IssueToken() (token string, expires time.Time) {
	expires = a.now().Add(TokenTTL).Truncate(time.Second)
	exp := strconv.FormatInt(expires.Unix(), 10)

	return exp + "." + a.sign(exp), expires
}

func (a *AuthUseCase) VerifyToken(token string) bool {
	exp, sig, ok := strings.Cut(token, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(a.sign(exp))) {
		return false
	}

	unix, err := strconv.ParseInt(exp, 10, 64)

	return err == nil && a.now().Unix() < unix
}

func (a *AuthUseCase) sign(exp string) string {
	mac := hmac.New(sha256.New, a.key)
	mac.Write([]byte("admin:" + exp))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
