package api

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

// Crypto provides cryptographic operations
type Crypto struct{}

// NewCrypto creates a new crypto API
func NewCrypto() *Crypto {
	return &Crypto{}
}

// MD5 computes MD5 hash
func (c *Crypto) MD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA1 computes SHA1 hash
func (c *Crypto) SHA1(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA256 computes SHA256 hash
func (c *Crypto) SHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SHA512 computes SHA512 hash
func (c *Crypto) SHA512(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

// RandomBytes generates random bytes
func (c *Crypto) RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// RandomHex generates a random hex string
func (c *Crypto) RandomHex(n int) (string, error) {
	bytes, err := c.RandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RandomUUID generates a random UUID (v4)
func (c *Crypto) RandomUUID() (string, error) {
	bytes, err := c.RandomBytes(16)
	if err != nil {
		return "", err
	}
	
	// Set version (4) and variant bits
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant 10
	
	// Format as UUID
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}

