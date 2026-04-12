package session_test

import (
	"testing"
	"time"

	"github.com/studiolambda/cosmos/framework/session"

	"github.com/stretchr/testify/require"
)

func TestNewSessionReturnsNonNilSession(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.NotNil(t, sess)
}

func TestNewSessionGeneratesSessionID(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.Len(t, sess.SessionID(), 43)
}

func TestNewSessionSetsOriginalIDSameAsSessionID(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.Equal(t, sess.SessionID(), sess.OriginalSessionID())
}

func TestNewSessionIsMarkedAsChanged(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.True(t, sess.HasChanged())
}

func TestNewSessionSetsCreatedAt(t *testing.T) {
	t.Parallel()

	before := time.Now()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.False(t, sess.CreatedAt().Before(before))
	require.False(t, sess.CreatedAt().After(time.Now()))
}

func TestNewSessionSetsExpiresAt(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(2 * time.Hour)

	sess, err := session.NewSession(expiresAt, map[string]any{})

	require.NoError(t, err)
	require.Equal(t, expiresAt, sess.ExpiresAt())
}

func TestNewSessionWithInitialStorage(t *testing.T) {
	t.Parallel()

	storage := map[string]any{"user_id": 42}

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		storage,
	)

	require.NoError(t, err)

	val, ok := sess.Get("user_id")

	require.True(t, ok)
	require.Equal(t, 42, val)
}

func TestNewSessionGeneratesUniqueIDs(t *testing.T) {
	t.Parallel()

	sess1, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)
	require.NoError(t, err)

	sess2, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)
	require.NoError(t, err)

	require.NotEqual(t, sess1.SessionID(), sess2.SessionID())
}

func TestSessionGetReturnsStoredValue(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"key": "value"},
	)

	require.NoError(t, err)

	val, ok := sess.Get("key")

	require.True(t, ok)
	require.Equal(t, "value", val)
}

func TestSessionGetReturnsFalseForMissingKey(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	_, ok := sess.Get("missing")

	require.False(t, ok)
}

func TestSessionPutStoresValue(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sess.MarkAsUnchanged()
	sess.Put("key", "value")

	val, ok := sess.Get("key")

	require.True(t, ok)
	require.Equal(t, "value", val)
}

func TestSessionPutMarksAsChanged(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sess.MarkAsUnchanged()
	sess.Put("key", "value")

	require.True(t, sess.HasChanged())
}

func TestSessionPutOverwritesExistingValue(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"key": "old"},
	)

	require.NoError(t, err)

	sess.Put("key", "new")

	val, ok := sess.Get("key")

	require.True(t, ok)
	require.Equal(t, "new", val)
}

func TestSessionDeleteRemovesKey(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"key": "value"},
	)

	require.NoError(t, err)

	sess.Delete("key")

	_, ok := sess.Get("key")

	require.False(t, ok)
}

func TestSessionDeleteMarksAsChanged(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"key": "value"},
	)

	require.NoError(t, err)

	sess.MarkAsUnchanged()
	sess.Delete("key")

	require.True(t, sess.HasChanged())
}

func TestSessionDeleteMissingKeyIsNoOp(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sess.MarkAsUnchanged()
	sess.Delete("nonexistent")

	require.True(t, sess.HasChanged())
}

func TestSessionClearRemovesAllData(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"a": 1, "b": 2, "c": 3},
	)

	require.NoError(t, err)

	sess.Clear()

	_, okA := sess.Get("a")
	_, okB := sess.Get("b")
	_, okC := sess.Get("c")

	require.False(t, okA)
	require.False(t, okB)
	require.False(t, okC)
}

func TestSessionClearMarksAsChanged(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"key": "value"},
	)

	require.NoError(t, err)

	sess.MarkAsUnchanged()
	sess.Clear()

	require.True(t, sess.HasChanged())
}

func TestSessionExtendUpdatesExpiresAt(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(1*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	newExpiry := time.Now().Add(5 * time.Hour)
	sess.Extend(newExpiry)

	require.Equal(t, newExpiry, sess.ExpiresAt())
}

func TestSessionExtendMarksAsChanged(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(1*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sess.MarkAsUnchanged()
	sess.Extend(time.Now().Add(5 * time.Hour))

	require.True(t, sess.HasChanged())
}

func TestSessionRegenerateChangesSessionID(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	originalID := sess.SessionID()

	err = sess.Regenerate()

	require.NoError(t, err)
	require.NotEqual(t, originalID, sess.SessionID())
}

func TestSessionRegeneratePreservesOriginalID(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	originalID := sess.OriginalSessionID()

	err = sess.Regenerate()

	require.NoError(t, err)
	require.Equal(t, originalID, sess.OriginalSessionID())
}

func TestSessionRegenerateMarksAsChanged(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	sess.MarkAsUnchanged()

	err = sess.Regenerate()

	require.NoError(t, err)
	require.True(t, sess.HasChanged())
}

func TestSessionHasExpiredWhenPastExpiresAt(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(-1*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.True(t, sess.HasExpired())
}

func TestSessionHasNotExpiredWhenBeforeExpiresAt(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.False(t, sess.HasExpired())
}

func TestSessionExpiresSoonWhenWithinDelta(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(5*time.Minute),
		map[string]any{},
	)

	require.NoError(t, err)
	require.True(t, sess.ExpiresSoon(10*time.Minute))
}

func TestSessionExpiresSoonReturnsFalseWhenFarFromExpiry(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.False(t, sess.ExpiresSoon(10*time.Minute))
}

func TestSessionExpiresSoonReturnsFalseWhenAlreadyExpired(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(-1*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.False(t, sess.ExpiresSoon(10*time.Minute))
}

func TestSessionHasRegeneratedAfterRegenerate(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.False(t, sess.HasRegenerated())

	err = sess.Regenerate()

	require.NoError(t, err)
	require.True(t, sess.HasRegenerated())
}

func TestSessionMarkAsUnchangedResetsChangedFlag(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)
	require.True(t, sess.HasChanged())

	sess.MarkAsUnchanged()

	require.False(t, sess.HasChanged())
}

func TestSessionRegenerateGeneratesValidID(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{},
	)

	require.NoError(t, err)

	err = sess.Regenerate()

	require.NoError(t, err)
	require.Len(t, sess.SessionID(), 43)
}

func TestSessionPreservesDataAfterRegenerate(t *testing.T) {
	t.Parallel()

	sess, err := session.NewSession(
		time.Now().Add(2*time.Hour),
		map[string]any{"user_id": 42},
	)

	require.NoError(t, err)

	err = sess.Regenerate()

	require.NoError(t, err)

	val, ok := sess.Get("user_id")

	require.True(t, ok)
	require.Equal(t, 42, val)
}
