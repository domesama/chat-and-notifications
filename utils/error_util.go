package utils

import (
	"fmt"
)

func WrapError(errorMessage error, errorType error) error {
	errStr := errorMessage.Error()
	return fmt.Errorf("%w: %s", errorType, errStr)
}
