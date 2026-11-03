package auth

import "errors"

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrForbidden    = errors.New("access denied")
)

type Reason uint8

const (
	ReasonInvalidToken Reason = iota
	ReasonTokenExpired
	ReasonMissingSubject
	ReasonInvalidTokenType
	ReasonUserNotFound
	ReasonUserDisabled
	ReasonIncompleteProfile
	ReasonEmailUnverified
	ReasonEmailConflict
)

func (r Reason) String() string {
	switch r {
	case ReasonTokenExpired:
		return "token_expired"
	case ReasonMissingSubject:
		return "missing_subject"
	case ReasonInvalidTokenType:
		return "invalid_token_type"
	case ReasonUserNotFound:
		return "user_not_found"
	case ReasonIncompleteProfile:
		return "incomplete_profile"
	case ReasonEmailUnverified:
		return "email_unverified"
	case ReasonEmailConflict:
		return "email_conflict"
	case ReasonUserDisabled:
		return "user_disabled"
	default:
		return "invalid_token"
	}
}

type RejectionError struct {
	Reason Reason
}

func (e *RejectionError) Error() string {
	return e.Unwrap().Error()
}

func (e *RejectionError) Unwrap() error {
	switch e.Reason {
	case ReasonUserNotFound, ReasonUserDisabled, ReasonIncompleteProfile, ReasonEmailUnverified, ReasonEmailConflict:
		return ErrForbidden
	}
	return ErrInvalidToken
}
