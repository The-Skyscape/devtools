package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides a simple Vault API client
type Client struct {
	address string
	token   string
	client  *http.Client
}

// NewClient creates a new Vault API client
func NewClient(address, token string) *Client {
	return &Client{
		address: address,
		token:   token,
		client: &http.Client{
			Timeout: 5 * time.Second, // Add 5 second timeout to prevent hanging
		},
	}
}

// StoreSecret stores a secret at the given path
func (c *Client) StoreSecret(path string, data map[string]any) error {
	url := fmt.Sprintf("%s/v1/secret/data/%s", c.address, path)

	payload := map[string]any{
		"data": data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to store secret: %s", body)
	}

	return nil
}

// GetSecret retrieves a secret from the given path
func (c *Client) GetSecret(path string) (map[string]any, error) {
	url := fmt.Sprintf("%s/v1/secret/data/%s", c.address, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Check for 404 or other errors
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get secret (status %d): %s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Check if data exists (vault returns {"data":null} for non-existent secrets)
	if result["data"] == nil {
		return nil, fmt.Errorf("secret not found")
	}

	// Extract the actual secret data
	if data, ok := result["data"].(map[string]any); ok {
		if secretData, ok := data["data"].(map[string]any); ok {
			return secretData, nil
		}
	}

	return nil, fmt.Errorf("unexpected response format")
}

// DeleteSecret removes a secret at the given path
func (c *Client) DeleteSecret(path string) error {
	url := fmt.Sprintf("%s/v1/secret/metadata/%s", c.address, path)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete secret: %s", body)
	}

	return nil
}
