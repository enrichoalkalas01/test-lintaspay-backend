package entity

import "errors"

var (
	ErrNotFound             = errors.New("record not found")
	ErrIdempotencyKeyExists = errors.New("idempotency key already exists")
)
