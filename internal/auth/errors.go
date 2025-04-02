package auth

import "errors"

var (
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrUserNotFound             = errors.New("user not found")
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrInvalidToken             = errors.New("invalid token")
	ErrTokenExpired             = errors.New("token expired")
	ErrTokenRevoked             = errors.New("token revoked")
	ErrTooManyRequests          = errors.New("too many requests")
	ErrInvalidRequest           = errors.New("invalid request")
	ErrEmailAlreadyVerified     = errors.New("email already verified")
	ErrVerificationCodeNotFound = errors.New("verification code not found")
	ErrVerificationCodeExpired  = errors.New("verification code expired")
	ErrVerificationCodeInvalid  = errors.New("invalid verification code")
)
