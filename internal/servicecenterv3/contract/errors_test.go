package contract

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFoundAndInvalid(t *testing.T) {
	wrapped := fmt.Errorf("wrap: %w", ErrServiceNotFound)
	if !IsNotFound(wrapped) {
		t.Fatal("wrapped SERVICE_NOT_FOUND should match IsNotFound")
	}
	if IsNotFound(ErrInvalidArgument) {
		t.Fatal("INVALID_ARGUMENT is not a not-found")
	}
	if !IsInvalid(errors.Join(ErrInvalidContext)) {
		t.Fatal("ErrInvalidContext should match IsInvalid")
	}
	if IsInvalid(ErrCenterNotFound) {
		t.Fatal("CENTER_NOT_FOUND is not invalid-input")
	}
}
