package validator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func Test_concurrentValidatorImpl_1(t *testing.T) {
	validator := NewValidator(context.Background())

	validator.AddValidator("Validator.TestOne", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	validator.AddValidator("Validator.TestTwo", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return errors.New("error")
	})

	validator.AddValidator("Validator.TestThree", func(ctx context.Context) error {
		time.Sleep(150 * time.Millisecond)
		return errors.New("error")
	})

	err := validator.Validate()
	reflect.DeepEqual(errors.New("Validator.TestTwo failed due to error\n; Validator.TestThree failed due to error\n"), err)
}

func Test_concurrentValidatorImpl_2(t *testing.T) {
	validator := NewValidator(context.Background(), ConcurrentValidator)

	var wg sync.WaitGroup
	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			validator.AddValidator(fmt.Sprintf("Validator.Test%v", n), func(ctx context.Context) error {
				time.Sleep(100 * time.Millisecond)
				if n == 500 {
					return errors.New("error")
				}
				return nil
			})
		}(i)
	}
	wg.Wait()

	err := validator.Validate()
	reflect.DeepEqual(errors.New("Validator.Test500 failed due to error\n"), err)
}

func Test_serializedValidatorImpl(t *testing.T) {
	validator := NewValidator(context.Background(), SerializedValidator)

	validator.AddValidator("Validator.TestOne", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	validator.AddValidator("Validator.TestTwo", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return errors.New("error")
	})

	validator.AddValidator("Validator.TestThree", func(ctx context.Context) error {
		time.Sleep(150 * time.Millisecond)
		return errors.New("error")
	})

	err := validator.Validate()
	reflect.DeepEqual(errors.New("Validator.TestTwo failed due to error\n"), err)
}
