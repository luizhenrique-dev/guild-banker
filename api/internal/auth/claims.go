package auth

import (
	"slices"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Claims struct {
	UserID            string
	PreferredUsername string
	RealmRoles        []string
}

func (c Claims) HasRole(role string) bool {
	return slices.Contains(c.RealmRoles, role)
}

type Principal struct {
	Claims Claims
	Actor  audit.Actor
}
