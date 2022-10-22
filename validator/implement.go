package validator

import (
	"context"
	"fmt"
	"github.com/Pangjiping/goutils/linked_map"
	"github.com/Pangjiping/goutils/utils"
	"strings"
	"sync"
)

type concurrentValidatorImpl struct {
	validatorMap map[string]validatorFunc
	errChan      chan error
	mu           sync.RWMutex
	ctx          context.Context
}

func newConcurrentValidator(ctx context.Context) Validator {
	return &concurrentValidatorImpl{
		validatorMap: make(map[string]validatorFunc),
		ctx:          ctx,
	}
}

func (validator *concurrentValidatorImpl) Validate() error {
	chanSize := len(validator.validatorMap)
	if chanSize <= 0 {
		return nil
	}

	validator.errChan = make(chan error, chanSize)
	var wg sync.WaitGroup
	for name, fn := range validator.validatorMap {
		wg.Add(1)
		go func(validatorName string, validatorFn validatorFunc) {
			defer utils.Recovery()
			defer wg.Done()

			if err := validatorFn(validator.ctx); err != nil {
				validator.errChan <- fmt.Errorf("%s failed due to %++v\n", validatorName, err)
			}

		}(name, fn)
	}
	wg.Wait()
	close(validator.errChan)

	if len(validator.errChan) > 0 {
		errSlice := make([]string, 0, len(validator.errChan))
		for err := range validator.errChan {
			errSlice = append(errSlice, err.Error())
		}
		return fmt.Errorf("%++v", strings.Join(errSlice, "; "))
	}
	return nil
}

func (validator *concurrentValidatorImpl) AddValidator(validatorName string, fn validatorFunc) {
	validator.mu.Lock()
	defer validator.mu.Unlock()

	validator.validatorMap[validatorName] = fn
}

type serializedValidatorImpl struct {
	validatorMap *linked_map.LinkedMap
	ctx          context.Context
}

func newSerializedValidator(ctx context.Context) Validator {
	return &serializedValidatorImpl{
		validatorMap: linked_map.NewLinkedMap(),
		ctx:          ctx,
	}
}

func (validator *serializedValidatorImpl) AddValidator(validatorName string, fn validatorFunc) {
	validator.validatorMap.Set(validatorName, fn)
}

func (validator *serializedValidatorImpl) Validate() error {
	if validator.validatorMap.Len() <= 0 {
		return nil
	}

	for elem := validator.validatorMap.Front(); elem != nil; elem = elem.Next() {
		validatorName := elem.Key.(string)
		validatorFn := elem.Value.(validatorFunc)
		if err := validatorFn(validator.ctx); err != nil {
			return fmt.Errorf("%s failed due to %++v", validatorName, err)
		}
	}
	return nil
}
