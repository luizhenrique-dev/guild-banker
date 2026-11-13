package guild

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"

	"errors"
	"fmt"
	"golang.org/x/text/unicode/norm"
	"time"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
)

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Create(
	ctx context.Context,
	displayName, description string,
	actor audit.Actor,
) (*Guild, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("create guild: %w", err)
	}
	displayName = strings.TrimSpace(displayName)
	base := generatedName(displayName)
	g, err := New(base, displayName, actor.Email)
	if err != nil {
		return nil, fmt.Errorf("create guild: %w", err)
	}
	g.Description = description
	for attempt := range 6 {
		if attempt > 0 {
			suffix := make([]byte, 4)
			if _, err := rand.Read(suffix); err != nil {
				return nil, fmt.Errorf("generate guild name: %w", err)
			}
			prefix := base[:min(len(base), 61)]
			g.Name = strings.TrimRight(prefix, "-") + "-" + hex.EncodeToString(suffix)
		}
		err = s.storage.Create(ctx, g, actor.UserID)
		if err == nil {
			return g, nil
		}
		if !errors.Is(err, ErrGuildNameAlreadyUsed) {
			return nil, fmt.Errorf("create guild: %w", err)
		}
	}
	return nil, ErrNameGenerationUnavailable
}

type UpdateInput struct {
	DisplayName string
	Description *string
}

func (s *Service) Update(ctx context.Context, guildID int64, input UpdateInput, actor audit.Actor) (*Guild, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("update guild name: %w", err)
	}
	if guildID == 0 {
		return nil, errors.New("update guild name: guildID is required")
	}

	isMember, err := s.storage.IsMember(ctx, guildID, actor.UserID)
	if err != nil {
		return nil, fmt.Errorf("check requester membership in update guild name: %w", err)
	}
	if !isMember {
		return nil, ErrRequesterIsNotMember
	}

	g, err := s.storage.GetByID(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("get guild by id in update guild name: %w", err)
	}

	if err := g.Rename(strings.TrimSpace(input.DisplayName), actor.Email); err != nil {
		return nil, fmt.Errorf("rename guild: %w", err)
	}

	if input.Description != nil {
		g.Description = *input.Description
	}

	if err := s.storage.Update(ctx, g); err != nil {
		return nil, fmt.Errorf("persist update guild name: %w", err)
	}

	return g, nil
}

func (s *Service) Enable(ctx context.Context, guildID int64, actor audit.Actor) error {
	if err := actor.Validate(); err != nil {
		return fmt.Errorf("enable guild: %w", err)
	}
	if err := s.ensureRequesterMember(ctx, actor.UserID, guildID, "enable guild"); err != nil {
		return err
	}
	g, err := s.storage.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("enable guild: get guild: %w", err)
	}
	if g.Enabled {
		return ErrGuildAlreadyEnabled
	}
	if err = s.storage.Enable(ctx, guildID, actor.Email, time.Now()); err != nil {
		return fmt.Errorf("enable guild: %w", err)
	}
	return nil
}

func (s *Service) Disable(ctx context.Context, guildID int64, actor audit.Actor) error {
	if err := actor.Validate(); err != nil {
		return fmt.Errorf("disable guild: %w", err)
	}
	if err := s.ensureRequesterMember(ctx, actor.UserID, guildID, "disable guild"); err != nil {
		return err
	}
	g, err := s.storage.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("disable guild: get guild: %w", err)
	}
	if !g.Enabled {
		return ErrGuildAlreadyDisabled
	}
	if err := s.storage.Disable(ctx, guildID, actor.Email, time.Now()); err != nil {
		return fmt.Errorf("disable guild: %w", err)
	}
	return nil
}

func (s *Service) ListByMember(ctx context.Context, userID int64) ([]*Guild, error) {
	if userID == 0 {
		return nil, errors.New("list guilds by member: userID is required")
	}
	guilds, err := s.storage.ListByMember(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list guilds by member: %w", err)
	}
	return guilds, nil
}

func (s *Service) InviteUser(ctx context.Context, requesterUserID, guildID int64, email string) error {
	if err := s.ensureRequesterMember(ctx, requesterUserID, guildID, "invite user"); err != nil {
		return err
	}
	email, err := invitationEmail(email)
	if err != nil {
		return err
	}
	if err := s.storage.InviteByEmail(ctx, guildID, email, requesterUserID); err != nil {
		return fmt.Errorf("invite user: %w", err)
	}
	return nil
}

func (s *Service) RemoveUser(ctx context.Context, requesterUserID, guildID, userID int64) error {
	if err := s.ensureRequesterMember(ctx, requesterUserID, guildID, "remove user"); err != nil {
		return err
	}
	if userID == 0 {
		return errors.New("remove user: userID is required")
	}
	if userID == requesterUserID {
		return errors.New("remove user: cannot remove yourself")
	}
	if err := s.storage.RemoveMember(ctx, guildID, userID); err != nil {
		return fmt.Errorf("remove user: %w", err)
	}
	return nil
}

func (s *Service) ensureRequesterMember(ctx context.Context, requesterUserID, guildID int64, action string) error {
	if requesterUserID == 0 {
		return fmt.Errorf("%s: requesterUserID is required", action)
	}
	if guildID == 0 {
		return fmt.Errorf("%s: guildID is required", action)
	}
	isMember, err := s.storage.IsMember(ctx, guildID, requesterUserID)
	if err != nil {
		return fmt.Errorf("%s: check requester membership: %w", action, err)
	}
	if !isMember {
		return ErrRequesterIsNotMember
	}
	return nil
}

func generatedName(displayName string) string {
	var name strings.Builder
	separator := false
	for _, r := range norm.NFD.String(strings.ToLower(displayName)) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsSpace(r) || r == '-' {
			separator = name.Len() > 0
			continue
		}
		if r < 'a' || r > 'z' {
			if r < '0' || r > '9' {
				continue
			}
		}
		if separator {
			name.WriteByte('-')
			separator = false
		}
		name.WriteRune(r)
	}
	if name.Len() == 0 {
		return "guild"
	}
	return name.String()
}
