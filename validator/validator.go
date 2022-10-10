package validator

import (
	"context"
)

type validatorFunc func(ctx context.Context) error

type validatorType int

const (
	ConcurrentValidator validatorType = iota
	SerializedValidator
)

type Validator interface {
	Validate() error
	AddValidator(validatorName string, fn validatorFunc)
}

func NewValidator(ctx context.Context, validatorType ...validatorType) Validator {
	types := validatorType
	if len(types) <= 0 {
		return newConcurrentValidator(ctx)
	}

	switch types[0] {
	case ConcurrentValidator:
		return newConcurrentValidator(ctx)
	case SerializedValidator:
		return newSerializedValidator(ctx)
	default:
		return newConcurrentValidator(ctx)
	}
}
