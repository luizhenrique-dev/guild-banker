//go:build integration && postgres

package guild

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/luizhenrique-dev/guild-banker/api/internal/audit"
	"github.com/luizhenrique-dev/guild-banker/api/internal/auth"
	"github.com/luizhenrique-dev/guild-banker/api/internal/user"
)

func TestGuildMembersAndCancellationPostgres(t *testing.T) {
	db := guildTestDB(t)
	accounts := user.NewService(user.NewStorage(db))
	owner, err := accounts.EnsureAccount(t.Context(), "owner", "Owner Full Name", "owner@example.test")
	require.NoError(t, err)
	member, err := accounts.EnsureAccount(t.Context(), "member", "Member Full Name", "member@example.test")
	require.NoError(t, err)
	recipient, err := accounts.EnsureAccount(t.Context(), "recipient", "Recipient", "recipient@example.test")
	require.NoError(t, err)
	service := NewService(NewStorage(db))
	actor := audit.Actor{UserID: owner.ID, Email: owner.Email}
	g, err := service.Create(t.Context(), "Members", "", actor)
	require.NoError(t, err)
	otherGuild, err := service.Create(t.Context(), "Other guild", "", actor)
	require.NoError(t, err)
	handler := NewHandler(service)
	router := gin.New()
	router.GET("/guilds/:guildID/members", handler.ListMembers)
	router.GET("/guilds/:guildID/invites", handler.ListSentInvitations)
	router.DELETE("/guilds/:guildID/invites/:inviteID", handler.CancelInvitation)
	request := func(method, path string, who *user.User, status int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, nil)
		if who != nil {
			req = req.WithContext(auth.WithPrincipal(req.Context(), auth.Principal{Actor: audit.Actor{UserID: who.ID, Email: who.Email}}))
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, status, response.Code, response.Body.String())
		return response
	}
	membersPath := fmt.Sprintf("/guilds/%d/members", g.ID)
	invitesPath := fmt.Sprintf("/guilds/%d/invites", g.ID)
	request("GET", membersPath, nil, 401)
	request("GET", "/guilds/-1/members", owner, 400)
	request("GET", membersPath, recipient, 403)
	request("GET", invitesPath, recipient, 403)
	require.JSONEq(t, "[]", request("GET", invitesPath, owner, 200).Body.String())
	var members []memberResponse
	require.NoError(t, json.Unmarshal(request("GET", membersPath, owner, 200).Body.Bytes(), &members))
	require.Len(t, members, 1)
	require.Equal(t, owner.ID, members[0].ID)
	require.Equal(t, owner.Name, members[0].Name)
	require.Equal(t, owner.Email, members[0].Email)
	require.Nil(t, members[0].LastAccessAt)
	require.False(t, members[0].MemberSince.IsZero())
	require.NoError(t, accounts.RecordAccess(t.Context(), owner.ID))
	var before time.Time
	require.NoError(t, db.Get(&before, "SELECT last_access_at FROM user_account WHERE id=$1", owner.ID))
	require.NoError(t, accounts.RecordAccess(t.Context(), owner.ID))
	require.NoError(t, json.Unmarshal(request("GET", membersPath, owner, 200).Body.Bytes(), &members))
	require.NotNil(t, members[0].LastAccessAt)
	require.False(t, members[0].LastAccessAt.Before(before))
	// A different member can cancel the owner's invitations.
	require.NoError(t, service.InviteUser(t.Context(), owner.ID, g.ID, member.Email))
	pending, err := service.ListSentInvitations(t.Context(), owner.ID, g.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	_, err = service.RespondInvitation(t.Context(), pending[0].ID, audit.Actor{UserID: member.ID, Email: member.Email}, true)
	require.NoError(t, err)
	require.NoError(t, service.InviteUser(t.Context(), owner.ID, g.ID, recipient.Email))
	var sent []sentInvitationResponse
	require.NoError(t, json.Unmarshal(request("GET", invitesPath, owner, 200).Body.Bytes(), &sent))
	require.Len(t, sent, 1)
	require.Equal(t, recipient.Email, sent[0].Email)
	require.Equal(t, owner.Name, sent[0].InvitedBy.Name)
	require.Equal(t, owner.Email, sent[0].InvitedBy.Email)
	require.False(t, sent[0].CreatedAt.IsZero())
	id := sent[0].ID
	cancelPath := fmt.Sprintf("%s/%d", invitesPath, id)
	request("DELETE", cancelPath, recipient, 403)
	request("DELETE", fmt.Sprintf("/guilds/%d/invites/%d", otherGuild.ID, id), owner, 404)
	request("DELETE", cancelPath, member, 204)
	request("DELETE", cancelPath, member, 409)
	require.JSONEq(t, "[]", request("GET", invitesPath, owner, 200).Body.String())
	var cancelled struct {
		Status      string     `db:"status"`
		RespondedAt *time.Time `db:"responded_at"`
		RespondedBy int64      `db:"responded_by"`
	}
	require.NoError(t, db.Get(&cancelled, "SELECT status,responded_at,responded_by FROM guild_invitation WHERE id=$1", id))
	require.Equal(t, "CANCELLED", cancelled.Status)
	require.NotNil(t, cancelled.RespondedAt)
	require.Equal(t, member.ID, cancelled.RespondedBy)
	_, err = service.RespondInvitation(t.Context(), id, audit.Actor{UserID: recipient.ID, Email: recipient.Email}, true)
	require.ErrorIs(t, err, ErrInvitationAlreadyAnswered)
	require.NoError(t, service.InviteUser(t.Context(), owner.ID, g.ID, recipient.Email))
	pending, err = service.ListSentInvitations(t.Context(), owner.ID, g.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.NotEqual(t, id, pending[0].ID)
	// The membership starts at acceptance, rather than the invitation's send date.
	_, err = db.Exec("UPDATE guild_invitation SET created_at=NOW()-INTERVAL '10 days' WHERE id=$1", pending[0].ID)
	require.NoError(t, err)
	_, err = service.RespondInvitation(t.Context(), pending[0].ID, audit.Actor{UserID: recipient.ID, Email: recipient.Email}, true)
	require.NoError(t, err)
	request("DELETE", fmt.Sprintf("%s/%d", invitesPath, pending[0].ID), owner, 409)
	require.NoError(t, json.Unmarshal(request("GET", membersPath, owner, 200).Body.Bytes(), &members))
	require.Len(t, members, 3)
	require.WithinDuration(t, time.Now(), members[2].MemberSince, time.Minute)
}

func TestInvitationCancelVersusAcceptPostgres(t *testing.T) {
	db := guildTestDB(t)
	accounts := user.NewService(user.NewStorage(db))
	owner, err := accounts.EnsureAccount(t.Context(), "owner", "Owner", "owner@example.test")
	require.NoError(t, err)
	recipient, err := accounts.EnsureAccount(t.Context(), "recipient", "Recipient", "recipient@example.test")
	require.NoError(t, err)
	service := NewService(NewStorage(db))
	for attempt := range 5 {
		g, err := service.Create(t.Context(), fmt.Sprintf("Race %d", attempt), "", audit.Actor{UserID: owner.ID, Email: owner.Email})
		require.NoError(t, err)
		require.NoError(t, service.InviteUser(t.Context(), owner.ID, g.ID, recipient.Email))
		pending, err := service.ListSentInvitations(t.Context(), owner.ID, g.ID)
		require.NoError(t, err)
		require.Len(t, pending, 1)
		var acceptedErr, cancelledErr error
		start := make(chan struct{})
		var workers sync.WaitGroup
		workers.Go(func() {
			<-start
			_, acceptedErr = service.RespondInvitation(t.Context(), pending[0].ID, audit.Actor{UserID: recipient.ID, Email: recipient.Email}, true)
		})
		workers.Go(func() {
			<-start
			cancelledErr = service.CancelInvitation(t.Context(), owner.ID, g.ID, pending[0].ID)
		})
		close(start)
		workers.Wait()
		isMember, err := NewStorage(db).IsMember(t.Context(), g.ID, recipient.ID)
		require.NoError(t, err)
		if acceptedErr == nil {
			require.ErrorIs(t, cancelledErr, ErrInvitationAlreadyAnswered)
			require.True(t, isMember)
		} else {
			require.NoError(t, cancelledErr)
			require.ErrorIs(t, acceptedErr, ErrInvitationAlreadyAnswered)
			require.False(t, isMember)
		}
	}
}
