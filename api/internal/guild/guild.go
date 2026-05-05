package guild

import (
	"errors"
	"time"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Guild struct {
	ID          int64
	Name        string
	DisplayName string
	audit.DisableEntry
}

func New(name, displayName, createdBy string) (*Guild, error) {
	g := &Guild{
		Name:        name,
		DisplayName: displayName,
		DisableEntry: audit.DisableEntry{
			Enabled: true,
			Entry: audit.Entry{
				CreatedAt: time.Now(),
				CreatedBy: createdBy,
			},
		},
	}

	if err := g.validate(); err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Guild) Rename(name, by string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if by == "" {
		return errors.New("by is required")
	}
	g.Name = name
	g.Update(by)
	return nil
}

func (g *Guild) validate() error {
	if g.Name == "" {
		return errors.New("name is required")
	}
	if g.DisplayName == "" {
		return errors.New("displayName is required")
	}
	if g.CreatedBy == "" {
		return errors.New("createdBy is required")
	}
	return nil
}
