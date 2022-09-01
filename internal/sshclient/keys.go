package sshclient

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
)

// ParseBase64PublicKey parses a base64 encoded SSH public key using ssh.ParsePublicKey
func ParseBase64PublicKey(publicKey string) (ssh.PublicKey, error) {
	publicKeyTrimmed := strings.TrimSpace(publicKey)

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyTrimmed)
	if err != nil {
		return nil, fmt.Errorf("sshclient: failed to decode public key: %w", err)
	}

	parsedPublicKey, err := ssh.ParsePublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("sshclient: failed to parse public key: %w", err)
	}

	return parsedPublicKey, nil
}
