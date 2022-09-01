package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sync"
)

// SyncResource holds a sync.RWMutex
type SyncResource struct {
	m sync.RWMutex

	CreateContext schema.CreateContextFunc
	ReadContext   schema.ReadContextFunc
	UpdateContext schema.UpdateContextFunc
	DeleteContext schema.DeleteContextFunc
}

func (r *SyncResource) CreateContextSync(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	r.m.Lock()
	defer r.m.Unlock()
	return r.CreateContext(ctx, d, meta)
}

func (r *SyncResource) ReadContextSync(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.ReadContext(ctx, d, meta)
}

func (r *SyncResource) UpdateContextSync(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	r.m.Lock()
	defer r.m.Unlock()
	return r.UpdateContext(ctx, d, meta)
}
func (r *SyncResource) DeleteContextSync(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	r.m.Lock()
	defer r.m.Unlock()
	return r.DeleteContext(ctx, d, meta)
}
