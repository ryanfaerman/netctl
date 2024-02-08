package finders

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidField      = errors.New("invalid field")
	ErrInvalidFieldType  = errors.New("invalid field type")
	ErrInvalidFieldValue = errors.New("invalid field value")
	ErrMissingWhere      = errors.New("at least one where type is required")
	ErrInvalidWhere      = errors.New("invalid where field")
)
