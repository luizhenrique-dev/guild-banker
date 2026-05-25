package guild

type createGuildRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type updateGuildRequest struct {
	Name string `json:"name"`
}

type inviteUserRequest struct {
	Email string `json:"email"`
}

type guildResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}
