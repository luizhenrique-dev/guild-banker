package emailaddress

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalid = errors.New("invalid email address")

const (
	maxEmailLength = 254
	emailPattern   = `^[a-z0-9_%+\-]+(?:\.[a-z0-9_%+\-]+)*@(?:[a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`
)

// Normalize validates a simple email address and returns its lowercase form.
func Normalize(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if len(email) > maxEmailLength || !regexp.MustCompile(emailPattern).MatchString(email) {
		return "", ErrInvalid
	}
	return email, nil
}
