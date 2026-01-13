package api

import (
	"os"
)

// Env provides environment variable operations
type Env struct{}

// NewEnv creates a new environment API
func NewEnv() *Env {
	return &Env{}
}

// Get gets an environment variable
func (e *Env) Get(key string) string {
	return os.Getenv(key)
}

// Set sets an environment variable
func (e *Env) Set(key, value string) error {
	return os.Setenv(key, value)
}

// Unset unsets an environment variable
func (e *Env) Unset(key string) error {
	return os.Unsetenv(key)
}

// GetAll gets all environment variables
func (e *Env) GetAll() map[string]string {
	env := make(map[string]string)
	for _, pair := range os.Environ() {
		for i := 0; i < len(pair); i++ {
			if pair[i] == '=' {
				key := pair[:i]
				value := pair[i+1:]
				env[key] = value
				break
			}
		}
	}
	return env
}

// LookupEnv looks up an environment variable
func (e *Env) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

