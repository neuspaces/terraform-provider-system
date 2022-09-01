package client

import (
	"context"
	"github.com/neuspaces/terraform-provider-system/internal/lib/typederror"
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
	ErrInfo = typederror.NewRoot("info resource")

	ErrInfoUnexpected = typederror.New("unexpected error", ErrInfo)
)

type infoClient struct {
	s system.System
}
