package provider

import (
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/sethvargo/go-retry"
	"time"
)

func resourceServiceClientGet(ctx context.Context, c client.ServiceClient, args client.ServiceGetArgs) (*client.Service, error) {
	var r *client.Service

	err := retry.Do(ctx, retry.NewConstant(5*time.Second), func(ctx context.Context) error {
		var err error

		r, err = c.Get(ctx, args)
		if err != nil {
			return err
		}

		if r.Status != nil && r.Status.IsPending() {
			return retry.RetryableError(fmt.Errorf("pending service status: %s", *r.Status))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}
