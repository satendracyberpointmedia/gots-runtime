package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"
)

// Vault stores encrypted secrets and configuration
type Vault struct {
	secrets map[string][]byte
	key     []byte
	mu      sync.RWMutex
}

// NewVault creates a new vault with a master key
func NewVault(masterKey string) (*Vault, error) {
	// Derive key from master key
	hash := sha256.Sum256([]byte(masterKey))

	return &Vault{
		secrets: make(map[string][]byte),
		key:     hash[:],
	}, nil
}

// Set stores a secret in the vault
func (v *Vault) Set(key string, value []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	encrypted, err := v.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	v.secrets[key] = encrypted
	return nil
}

// Get retrieves a secret from the vault
func (v *Vault) Get(key string) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	encrypted, ok := v.secrets[key]
	if !ok {
		return nil, fmt.Errorf("secret not found: %s", key)
	}

	decrypted, err := v.decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return decrypted, nil
}

// Delete removes a secret from the vault
func (v *Vault) Delete(key string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.secrets, key)
}

// List returns all secret keys
func (v *Vault) List() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	keys := make([]string, 0, len(v.secrets))
	for k := range v.secrets {
		keys = append(keys, k)
	}
	return keys
}

// encrypt encrypts data using AES-GCM
func (v *Vault) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (v *Vault) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// SetString stores a string secret
func (v *Vault) SetString(key, value string) error {
	return v.Set(key, []byte(value))
}

// GetString retrieves a string secret
func (v *Vault) GetString(key string) (string, error) {
	data, err := v.Get(key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// VaultManager manages multiple vaults
type VaultManager struct {
	vaults     map[string]*Vault
	defaultKey string
	mu         sync.RWMutex
}

// NewVaultManager creates a new vault manager
func NewVaultManager() *VaultManager {
	return &VaultManager{
		vaults: make(map[string]*Vault),
	}
}

// SetDefaultKey sets the default master key
func (vm *VaultManager) SetDefaultKey(key string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.defaultKey = key
}

// CreateVault creates a new vault
func (vm *VaultManager) CreateVault(name, masterKey string) (*Vault, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vault, err := NewVault(masterKey)
	if err != nil {
		return nil, err
	}

	vm.vaults[name] = vault
	return vault, nil
}

// CreateDefaultVault creates a vault using the default key
func (vm *VaultManager) CreateDefaultVault(name string) (*Vault, error) {
	return vm.CreateVault(name, vm.defaultKey)
}

// GetVault gets a vault by name
func (vm *VaultManager) GetVault(name string) (*Vault, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	vault, ok := vm.vaults[name]
	return vault, ok
}

// DeleteVault deletes a vault
func (vm *VaultManager) DeleteVault(name string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	delete(vm.vaults, name)
}

// ListVaults returns all vault names
func (vm *VaultManager) ListVaults() []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	names := make([]string, 0, len(vm.vaults))
	for name := range vm.vaults {
		names = append(names, name)
	}
	return names
}

// ConfigVault is a specialized vault for configuration
type ConfigVault struct {
	*Vault
	config       map[string]interface{}
	configLocked bool
	mu           sync.RWMutex
}

// NewConfigVault creates a new config vault
func NewConfigVault(masterKey string) (*ConfigVault, error) {
	vault, err := NewVault(masterKey)
	if err != nil {
		return nil, err
	}

	return &ConfigVault{
		Vault:        vault,
		config:       make(map[string]interface{}),
		configLocked: false,
	}, nil
}

// SetConfig sets a configuration value
func (cv *ConfigVault) SetConfig(key string, value interface{}) error {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	if cv.configLocked {
		return fmt.Errorf("config vault is locked")
	}

	cv.config[key] = value
	return nil
}

// GetConfig gets a configuration value
func (cv *ConfigVault) GetConfig(key string) (interface{}, bool) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	value, ok := cv.config[key]
	return value, ok
}

// LockConfig locks the config vault to prevent further modifications
func (cv *ConfigVault) LockConfig() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.configLocked = true
}

// IsLocked checks if config vault is locked
func (cv *ConfigVault) IsLocked() bool {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return cv.configLocked
}

// GetAllConfig returns all configuration as a copy
func (cv *ConfigVault) GetAllConfig() map[string]interface{} {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range cv.config {
		result[k] = v
	}
	return result
}

// Export exports configuration as base64-encoded JSON (for secure transfer)
func (cv *ConfigVault) Export() (string, error) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	// In a real implementation, this would serialize config to JSON and encrypt it
	// For now, return base64 encoded placeholder
	return base64.StdEncoding.EncodeToString([]byte("config")), nil
}

// SecretRotation handles secret rotation and expiration
type SecretRotation struct {
	vault            *Vault
	secrets          map[string]*RotatableSecret
	rotationInterval time.Duration
	notificationFunc func(string, error)
	mu               sync.RWMutex
}

// RotatableSecret represents a secret that can be rotated
type RotatableSecret struct {
	Value     []byte
	CreatedAt time.Time
	RotatedAt time.Time
	ExpiresAt time.Time
	Rotations int
}

// NewSecretRotation creates a new secret rotation manager
func NewSecretRotation(vault *Vault, interval time.Duration) *SecretRotation {
	return &SecretRotation{
		vault:            vault,
		secrets:          make(map[string]*RotatableSecret),
		rotationInterval: interval,
	}
}

// RegisterSecret registers a secret for rotation tracking
func (sr *SecretRotation) RegisterSecret(key string, value []byte, ttl time.Duration) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	now := time.Now()
	sr.secrets[key] = &RotatableSecret{
		Value:     value,
		CreatedAt: now,
		RotatedAt: now,
		ExpiresAt: now.Add(ttl),
		Rotations: 0,
	}

	return sr.vault.Set(key, value)
}

// NeedsRotation checks if a secret needs rotation
func (sr *SecretRotation) NeedsRotation(key string) bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	if secret, ok := sr.secrets[key]; ok {
		return time.Now().After(secret.ExpiresAt)
	}
	return false
}

// RotateSecret rotates a secret
func (sr *SecretRotation) RotateSecret(key string, newValue []byte, ttl time.Duration) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	secret, ok := sr.secrets[key]
	if !ok {
		return fmt.Errorf("secret not found: %s", key)
	}

	now := time.Now()
	secret.Value = newValue
	secret.RotatedAt = now
	secret.ExpiresAt = now.Add(ttl)
	secret.Rotations++

	return sr.vault.Set(key, newValue)
}

// SetNotificationFunc sets the notification function for rotation events
func (sr *SecretRotation) SetNotificationFunc(fn func(string, error)) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.notificationFunc = fn
}

// GetSecretMetadata returns metadata about a secret
func (sr *SecretRotation) GetSecretMetadata(key string) (*RotatableSecret, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	secret, ok := sr.secrets[key]
	return secret, ok
}
