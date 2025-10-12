package infra

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

// Private helper function for reading local SSH keys
func (c *CloudProvider) ensureKeys() (key *godo.Key, err error) {
	var homeDir string
	if homeDir, err = os.UserHomeDir(); err != nil {
		return nil, errors.Wrap(err, "failed to get home directory")
	}

	// Now read the public key
	pubBytes, err := os.ReadFile(filepath.Join(homeDir, ".ssh", "id_rsa.pub"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read public key")
	}

	// Check if key already exists in DigitalOcean
	keys, _, err := c.Keys.List(context.Background(), &godo.ListOptions{PerPage: 100})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get keys from Digital Ocean")
	}

	keyData := strings.TrimSpace(string(pubBytes))
	for _, k := range keys {
		if strings.TrimSpace(k.PublicKey) == keyData {
			return &k, nil
		}
	}

	key, _, err = c.Keys.Create(context.Background(), &godo.KeyCreateRequest{
		Name:      "The Skyscape HQ",
		PublicKey: string(pubBytes),
	})

	return key, errors.Wrap(err, "failed to create access key")
}
