package client

import (
	"context"
	"errors"
	"github.com/neuspaces/terraform-provider-system/internal/system"
)

type InfoClient interface {
	GetRelease(ctx context.Context) (*ReleaseInfo, error)
	GetIdentity(ctx context.Context) (*IdentityInfo, error)
}

func NewInfoClient(s system.System) InfoClient {
	return &infoClient{
		s: s,
	}
}

var (
	ErrInfo = errors.New("info resource")

	ErrInfoUnexpected = errors.Join(ErrInfo, errors.New("unexpected error"))
)

type infoClient struct {
	s system.System
}
