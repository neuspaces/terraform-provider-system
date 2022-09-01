package sshclient

import (
	"fmt"
	sshagent "github.com/xanzy/ssh-agent"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"net"
)

func Agent() AuthMethod {
	return func() ([]ssh.AuthMethod, error) {
		return []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
				return agentConnectFuncSigners(sshagent.New)
			})}, nil
	}
}

func AgentExplicitIdentities(identities ...string) AuthMethod {
	return func() ([]ssh.AuthMethod, error) {
		return []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
				explicitPublicKeys := map[string]struct{}{}

				// Parse identities
				for _, identity := range identities {
					publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(identity))
					if err != nil {
						return nil, err
					}
					explicitPublicKeys[string(publicKey.Marshal())] = struct{}{}
				}

				signers, err := agentConnectFuncSigners(sshagent.New)
				if err != nil {
					return nil, err
				}

				// Filter signers by explicit identities
				var explicitSigners []ssh.Signer
				for _, signer := range signers {
					if _, ok := explicitPublicKeys[string(signer.PublicKey().Marshal())]; ok {
						explicitSigners = append(explicitSigners, signer)
					}
				}

				// Require at least one explicit identity which provided the agent
				if len(explicitSigners) == 0 {
					return nil, fmt.Errorf("sshclient: agent does not provide any explicit identity")
				}

				return explicitSigners, nil
			})}, nil
	}
}

type agentConnectFunc func() (agent.Agent, net.Conn, error)

func agentConnectFuncSigners(agentConnectFunc agentConnectFunc) ([]ssh.Signer, error) {
	a, conn, err := agentConnectFunc()
	defer func() {
		if a != nil {
			_ = conn.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	keys, err := a.List()
	if err != nil {
		return nil, err
	}

	var signers []ssh.Signer
	for _, k := range keys {
		signers = append(signers, &agentConnectFuncSigner{agentConnectFunc, k})
	}
	return signers, nil
}

// agentConnectFuncSigner is like ssh.agentConnectFuncSigner but creates a new agent.Agent for each sign operation
type agentConnectFuncSigner struct {
	agentFunc agentConnectFunc
	pub       ssh.PublicKey
}

func (s *agentConnectFuncSigner) PublicKey() ssh.PublicKey {
	return s.pub
}

func (s *agentConnectFuncSigner) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	a, conn, err := s.agentFunc()
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()
	if err != nil {
		return nil, err
	}

	return a.Sign(s.pub, data)
}
