package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRejectionError(t *testing.T) {
	for _, tt := range []struct {
		name   string
		reason Reason
		kind   error
	}{
		{name: "invalid token", reason: ReasonInvalidToken, kind: ErrInvalidToken},
		{name: "expired token", reason: ReasonTokenExpired, kind: ErrInvalidToken},
		{name: "missing subject", reason: ReasonMissingSubject, kind: ErrInvalidToken},
		{name: "invalid token type", reason: ReasonInvalidTokenType, kind: ErrInvalidToken},
		{name: "missing local user", reason: ReasonUserNotFound, kind: ErrForbidden},
		{name: "disabled local user", reason: ReasonUserDisabled, kind: ErrForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("authenticate: %w", &RejectionError{Reason: tt.reason})
			assert.ErrorIs(t, err, tt.kind)
			var rejection *RejectionError
			require.ErrorAs(t, err, &rejection)
			assert.Equal(t, tt.reason, rejection.Reason)
			assert.Equal(t, tt.kind.Error(), rejection.Error())
		})
	}
}

func TestReasonString(t *testing.T) {
	for reason, want := range map[Reason]string{
		ReasonInvalidToken:     "invalid_token",
		ReasonTokenExpired:     "token_expired",
		ReasonMissingSubject:   "missing_subject",
		ReasonInvalidTokenType: "invalid_token_type",
		ReasonUserNotFound:     "user_not_found",
		ReasonUserDisabled:     "user_disabled",
		Reason(255):            "invalid_token",
	} {
		assert.Equal(t, want, reason.String())
	}
}
