package fixedexpense

import "errors"

var (
	ErrFixedExpenseNotFound      = errors.New("fixed expense not found")
	ErrInvalidFixedExpenseStatus = errors.New("invalid fixed expense status")
)
