package nats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertSubjectHashToAngleBracket(t *testing.T) {
	t.Parallel()

	result := convertSubject("user.#")

	require.Equal(t, "user.>", result)
}

func TestConvertSubjectNoWildcardsUnchanged(t *testing.T) {
	t.Parallel()

	result := convertSubject("user.created")

	require.Equal(t, "user.created", result)
}

func TestConvertSubjectStarUnchanged(t *testing.T) {
	t.Parallel()

	result := convertSubject("user.*.created")

	require.Equal(t, "user.*.created", result)
}

func TestConvertSubjectHashOnly(t *testing.T) {
	t.Parallel()

	result := convertSubject("#")

	require.Equal(t, ">", result)
}

func TestNATSSubjectsHashIncludesExactSubject(t *testing.T) {
	t.Parallel()

	result := natsSubjects("logs.#")

	require.Equal(t, []string{"logs", "logs.>"}, result)
}

func TestNATSSubjectsHashOnlyUsesMultiTokenSubject(t *testing.T) {
	t.Parallel()

	result := natsSubjects("#")

	require.Equal(t, []string{">"}, result)
}

func TestConvertSubjectMultipleHashes(t *testing.T) {
	t.Parallel()

	result := convertSubject("a.#.b.#")

	require.Equal(t, "a.>.b.>", result)
}
