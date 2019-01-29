package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

func CalculateHash(input []byte) string {
	checksum := sha256.Sum256(input)
	return hex.EncodeToString(checksum[:])
}
