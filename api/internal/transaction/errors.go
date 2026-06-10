package transaction

import "errors"

var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrRequesterIsNotMember   = errors.New("requester is not a member of guild")
	ErrTransactionCancelled   = errors.New("transaction is cancelled")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
)
