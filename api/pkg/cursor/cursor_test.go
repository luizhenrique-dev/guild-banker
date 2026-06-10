package cursor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, time.June, 9, 20, 0, 0, 0, time.UTC)

	encoded, err := Encode(occurredAt, 42)
	require.NoError(t, err)

	decodedOccurredAt, decodedID, err := Decode(encoded)
	require.NoError(t, err)

	require.True(t, decodedOccurredAt.Equal(occurredAt))
	require.Equal(t, int64(42), decodedID)
}

func TestDecodeInvalidCursor(t *testing.T) {
	t.Parallel()

	_, _, err := Decode("invalido")
	require.ErrorIs(t, err, ErrInvalid)
}
