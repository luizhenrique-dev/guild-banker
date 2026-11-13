package guild

import (
	"fmt"

	"github.com/luizhenrique-dev/guild-banker/api/internal/emailaddress"
)

type Invitation struct {
	ID             int64
	Guild          Guild
	InvitedByName  string
	InvitedByEmail string
}

func invitationEmail(email string) (string, error) {
	address, err := emailaddress.Normalize(email)
	if err != nil {
		return "", fmt.Errorf("%w: invalid email", ErrInvalidInvitation)
	}
	return address, nil
}
