package helpers

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

func GenerateUniqueID() uint {
	u := uuid.New()
	hash := sha256.Sum256(u[:])
	return uint(binary.BigEndian.Uint64(hash[:8]))
}
