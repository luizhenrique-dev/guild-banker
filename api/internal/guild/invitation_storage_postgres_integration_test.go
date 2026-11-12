//go:build integration && postgres

package guild

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestInvitationResponseRollbackPostgres(t *testing.T) {
	db := guildTestDB(t)
	var inviterID int64
	require.NoError(t, db.Get(&inviterID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('inviter','Inviter','inviter@example.test','fixture') RETURNING id`))
	service := NewService(NewStorage(db))
	g, err := service.Create(t.Context(), "Family", "Description", audit.Actor{UserID: inviterID, Email: "inviter@example.test"})
	require.NoError(t, err)
	require.NoError(t, service.InviteUser(t.Context(), inviterID, g.ID, "new@example.test"))
	var id int64
	require.NoError(t, db.Get(&id, "SELECT id FROM guild_invitation"))
	responder, ok := NewStorage(db).(interface {
		RespondInvitation(context.Context, int64, audit.Actor, bool) (*Guild, error)
	})
	require.True(t, ok, "storage must support atomic invitation responses")
	_, err = responder.RespondInvitation(t.Context(), id, audit.Actor{UserID: inviterID + 100, Email: "new@example.test"}, true)
	require.Error(t, err)
	var status string
	require.NoError(t, db.Get(&status, "SELECT status FROM guild_invitation WHERE id=$1", id))
	require.Equal(t, "PENDING", status)
}

func TestInvitationLifecyclePostgres(t *testing.T) {
	db := guildTestDB(t)
	var ownerID, recipientID int64
	require.NoError(t, db.Get(&ownerID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('owner','Inviter','owner@example.test','fixture') RETURNING id`))
	service := NewService(NewStorage(db))
	owner := audit.Actor{UserID: ownerID, Email: "owner@example.test"}
	g, err := service.Create(t.Context(), "Family", "Shared finances", owner)
	require.NoError(t, err)
	require.NoError(t, service.InviteUser(t.Context(), ownerID, g.ID, "Recipient@example.test"))
	// The account is provisioned after the invitation already exists.
	require.NoError(t, db.Get(&recipientID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('recipient','Recipient','"recipient"@example.test','fixture') RETURNING id`))
	recipient := audit.Actor{UserID: recipientID, Email: `"RECIPIENT"@example.test`}
	invitations, err := service.ListInvitations(t.Context(), recipient)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
	require.Equal(t, "Inviter", invitations[0].InvitedByName)
	require.Equal(t, owner.Email, invitations[0].InvitedByEmail)
	require.Equal(t, "Shared finances", invitations[0].Guild.Description)
	id := invitations[0].ID
	_, err = service.RespondInvitation(t.Context(), id, owner, true)
	require.ErrorIs(t, err, ErrInvitationNotFound)
	_, err = service.RespondInvitation(t.Context(), id, recipient, false)
	require.NoError(t, err)
	member, err := NewStorage(db).IsMember(t.Context(), g.ID, recipientID)
	require.NoError(t, err)
	require.False(t, member)
	_, err = service.RespondInvitation(t.Context(), id, recipient, true)
	require.ErrorIs(t, err, ErrInvitationAlreadyAnswered)
	require.NoError(t, service.InviteUser(t.Context(), ownerID, g.ID, "recipient@example.test"))
	invitations, err = service.ListInvitations(t.Context(), recipient)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
	joined, err := service.RespondInvitation(t.Context(), invitations[0].ID, recipient, true)
	require.NoError(t, err)
	require.Equal(t, g.ID, joined.ID)
	member, err = NewStorage(db).IsMember(t.Context(), g.ID, recipientID)
	require.NoError(t, err)
	require.True(t, member)
	invitations, err = service.ListInvitations(t.Context(), recipient)
	require.NoError(t, err)
	require.Empty(t, invitations)
	require.ErrorIs(t, service.InviteUser(t.Context(), ownerID, g.ID, "recipient@example.test"), ErrGuildMemberAlreadySet)
}

func TestInvitationDecisionsRacePostgres(t *testing.T) {
	for _, tt := range []struct {
		name   string
		second bool
	}{{name: "two accepts", second: true}, {name: "accept versus decline", second: false}} {
		t.Run(tt.name, func(t *testing.T) {
			db := guildTestDB(t)
			var ownerID, recipientID int64
			require.NoError(t, db.Get(&ownerID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('owner','Owner','owner@example.test','fixture') RETURNING id`))
			require.NoError(t, db.Get(&recipientID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('recipient','Recipient','"recipient"@example.test','fixture') RETURNING id`))
			service := NewService(NewStorage(db))
			owner := audit.Actor{UserID: ownerID, Email: "owner@example.test"}
			recipient := audit.Actor{UserID: recipientID, Email: "recipient@example.test"}
			g, err := service.Create(t.Context(), "Concurrent", "", owner)
			require.NoError(t, err)
			require.NoError(t, service.InviteUser(t.Context(), ownerID, g.ID, recipient.Email))
			pending, err := service.ListInvitations(t.Context(), recipient)
			require.NoError(t, err)
			results := make(chan error, 2)
			var wg sync.WaitGroup
			for _, accept := range []bool{true, tt.second} {
				wg.Go(func() {
					_, err := service.RespondInvitation(t.Context(), pending[0].ID, recipient, accept)
					results <- err
				})
			}
			wg.Wait()
			close(results)
			successes := 0
			for err := range results {
				if err == nil {
					successes++
					continue
				}
				require.ErrorIs(t, err, ErrInvitationAlreadyAnswered)
			}
			require.Equal(t, 1, successes)
			var status string
			require.NoError(t, db.Get(&status, "SELECT status FROM guild_invitation WHERE id=$1", pending[0].ID))
			member, err := NewStorage(db).IsMember(t.Context(), g.ID, recipientID)
			require.NoError(t, err)
			require.Equal(t, status == "ACCEPTED", member)
		})
	}
}

func TestInvitationEmissionVersusAcceptancePostgres(t *testing.T) {
	db := guildTestDB(t)
	var ownerID, recipientID int64
	require.NoError(t, db.Get(&ownerID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('owner','Owner','owner@example.test','fixture') RETURNING id`))
	require.NoError(t, db.Get(&recipientID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('recipient','Recipient','"recipient"@example.test','fixture') RETURNING id`))
	service := NewService(NewStorage(db))
	owner := audit.Actor{UserID: ownerID, Email: "owner@example.test"}
	recipient := audit.Actor{UserID: recipientID, Email: "recipient@example.test"}
	g, err := service.Create(t.Context(), "Race", "", owner)
	require.NoError(t, err)
	require.NoError(t, service.InviteUser(t.Context(), ownerID, g.ID, recipient.Email))
	pending, err := service.ListInvitations(t.Context(), recipient)
	require.NoError(t, err)
	var wg sync.WaitGroup
	accepted := make(chan error, 1)
	emitted := make(chan error, 1)
	wg.Go(func() {
		_, err := service.RespondInvitation(t.Context(), pending[0].ID, recipient, true)
		accepted <- err
	})
	wg.Go(func() { emitted <- service.InviteUser(t.Context(), ownerID, g.ID, recipient.Email) })
	wg.Wait()
	require.NoError(t, <-accepted)
	err = <-emitted
	require.True(t, errors.Is(err, ErrInvitationAlreadyPending) || errors.Is(err, ErrGuildMemberAlreadySet))
	pending, err = service.ListInvitations(t.Context(), recipient)
	require.NoError(t, err)
	require.Empty(t, pending)
}

func TestInvitationHTTPPostgres(t *testing.T) {
	db := guildTestDB(t)
	var ownerID, recipientID int64
	require.NoError(t, db.Get(&ownerID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('owner','Owner','owner@example.test','fixture') RETURNING id`))
	require.NoError(t, db.Get(&recipientID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('recipient','Recipient','"recipient"@example.test','fixture') RETURNING id`))
	service := NewService(NewStorage(db))
	owner := audit.Actor{UserID: ownerID, Email: "owner@example.test"}
	recipient := audit.Actor{UserID: recipientID, Email: "recipient@example.test"}
	g, err := service.Create(t.Context(), "Family", "Description", owner)
	require.NoError(t, err)
	require.NoError(t, service.InviteUser(t.Context(), ownerID, g.ID, recipient.Email))
	router := gin.New()
	handler := NewHandler(service)
	router.GET("/guild-invites", handler.ListInvitations)
	router.POST("/guild-invites/:inviteID/accept", handler.AcceptInvitation)
	router.POST("/guild-invites/:inviteID/decline", handler.DeclineInvitation)
	request := func(method, path string, actor audit.Actor, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(`{"email":"recipient@example.test","user_id":999}`))
		req.Header.Set("X-User-Email", "recipient@example.test")
		if actor.UserID != 0 {
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: actor}))
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		require.Equal(t, status, recorder.Code, recorder.Body.String())
		return recorder.Body.Bytes()
	}
	request("GET", "/guild-invites", audit.Actor{}, 401)
	var invitations []invitationResponse
	require.NoError(t, json.Unmarshal(request("GET", "/guild-invites", recipient, 200), &invitations))
	require.Len(t, invitations, 1)
	require.Equal(t, "Owner", invitations[0].InvitedBy.Name)
	require.Equal(t, "Description", invitations[0].Guild.Description)
	path := fmt.Sprintf("/guild-invites/%d/accept", invitations[0].ID)
	request("POST", path, owner, 404)
	var joined guildResponse
	require.NoError(t, json.Unmarshal(request("POST", path, recipient, 200), &joined))
	require.Equal(t, g.ID, joined.ID)
	require.Contains(t, string(request("POST", path, recipient, 409)), "invitation_already_answered")
	require.JSONEq(t, "[]", string(request("GET", "/guild-invites", recipient, 200)))
}

func TestAcceptancePreservesOtherInvitationsPostgres(t *testing.T) {
	db := guildTestDB(t)
	var ownerID, recipientID int64
	require.NoError(t, db.Get(&ownerID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('owner','Owner','owner@example.test','fixture') RETURNING id`))
	require.NoError(t, db.Get(&recipientID, `INSERT INTO user_account(external_id,name,email,created_by) VALUES ('recipient','Recipient','recipient@example.test','fixture') RETURNING id`))
	service := NewService(NewStorage(db))
	owner := audit.Actor{UserID: ownerID, Email: "owner@example.test"}
	recipient := audit.Actor{UserID: recipientID, Email: "recipient@example.test"}
	for _, name := range []string{"First", "Second"} {
		g, err := service.Create(t.Context(), name, "", owner)
		require.NoError(t, err)
		require.NoError(t, service.InviteUser(t.Context(), ownerID, g.ID, recipient.Email))
	}
	pending, err := service.ListInvitations(t.Context(), recipient)
	require.NoError(t, err)
	require.Len(t, pending, 2)
	_, err = service.RespondInvitation(t.Context(), pending[0].ID, recipient, true)
	require.NoError(t, err)
	remaining, err := service.ListInvitations(t.Context(), recipient)
	require.NoError(t, err)
	require.Len(t, remaining, 1)
	require.Equal(t, pending[1].ID, remaining[0].ID)
	var status string
	require.NoError(t, db.Get(&status, "SELECT status FROM guild_invitation WHERE id=$1", pending[1].ID))
	require.Equal(t, "PENDING", status)
}
