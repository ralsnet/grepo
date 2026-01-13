package grepo

import (
	"fmt"
)

var (
	ErrNotFound = fmt.Errorf("NotFound")
	ErrInvalid  = fmt.Errorf("Invalid")
)
