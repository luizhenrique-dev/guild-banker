package emailaddress

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	for _, tt := range []struct {
		name, raw, want string
	}{
		{name: "plain", raw: "member@example.test", want: "member@example.test"},
		{name: "spaces and case", raw: " \tMEMBER@Example.test \n", want: "member@example.test"},
		{name: "alias and subdomain", raw: "MEMBER.Name+bank@My-Group.Example.test", want: "member.name+bank@my-group.example.test"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.raw)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			again, err := Normalize(got)
			require.NoError(t, err)
			require.Equal(t, got, again, "normalization must be idempotent")
		})
	}
	for _, raw := range []string{
		"private invalid address", `"MEMBER"@example.test`, `Member <member@example.test>`,
		`"A B"@example.test`, `"A@B"@example.test`, `"A\"B"@example.test`,
		"member..name@example.test", ".member@example.test", "member.@example.test",
		"member@example", "member@-example.test", "member@example-.test", "member@exam_ple.test",
	} {
		t.Run("rejects "+raw, func(t *testing.T) {
			_, err := Normalize(raw)
			require.ErrorIs(t, err, ErrInvalid)
			require.NotContains(t, err.Error(), raw)
		})
	}
}
