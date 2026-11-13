package guild

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Guild struct {
	ID          int64
	Name        string
	DisplayName string
	Description string
	audit.DisableEntry
}

func New(name, displayName, createdBy string) (*Guild, error) {
	g := &Guild{
		Name:        name,
		DisplayName: displayName,
		Enabled:     true,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
	}

	if err := g.validate(); err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Guild) Rename(displayName, changedBy string) error {
	if err := validateDisplayName(displayName); err != nil {
		return err
	}
	if changedBy == "" {
		return errors.New("by is required")
	}
	g.DisplayName = displayName
	g.Update(changedBy)
	return nil
}

func (g *Guild) validate() error {
	if g.Name == "" {
		return errors.New("name is required")
	}
	if err := validateDisplayName(g.DisplayName); err != nil {
		return err
	}
	if len(g.Name) > 70 {
		return fmt.Errorf("%w: name exceeds 70 characters", ErrInvalidGuild)
	}
	for _, r := range g.Name {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			continue
		}
		return fmt.Errorf("%w: name contains invalid characters", ErrInvalidGuild)
	}
	if g.CreatedBy == "" {
		return errors.New("createdBy is required")
	}
	return nil
}

func validateDisplayName(displayName string) error {
	if strings.TrimSpace(displayName) == "" || utf8.RuneCountInString(displayName) > 70 {
		return fmt.Errorf("%w: display name must contain 1 to 70 characters", ErrInvalidGuild)
	}
	return nil
}
