package sshclient

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
)

type HostKeyVerifier func() (ssh.HostKeyCallback, error)

// StaticHostKey returns an ssh.HostKeyCallback which accepts only the provided public key.
// StaticHostKey expects a base64 encoded public key which is parsed using ssh.ParsePublicKey.
// StaticHostKey uses ssh.FixedHostKey with the decoded public key.
// Use ssh.FixedHostKey instead of StaticHostKey if an ssh.PublicKey is already available.
func StaticHostKey(publicKey string) HostKeyVerifier {
	return func() (ssh.HostKeyCallback, error) {
		pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
		if err != nil {
			return nil, err
		}

		return ssh.FixedHostKey(pk), nil
	}
}

// unconfiguredHostKey returns an ssh.HostKeyCallback which always fails and reminds to configure host key validation.
func unconfiguredHostKey() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return fmt.Errorf("sshclient: host key validation is not configured")
	}
}
