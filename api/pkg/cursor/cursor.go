package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalid = errors.New("invalid cursor")

type payload struct {
	OccurredAt time.Time `json:"occurredAt"`
	ID         int64     `json:"id"`
}

func Encode(occurredAt time.Time, id int64) (string, error) {
	b, err := json.Marshal(payload{OccurredAt: occurredAt, ID: id})
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func Decode(value string) (time.Time, int64, error) {
	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return time.Time{}, 0, ErrInvalid
	}

	var parsed payload
	if err := json.Unmarshal(b, &parsed); err != nil {
		return time.Time{}, 0, ErrInvalid
	}

	if parsed.OccurredAt.IsZero() || parsed.ID == 0 {
		return time.Time{}, 0, ErrInvalid
	}

	return parsed.OccurredAt, parsed.ID, nil
}
