package chatstream

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// ComputeStreamID generates a deterministic stream ID from two user IDs.
// The result is the same regardless of the order of userID1 and userID2.
// This ensures both participants in a conversation always compute the same stream_id.
func ComputeStreamID(userID1, userID2 string) string {
	// Sort user IDs to ensure deterministic output
	ids := []string{userID1, userID2}
	sort.Strings(ids)

	combined := ids[0] + ids[1]
	hash := sha256.Sum256([]byte(combined))

	// Return first 16 characters of hex-encoded hash
	return hex.EncodeToString(hash[:])[:16]
}
