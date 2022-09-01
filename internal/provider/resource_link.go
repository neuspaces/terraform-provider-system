package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
)

const resourceLinkName = "system_link"

const (
	resourceLinkAttrId     = "id"
	resourceLinkAttrPath   = "path"
	resourceLinkAttrTarget = "target"
	resourceLinkAttrUser   = "user"
	resourceLinkAttrUid    = "uid"
	resourceLinkAttrGroup  = "group"
	resourceLinkAttrGid    = "gid"
)

func resourceLink() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages a symbolic link on the remote system.", resourceLinkName),

		CreateContext: resourceLinkCreate,
		ReadContext:   resourceLinkRead,
		UpdateContext: resourceLinkUpdate,
		DeleteContext: resourceLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceLinkAttrId: {
				Description: "ID of the link",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceLinkAttrPath: {
				Description:      "Path of the link. Must be an absolute path. Not to be confused with the target.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.AbsolutePath(),
			},
			resourceLinkAttrTarget: {
				Description: "Target of the link. Can be either an absolute or a relative path. Target is not required to exist when link is created.",
				Type:        schema.TypeString,
				Required:    true,
			},
			resourceLinkAttrUser: {
				Description:   "Name of the user who owns the link. Does *not* change the user owning the target.",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceLinkAttrUid},
			},
			resourceLinkAttrUid: {
				Description:   "ID of the user who owns the link. Does *not* change the user owning the target.",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceLinkAttrUser},
			},
			resourceLinkAttrGroup: {
				Description:   "Name of the group that owns the link. Does *not* change the group owning the target.",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceLinkAttrGid},
			},
			resourceLinkAttrGid: {
				Description:   "ID of the group that owns the link. Does *not* change the group owning the target.",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceLinkAttrGroup},
			},
		},
	}
}

func resourceLinkGetResourceData(d *schema.ResourceData) (*client.Link, diag.Diagnostics) {
	r := &client.Link{
		Path:   d.Get(resourceLinkAttrPath).(string),
		Target: "",
		User:   "",
		Uid:    -1,
		Group:  "",
		Gid:    -1,
	}

	if d.HasChange(resourceLinkAttrTarget) {
		r.Target = d.Get(resourceLinkAttrTarget).(string)
	}

	if d.HasChange(resourceLinkAttrUser) {
		r.User = d.Get(resourceLinkAttrUser).(string)
	}

	if d.HasChange(resourceLinkAttrUid) {
		r.Uid = intOrDefault(optional(d.GetOk(resourceLinkAttrUid)), -1)
	}

	if d.HasChange(resourceLinkAttrGroup) {
		r.Group = d.Get(resourceLinkAttrGroup).(string)
	}

	if d.HasChange(resourceLinkAttrGid) {
		r.Gid = intOrDefault(optional(d.GetOk(resourceLinkAttrGid)), -1)
	}

	return r, nil
}

func resourceLinkSetResourceData(r *client.Link, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceLinkAttrPath, r.Path)
	_ = d.Set(resourceLinkAttrTarget, r.Target)
	_ = d.Set(resourceLinkAttrUser, r.User)
	_ = d.Set(resourceLinkAttrUid, r.Uid)
	_ = d.Set(resourceLinkAttrGroup, r.Group)
	_ = d.Set(resourceLinkAttrGid, r.Gid)

	return nil
}

func resourceLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewLinkClient(p.System)

	r, diagErr := resourceLinkGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	err := c.Create(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Path)

	return resourceLinkRead(ctx, d, meta)
}

func resourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewLinkClient(p.System)

	id := d.Id()

	r, err := c.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr = resourceLinkSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewLinkClient(p.System)

	r, diagErr := resourceLinkGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	err := c.Update(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceLinkRead(ctx, d, meta)
}

func resourceLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewLinkClient(p.System)

	id := d.Id()

	err := c.Delete(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
