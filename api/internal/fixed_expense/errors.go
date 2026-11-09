package fixedexpense

import "errors"

var (
	ErrInvalidInput              = errors.New("invalid fixed expense input")
	ErrRequesterIsNotMember      = errors.New("requester is not a member of this guild")
	ErrFixedExpenseNotFound      = errors.New("fixed expense not found")
	ErrInvalidFixedExpenseStatus = errors.New("invalid fixed expense status")
)
