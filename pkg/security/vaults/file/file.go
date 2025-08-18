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

	"github.com/The-Skyscape/devtools/pkg/database"
)

// File implements the Secrets interface using encrypted file storage
type File struct {
	baseDir  string
	key      []byte
	commands chan command
}

// command represents a command to the file vault
type command struct {
	action   string
	path     string
	data     map[string]interface{}
	response chan response
}

// response represents a response from the file vault
type response struct {
	data  interface{}
	error error
}

// New creates a new file-based storage vault
func New(dir ...string) *File {
	baseDir := filepath.Join(database.DataDir(), "secrets")
	if len(dir) > 0 && dir[0] != "" {
		baseDir = dir[0]
	}
	
	return &File{
		baseDir:  baseDir,
		commands: make(chan command),
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
	
	// Start command processor
	go f.run()
	
	log.Printf("File: Initialized encrypted storage at %s", f.baseDir)
	return nil
}

// run handles all commands in a single goroutine
func (f *File) run() {
	secrets := make(map[string]map[string]interface{})
	
	// Load existing secrets on startup
	if loadedSecrets, err := f.loadSecretsFromDisk(); err == nil {
		secrets = loadedSecrets
	}
	
	for cmd := range f.commands {
		switch cmd.action {
		case "store":
			// Create a copy to avoid reference issues
			secretCopy := make(map[string]interface{})
			for k, v := range cmd.data {
				secretCopy[k] = v
			}
			secrets[cmd.path] = secretCopy
			
			// Save to disk
			err := f.saveSecretsToDisk(secrets)
			cmd.response <- response{error: err}
			
		case "get":
			secret, exists := secrets[cmd.path]
			if !exists {
				cmd.response <- response{
					error: fmt.Errorf("secret not found: %s", cmd.path),
				}
			} else {
				// Return a copy to avoid modification
				secretCopy := make(map[string]interface{})
				for k, v := range secret {
					secretCopy[k] = v
				}
				cmd.response <- response{data: secretCopy}
			}
			
		case "delete":
			delete(secrets, cmd.path)
			err := f.saveSecretsToDisk(secrets)
			cmd.response <- response{error: err}
			
		case "list":
			paths := make([]string, 0, len(secrets))
			for path := range secrets {
				paths = append(paths, path)
			}
			cmd.response <- response{data: paths}
			
		case "close":
			cmd.response <- response{error: nil}
			close(f.commands)
			return
		}
	}
}

// Close closes the file vault
func (f *File) Close() error {
	if f.commands == nil {
		return nil
	}
	
	resp := make(chan response)
	f.commands <- command{
		action:   "close",
		response: resp,
	}
	result := <-resp
	return result.error
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
	resp := make(chan response)
	f.commands <- command{
		action:   "store",
		path:     path,
		data:     data,
		response: resp,
	}
	result := <-resp
	return result.error
}

// GetSecret retrieves a secret from the file vault
func (f *File) GetSecret(path string) (map[string]interface{}, error) {
	resp := make(chan response)
	f.commands <- command{
		action:   "get",
		path:     path,
		response: resp,
	}
	result := <-resp
	if result.error != nil {
		return nil, result.error
	}
	return result.data.(map[string]interface{}), nil
}

// DeleteSecret removes a secret from the file vault
func (f *File) DeleteSecret(path string) error {
	resp := make(chan response)
	f.commands <- command{
		action:   "delete",
		path:     path,
		response: resp,
	}
	result := <-resp
	return result.error
}

// ListSecrets returns all secret paths
func (f *File) ListSecrets() ([]string, error) {
	resp := make(chan response)
	f.commands <- command{
		action:   "list",
		response: resp,
	}
	result := <-resp
	if result.error != nil {
		return nil, result.error
	}
	return result.data.([]string), nil
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