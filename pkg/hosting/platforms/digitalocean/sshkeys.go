package digitalocean

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const (
	RSA_KEY_NAME = "SkyGuard Access Key"
	RSA_KEY_TYPE = "RSA PRIVATE KEY"

	ID_RSA     = "id_rsa"
	ID_RSA_PUB = "id_rsa.pub"
)

// If you are trying to read this function, why?
// I'm not a bad person, this is just a very boring function
// and I can't be asked to break this down into smaller
// functions. There are no needs to reuse this code, a lot of
// it is boring crypto code that the CPU can cache if it is
// in one function so read it like a wizard's spell if you must.
func (s *Server) accessKey() (key *godo.Key, err error) {
	ctx := context.Background()

	var homeDir string
	if homeDir, err = os.UserHomeDir(); err != nil {
		return nil, errors.Wrap(err, "failed to get home directory")
	}

	var pubBytes []byte
	if pubBytes, err = os.ReadFile(filepath.Join(homeDir, ".ssh", ID_RSA_PUB)); err != nil {
		return nil, errors.Wrap(err, "failed to read public key")
	}

	var keys []godo.Key
	if keys, _, err = s.client.Keys.List(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "failed to get keys from Digital Ocean")
	}

	keyData := strings.TrimSpace(string(pubBytes))
	for _, k := range keys {
		if k.Name == RSA_KEY_NAME && strings.TrimSpace(k.PublicKey) == keyData {
			return &k, nil
		}
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	os.MkdirAll(sshDir, 0700)

	privKey := fmt.Sprintf("%s/%s", sshDir, ID_RSA)
	if _, err = os.Stat(privKey); err != nil {
		var privateKey *rsa.PrivateKey
		if privateKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
			return nil, errors.Wrap(err, "failed to generate access key")
		}

		var privateKeyFile *os.File
		if privateKeyFile, err = os.Create(privKey); err != nil {
			return nil, errors.Wrap(err, "failed to create access key")
		}

		defer privateKeyFile.Close()
		privateKeyPEM := &pem.Block{
			Type:  RSA_KEY_TYPE,
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		if err = pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
			return nil, errors.Wrap(err, "failed to encode access key")
		}

		if err = os.Chmod(privKey, 0600); err != nil {
			return nil, errors.Wrap(err, "failed to chmod access key")
		}

		var publicKey ssh.PublicKey
		if publicKey, err = ssh.NewPublicKey(&privateKey.PublicKey); err != nil {
			return nil, errors.Wrap(err, "failed to create public key")
		}

		pubKey := fmt.Sprintf("%s/%s", sshDir, ID_RSA_PUB)
		if err = os.WriteFile(pubKey, ssh.MarshalAuthorizedKey(publicKey), 0644); err != nil {
			return nil, errors.Wrap(err, "failed to write public key")
		}
	}

	var data []byte
	if data, err = os.ReadFile(filepath.Join(homeDir, ".ssh", ID_RSA_PUB)); err != nil {
		return nil, errors.Wrap(err, "failed to read public key")
	}

	key, _, err = s.client.Keys.Create(ctx, &godo.KeyCreateRequest{
		Name:      RSA_KEY_NAME,
		PublicKey: string(data),
	})

	return key, errors.Wrap(err, "failed to create access key")
}
