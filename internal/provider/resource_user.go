package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
	"strconv"
)

const resourceUserName = "system_user"

const (
	resourceUserAttrId     = "id"
	resourceUserAttrName   = "name"
	resourceUserAttrUid    = "uid"
	resourceUserAttrGroup  = "group"
	resourceUserAttrGid    = "gid"
	resourceUserAttrSystem = "system"
	resourceUserAttrHome   = "home"
	resourceUserAttrShell  = "shell"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages a user on the remote system.", resourceUserName),

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceUserAttrId: {
				Description: "ID of the user",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceUserAttrName: {
				Description: "Name of the user",
				Type:        schema.TypeString,
				Required:    true,
			},
			resourceUserAttrUid: {
				Description: "Uid of the user",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			resourceUserAttrGroup: {
				Description:  "Name of the primary group of the user. Group must exist. Mutually exclusive with `gid`.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{resourceUserAttrGid, resourceUserAttrGroup},
			},
			resourceUserAttrGid: {
				Description:  fmt.Sprintf("Gid of the primary group of the user. Group must exist. Either `%s` or `%s` must be provided.", resourceUserAttrGid, resourceUserAttrGroup),
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{resourceUserAttrGid, resourceUserAttrGroup},
			},
			resourceUserAttrSystem: {
				Description: "Set to `true` to create a system user.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			resourceUserAttrHome: {
				Description:      "Path to the home folder of the user. The folder is expected to exist and will not be created.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validate.AbsolutePath(),
			},
			resourceUserAttrShell: {
				Description:      "Login shell of the user.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validate.AbsolutePath(),
			},
		},
	}
}

func resourceUserGetResourceData(d *schema.ResourceData) (*client.User, diag.Diagnostics) {
	r := &client.User{}

	if v, ok := d.GetOk(resourceUserAttrUid); ok {
		r.Uid = to.IntPtr(v.(int))
	}

	if d.HasChange(resourceUserAttrName) {
		r.Name = d.Get(resourceUserAttrName).(string)
	}

	if d.HasChange(resourceUserAttrGroup) {
		r.Group = d.Get(resourceUserAttrGroup).(string)
	}

	if d.HasChange(resourceUserAttrGid) {
		r.Gid = to.IntPtr(d.Get(resourceUserAttrGid).(int))
	}

	if d.HasChange(resourceUserAttrSystem) {
		r.System = to.BoolPtr(d.Get(resourceUserAttrSystem).(bool))
	}

	if d.HasChange(resourceUserAttrHome) {
		r.Home = d.Get(resourceUserAttrHome).(string)
	}

	if d.HasChange(resourceUserAttrShell) {
		r.Shell = d.Get(resourceUserAttrShell).(string)
	}

	return r, nil
}

func resourceUserSetResourceData(r *client.User, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceUserAttrName, r.Name)
	_ = d.Set(resourceUserAttrUid, to.Int(r.Uid))
	_ = d.Set(resourceUserAttrGroup, r.Group)
	_ = d.Set(resourceUserAttrGid, to.Int(r.Gid))
	_ = d.Set(resourceUserAttrSystem, to.Bool(r.System))
	_ = d.Set(resourceUserAttrHome, r.Home)
	_ = d.Set(resourceUserAttrShell, r.Shell)

	return nil
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewUserClient(p.System)

	r, diagErr := resourceUserGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	id, err := c.Create(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(id))

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewUserClient(p.System)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := c.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr = resourceUserSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewUserClient(p.System)

	r, diagErr := resourceUserGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	err := c.Update(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewUserClient(p.System)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.Delete(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
