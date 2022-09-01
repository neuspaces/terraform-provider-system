package sshclient

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

func Certificate(cert string, privateKey string) AuthMethod {
	return func() ([]ssh.AuthMethod, error) {
		certSigner, err := signCertWithPrivateKey(privateKey, cert)
		if err != nil {
			return nil, err
		}

		return []ssh.AuthMethod{
			certSigner,
		}, nil
	}
}

// signCertWithPrivateKey returns an ssh.AuthMethod using a client certificate and a private key
func signCertWithPrivateKey(privateKey string, cert string) (ssh.AuthMethod, error) {
	rawPk, err := ssh.ParseRawPrivateKey([]byte(privateKey))
	if err != nil {
		if err.Error() == (&ssh.PassphraseMissingError{}).Error() {
			return nil, fmt.Errorf("sshclient: failed to read ssh private key: password protected keys are not supported. decrypt the key before use")
		}

		return nil, fmt.Errorf("sshclient: failed to parse private key %q: %w", privateKey, err)
	}

	parsedCert, _, _, _, err := ssh.ParseAuthorizedKey([]byte(cert))
	if err != nil {
		return nil, fmt.Errorf("sshclient: failed to parse certificate %q: %w", cert, err)
	}

	signer, err := ssh.NewSignerFromKey(rawPk)
	if err != nil {
		return nil, fmt.Errorf("sshclient: failed to create signer from raw private key %q: %w", rawPk, err)
	}

	certSigner, err := ssh.NewCertSigner(parsedCert.(*ssh.Certificate), signer)
	if err != nil {
		return nil, fmt.Errorf("sshclient: failed to create cert signer %q: %w", signer, err)
	}

	return ssh.PublicKeys(certSigner), nil
}
