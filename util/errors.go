package util

import "errors"

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidKeySize  = errors.New("invalid key size: must be exactly 32 bytes")
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("token has expired")
)
