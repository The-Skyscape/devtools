package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VaultClient provides API access to Vault
type VaultClient struct {
	addr   string
	token  string
	client *http.Client
}

// NewVaultClient creates a new Vault API client
func NewVaultClient(addr, token string) *VaultClient {
	return &VaultClient{
		addr:  addr,
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// StoreSecret stores a secret at the given path
func (c *VaultClient) StoreSecret(path string, data map[string]interface{}) error {
	url := fmt.Sprintf("%s/v1/secret/data/%s", c.addr, path)
	
	payload := map[string]interface{}{
		"data": data,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal secret data: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GetSecret retrieves a secret from the given path
func (c *VaultClient) GetSecret(path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/secret/data/%s", c.addr, path)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", c.token)
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found")
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Extract the actual secret data from the Vault response structure
	if data, ok := result["data"].(map[string]interface{}); ok {
		if secretData, ok := data["data"].(map[string]interface{}); ok {
			return secretData, nil
		}
	}
	
	return nil, fmt.Errorf("unexpected vault response format")
}

// DeleteSecret removes a secret at the given path
func (c *VaultClient) DeleteSecret(path string) error {
	url := fmt.Sprintf("%s/v1/secret/metadata/%s", c.addr, path)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Vault-Token", c.token)
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault returned status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Health checks if Vault is healthy
func (c *VaultClient) Health() error {
	url := fmt.Sprintf("%s/v1/sys/health", c.addr)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()
	
	// Vault returns various status codes for health
	// 200 = initialized, unsealed, active
	// 429 = unsealed, standby
	// 501 = not initialized
	// 503 = sealed
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusTooManyRequests {
		return nil
	}
	
	return fmt.Errorf("vault is not healthy (status: %d)", resp.StatusCode)
}

// SetToken updates the client's authentication token
func (c *VaultClient) SetToken(token string) {
	c.token = token
}

// SetAddress updates the client's Vault address
func (c *VaultClient) SetAddress(addr string) {
	c.addr = addr
}