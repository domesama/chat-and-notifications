package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeStreamID(t *testing.T) {
	t.Run(
		"same result for reversed user IDs", func(t *testing.T) {
			streamID1 := ComputeStreamID("user1", "user2")
			streamID2 := ComputeStreamID("user2", "user1")

			assert.Equal(t, streamID1, streamID2, "Stream ID should be the same regardless of user order")
		},
	)

	t.Run(
		"different user pairs produce different stream IDs", func(t *testing.T) {
			streamID1 := ComputeStreamID("user1", "user2")
			streamID2 := ComputeStreamID("user1", "user3")

			assert.NotEqual(t, streamID1, streamID2, "Different user pairs should produce different stream IDs")
		},
	)

	t.Run(
		"same user pair always produces same stream ID", func(t *testing.T) {
			streamID1 := ComputeStreamID("alice", "bob")
			streamID2 := ComputeStreamID("alice", "bob")
			streamID3 := ComputeStreamID("bob", "alice")

			assert.Equal(t, streamID1, streamID2, "Same user pair should always produce the same stream ID")
			assert.Equal(t, streamID1, streamID3, "Order should not matter for same user pair")
		},
	)

	t.Run(
		"output length is 16 characters", func(t *testing.T) {
			streamID := ComputeStreamID("user1", "user2")

			assert.Len(t, streamID, 16, "Stream ID should be 16 characters long")
		},
	)

	t.Run(
		"output is hexadecimal", func(t *testing.T) {
			streamID := ComputeStreamID("user1", "user2")

			// Check if string contains only hex characters
			for _, char := range streamID {
				assert.True(
					t, (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'),
					"Stream ID should contain only hexadecimal characters",
				)
			}
		},
	)

	t.Run(
		"handles empty user IDs", func(t *testing.T) {
			streamID1 := ComputeStreamID("", "user1")
			streamID2 := ComputeStreamID("user1", "")

			assert.Equal(t, streamID1, streamID2, "Should handle empty user IDs consistently")
			assert.NotEmpty(t, streamID1, "Should still produce a stream ID")
		},
	)

	t.Run(
		"handles identical user IDs", func(t *testing.T) {
			streamID := ComputeStreamID("user1", "user1")

			assert.NotEmpty(t, streamID, "Should handle identical user IDs")
			assert.Len(t, streamID, 16, "Should produce proper length stream ID")
		},
	)
}
