package utils

import (
	"crypto/rand"
	"encoding/hex"
)

const defaultIDLength = 12
const defaultTokenLength = 32

func GenerateFileID() string {
	return GenerateID(defaultIDLength)
}

func GenerateFolderID() string {
	return GenerateID(defaultIDLength)
}

func GenerateID(length int) string {
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)[:length]
}

func GenerateToken() string {
	bytes := make([]byte, defaultTokenLength)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GenerateShareToken() string {
	return GenerateToken()
}
