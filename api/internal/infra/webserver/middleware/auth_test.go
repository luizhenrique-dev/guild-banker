package middleware

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractBearerToken(t *testing.T) {
	for _, tt := range []struct {
		name, header, want string
		ok                 bool
	}{
		{name: "absent"},
		{name: "empty", header: "Bearer "},
		{name: "wrong scheme", header: "Basic secret"},
		{name: "extra credential", header: "Bearer foo bar"},
		{name: "joined duplicates", header: "Bearer foo,Bearer bar"},
		{name: "tab", header: "Bearer\tfoo"},
		{name: "newline", header: "Bearer foo\n"},
		{name: "valid", header: "Bearer fixture.jwt.token", want: "fixture.jwt.token", ok: true},
		{name: "scheme case", header: "bEaReR token", want: "token", ok: true},
		{name: "multiple spaces", header: "Bearer   token", want: "token", ok: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			raw, ok := extractBearerToken(tt.header)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, raw)
		})
	}
}
