package guild

import (
	"fmt"
	"net/mail"
	"strings"
)

type Invitation struct {
	ID             int64
	Guild          Guild
	InvitedByName  string
	InvitedByEmail string
}

func invitationEmail(email string) (string, error) {
	address, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", fmt.Errorf("%w: invalid email", ErrInvalidInvitation)
	}
	return strings.ToLower(address.Address), nil
}
