package auth

import "time"

// AuthUseCase — вход главы семьи в панель по одному паролю.
type AuthUseCase interface {
	CheckPassword(given string) bool
	IssueToken() (token string, expires time.Time)
	VerifyToken(token string) bool
}
