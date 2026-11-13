package guild

import "time"

type Member struct {
	ID           int64
	Name         string
	Email        string
	LastAccessAt *time.Time
	MemberSince  time.Time
}

type SentInvitation struct {
	ID             int64
	Email          string
	InvitedByName  string
	InvitedByEmail string
	CreatedAt      time.Time
}
