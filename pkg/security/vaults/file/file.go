package file

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
	"sync"

	"github.com/The-Skyscape/devtools/pkg/database"
)

// File implements the Secrets interface using encrypted file storage
type File struct {
	baseDir string
	key     []byte
	mu      sync.RWMutex // Simple mutex for thread safety
	secrets map[string]map[string]interface{}
}

// New creates a new file-based storage vault
func New(dir ...string) *File {
	baseDir := filepath.Join(database.DataDir(), "secrets")
	if len(dir) > 0 && dir[0] != "" {
		baseDir = dir[0]
	}
	
	return &File{
		baseDir: baseDir,
		secrets: make(map[string]map[string]interface{}),
	}
}

// Init initializes the file vault
func (f *File) Init() error {
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
	
	// Load existing secrets
	if loadedSecrets, err := f.loadSecretsFromDisk(); err == nil {
		f.secrets = loadedSecrets
	}
	
	log.Printf("File: Initialized encrypted storage at %s", f.baseDir)
	return nil
}

// Close closes the file vault
func (f *File) Close() error {
	// Nothing to close for file storage
	return nil
}

// IsAvailable returns true (file vault is always available)
func (f *File) IsAvailable() bool {
	return true
}

// GetStorageMode returns "file"
func (f *File) GetStorageMode() string {
	return "file"
}

// GetStatus returns the file vault status
func (f *File) GetStatus() interface{} {
	return map[string]interface{}{
		"mode":      "file",
		"directory": f.baseDir,
		"encrypted": true,
	}
}

// StoreSecret stores a secret in the file vault
func (f *File) StoreSecret(path string, data map[string]interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// Create a copy to avoid reference issues
	secretCopy := make(map[string]interface{})
	for k, v := range data {
		secretCopy[k] = v
	}
	f.secrets[path] = secretCopy
	
	// Save to disk
	return f.saveSecretsToDisk(f.secrets)
}

// GetSecret retrieves a secret from the file vault
func (f *File) GetSecret(path string) (map[string]interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	secret, exists := f.secrets[path]
	if !exists {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	
	// Return a copy to avoid modification
	secretCopy := make(map[string]interface{})
	for k, v := range secret {
		secretCopy[k] = v
	}
	
	return secretCopy, nil
}

// DeleteSecret removes a secret from the file vault
func (f *File) DeleteSecret(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	delete(f.secrets, path)
	return f.saveSecretsToDisk(f.secrets)
}

// ListSecrets returns all secret paths
func (f *File) ListSecrets() ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	paths := make([]string, 0, len(f.secrets))
	for path := range f.secrets {
		paths = append(paths, path)
	}
	
	return paths, nil
}

// loadSecretsFromDisk loads all secrets from disk
func (f *File) loadSecretsFromDisk() (map[string]map[string]interface{}, error) {
	secretsFile := filepath.Join(f.baseDir, "secrets.enc")
	
	if _, err := os.Stat(secretsFile); os.IsNotExist(err) {
		// No existing secrets
		return make(map[string]map[string]interface{}), nil
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
	var secrets map[string]map[string]interface{}
	if err := json.Unmarshal(decryptedData, &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secrets: %w", err)
	}
	
	return secrets, nil
}

// saveSecretsToDisk saves all secrets to disk
func (f *File) saveSecretsToDisk(secrets map[string]map[string]interface{}) error {
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
func (f *File) encrypt(plaintext []byte) ([]byte, error) {
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
func (f *File) decrypt(ciphertext []byte) ([]byte, error) {
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