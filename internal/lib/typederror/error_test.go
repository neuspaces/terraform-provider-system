package typederror_test

import (
	"errors"
	"github.com/neuspaces/terraform-provider-system/internal/lib/typederror"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorType(t *testing.T) {
	var ErrRoot = typederror.NewRoot("root error")

	var ErrNotExists = typederror.New("does not exist", ErrRoot)
	var ErrInvalidResult = typederror.New("invalid result", ErrRoot)

	var ErrOperation = typederror.New("operation failed", ErrRoot)

	var ErrOperationUnexpected = typederror.New("unexpected outcome", ErrOperation)

	t.Run("error type identity", func(t *testing.T) {
		t.Parallel()

		assert.True(t, ErrRoot.Is(ErrRoot), "ErrRoot.Is(ErrRoot) should be true")
		assert.False(t, ErrNotExists.Is(ErrRoot), "ErrNotExists.Is(ErrRoot) should be false")
		assert.False(t, ErrRoot.Is(ErrNotExists), "ErrRoot.Is(ErrNotExists) should be false")
	})

	t.Run("error type hierarchy", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, ErrRoot.Unwrap(), "ErrRoot.Unwrap() == nil")

		if errRoot := ErrInvalidResult.Unwrap().(*typederror.ErrorType); assert.NotNil(t, errRoot, "ErrInvalidResult should unwrap an ErrorType") {
			assert.True(t, errRoot.Is(ErrRoot), "ErrInvalidResult should unwrap to ErrRoot")
		}

		if errOperation := ErrOperationUnexpected.Unwrap().(*typederror.ErrorType); assert.NotNil(t, errOperation, "ErrOperationUnexpected should unwrap an ErrOperation") {
			assert.True(t, errOperation.Is(ErrOperation), "ErrOperationUnexpected should unwrap to ErrOperation")
		}
	})

	t.Run("error type string", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "root error", ErrRoot.Error())
		assert.Equal(t, "root error: does not exist", ErrNotExists.Error())
		assert.Equal(t, "root error: invalid result", ErrInvalidResult.Error())
		assert.Equal(t, "root error: operation failed", ErrOperation.Error())
		assert.Equal(t, "root error: operation failed: unexpected outcome", ErrOperationUnexpected.Error())
	})

	t.Run("raise specific error from type", func(t *testing.T) {
		t.Parallel()

		specificErr := ErrInvalidResult.Raise(errors.New("numbers do not match"))

		assert.Equal(t, "root error: invalid result: numbers do not match", specificErr.Error())
		assert.True(t, errors.Is(specificErr, ErrInvalidResult), "errors.Is(specificErr, ErrInvalidResult) == true")
		assert.True(t, errors.Is(specificErr, ErrRoot), "errors.Is(specificErr, ErrRoot) == true")
		assert.False(t, errors.Is(specificErr, ErrNotExists), "errors.Is(specificErr, ErrInvalidResult) == false")
	})
}
