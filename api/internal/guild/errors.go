package guild

import "errors"

var (
	ErrGuildNotFound         = errors.New("guild not found")
	ErrGuildNameAlreadyUsed  = errors.New("guild name already used")
	ErrRequesterIsNotMember  = errors.New("requester is not a guild member")
	ErrUserNotFound          = errors.New("user not found")
	ErrGuildMemberNotFound   = errors.New("guild member not found")
	ErrGuildMemberAlreadySet = errors.New("user is already a guild member")
	ErrGuildAlreadyEnabled   = errors.New("guild already enabled")
	ErrGuildAlreadyDisabled  = errors.New("guild already disabled")
)
