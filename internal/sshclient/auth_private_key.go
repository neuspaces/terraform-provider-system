package sshclient

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

// PrivateKey returns an AuthMethod which authenticates using a private key.
func PrivateKey(privateKey string) AuthMethod {
	return func() ([]ssh.AuthMethod, error) {
		privateKeySigner, err := parsePrivateKey(privateKey)
		if err != nil {
			return nil, err
		}

		return []ssh.AuthMethod{
			ssh.PublicKeys(privateKeySigner),
		}, nil
	}
}

func parsePrivateKey(privateKey string) (ssh.Signer, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		if err.Error() == (&ssh.PassphraseMissingError{}).Error() {
			return nil, fmt.Errorf("sshclient: failed to read ssh private key: password protected keys are not supported. decrypt the key before use")
		}

		return nil, fmt.Errorf("sshclient: failed to read ssh private key: %w", err)
	}

	return signer, nil
}
