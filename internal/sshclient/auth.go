package sshclient

import "golang.org/x/crypto/ssh"

type AuthMethod func() ([]ssh.AuthMethod, error)
