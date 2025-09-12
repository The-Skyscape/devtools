package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/The-Skyscape/devtools/pkg/database"
)

// FileBackend implements the Secrets interface using encrypted file storage
type FileBackend struct {
	baseDir  string
	key      []byte
	commands chan fileCommand
}

// fileCommand represents a command to the file backend
type fileCommand struct {
	action   string
	path     string
	data     map[string]any
	response chan fileResponse
}

// fileResponse represents a response from the file backend
type fileResponse struct {
	data  any
	error error
}

// NewFileBackend creates a new file-based storage backend
func NewFileBackend() *FileBackend {
	return &FileBackend{
		baseDir:  filepath.Join(database.DataDir(), "secrets"),
		commands: make(chan fileCommand),
	}
}

// NewFileBackendWithDir creates a new file-based storage backend with custom directory
func NewFileBackendWithDir(dir string) *FileBackend {
	return &FileBackend{
		baseDir:  dir,
		commands: make(chan fileCommand),
	}
}

// Init initializes the file backend
func (f *FileBackend) Init() error {
	// Create secrets directory if it doesn't exist
	if err := os.MkdirAll(f.baseDir, 0700); err != nil {
		return fmt.Errorf("failed to create secrets directory: %w", err)
	}

	// Generate or load encryption key
	keyPath := filepath.Join(f.baseDir, ".key")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// Generate new key
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}

		// Save key
		if err := os.WriteFile(keyPath, key, 0600); err != nil {
			return fmt.Errorf("failed to save encryption key: %w", err)
		}

		f.key = key
	} else {
		// Load existing key
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to load encryption key: %w", err)
		}

		if len(key) != 32 {
			return fmt.Errorf("invalid encryption key length")
		}

		f.key = key
	}

	// Start command processor
	go f.run()

	log.Printf("FileBackend: Initialized file-based storage at %s", f.baseDir)
	return nil
}

// run handles all commands in a single goroutine
func (f *FileBackend) run() {
	secrets := make(map[string]map[string]any)

	// Load existing secrets on startup
	if loadedSecrets, err := f.loadSecretsFromDisk(); err == nil {
		secrets = loadedSecrets
	}

	for cmd := range f.commands {
		switch cmd.action {
		case "store":
			// Create a copy to avoid reference issues
			secretCopy := make(map[string]any)
			for k, v := range cmd.data {
				secretCopy[k] = v
			}
			secrets[cmd.path] = secretCopy

			// Save to disk
			err := f.saveSecretsToDisk(secrets)
			cmd.response <- fileResponse{error: err}

		case "get":
			secret, exists := secrets[cmd.path]
			if !exists {
				cmd.response <- fileResponse{
					error: fmt.Errorf("secret not found: %s", cmd.path),
				}
			} else {
				// Return a copy to avoid modification
				secretCopy := make(map[string]any)
				for k, v := range secret {
					secretCopy[k] = v
				}
				cmd.response <- fileResponse{data: secretCopy}
			}

		case "delete":
			delete(secrets, cmd.path)
			err := f.saveSecretsToDisk(secrets)
			cmd.response <- fileResponse{error: err}

		case "list":
			paths := make([]string, 0, len(secrets))
			for path := range secrets {
				paths = append(paths, path)
			}
			cmd.response <- fileResponse{data: paths}

		case "close":
			cmd.response <- fileResponse{error: nil}
			return
		}
	}
}

// Close closes the file backend
func (f *FileBackend) Close() error {
	response := make(chan fileResponse)
	f.commands <- fileCommand{
		action:   "close",
		response: response,
	}
	result := <-response
	close(f.commands)
	return result.error
}

// IsAvailable returns true (file backend is always available)
func (f *FileBackend) IsAvailable() bool {
	return true
}

// GetStorageMode returns "file"
func (f *FileBackend) GetStorageMode() string {
	return "file"
}

// GetStatus returns the file backend status
func (f *FileBackend) GetStatus() any {
	return map[string]any{
		"mode":      "file",
		"directory": f.baseDir,
		"encrypted": true,
	}
}

// StoreSecret stores a secret in the file backend
func (f *FileBackend) StoreSecret(path string, data map[string]any) error {
	response := make(chan fileResponse)
	f.commands <- fileCommand{
		action:   "store",
		path:     path,
		data:     data,
		response: response,
	}
	result := <-response
	return result.error
}

// GetSecret retrieves a secret from the file backend
func (f *FileBackend) GetSecret(path string) (map[string]any, error) {
	response := make(chan fileResponse)
	f.commands <- fileCommand{
		action:   "get",
		path:     path,
		response: response,
	}
	result := <-response
	if result.error != nil {
		return nil, result.error
	}
	return result.data.(map[string]any), nil
}

// DeleteSecret removes a secret from the file backend
func (f *FileBackend) DeleteSecret(path string) error {
	response := make(chan fileResponse)
	f.commands <- fileCommand{
		action:   "delete",
		path:     path,
		response: response,
	}
	result := <-response
	return result.error
}

// ListSecrets returns all secret paths
func (f *FileBackend) ListSecrets() ([]string, error) {
	response := make(chan fileResponse)
	f.commands <- fileCommand{
		action:   "list",
		response: response,
	}
	result := <-response
	if result.error != nil {
		return nil, result.error
	}
	return result.data.([]string), nil
}

// loadSecretsFromDisk loads all secrets from disk
func (f *FileBackend) loadSecretsFromDisk() (map[string]map[string]any, error) {
	secretsFile := filepath.Join(f.baseDir, "secrets.enc")

	if _, err := os.Stat(secretsFile); os.IsNotExist(err) {
		// No existing secrets
		return make(map[string]map[string]any), nil
	}

	// Read encrypted data
	encryptedData, err := os.ReadFile(secretsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	// Decrypt data
	decryptedData, err := f.decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secrets: %w", err)
	}

	// Unmarshal secrets
	var secrets map[string]map[string]any
	if err := json.Unmarshal(decryptedData, &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets: %w", err)
	}

	return secrets, nil
}

// saveSecretsToDisk saves all secrets to disk
func (f *FileBackend) saveSecretsToDisk(secrets map[string]map[string]any) error {
	// Marshal secrets
	data, err := json.Marshal(secrets)
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	// Encrypt data
	encryptedData, err := f.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt secrets: %w", err)
	}

	// Write to temp file first
	secretsFile := filepath.Join(f.baseDir, "secrets.enc")
	tempFile := secretsFile + ".tmp"

	if err := os.WriteFile(tempFile, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, secretsFile); err != nil {
		return fmt.Errorf("failed to save secrets file: %w", err)
	}

	return nil
}

// encrypt encrypts data using AES-GCM
func (f *FileBackend) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(f.key)
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
func (f *FileBackend) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(f.key)
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
