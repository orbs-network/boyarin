package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
)

func CalculateHash(input []byte) string {
	checksum := sha256.Sum256(input)
	return hex.EncodeToString(checksum[:])
}

func CalculateFileHash(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return CalculateHash(data), nil
}
