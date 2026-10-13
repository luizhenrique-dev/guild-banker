package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaimsHasRole(t *testing.T) {
	for _, tt := range []struct {
		name  string
		roles []string
		role  string
		want  bool
	}{
		{name: "member", roles: []string{"member"}, role: "member", want: true},
		{name: "admin", roles: []string{"admin"}, role: "admin", want: true},
		{name: "no hierarchy", roles: []string{"admin"}, role: "member"},
		{name: "unknown", roles: []string{"member"}, role: "owner"},
		{name: "case sensitive", roles: []string{"admin"}, role: "Admin"},
		{name: "absent", role: "member"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, (Claims{RealmRoles: tt.roles}).HasRole(tt.role))
		})
	}
}
