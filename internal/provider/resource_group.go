package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"strconv"
)

const resourceGroupName = "system_group"

const (
	resourceGroupAttrId     = "id"
	resourceGroupAttrName   = "name"
	resourceGroupAttrGid    = "gid"
	resourceGroupAttrSystem = "system"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages a group on the remote system.", resourceGroupName),

		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceGroupAttrId: {
				Description: "ID of the group",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceGroupAttrName: {
				Description: "Name of the group",
				Type:        schema.TypeString,
				Required:    true,
			},
			resourceGroupAttrGid: {
				Description: "Gid of the group. If not defined, a gid will be generated.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			resourceGroupAttrSystem: {
				Description: "Set to `true` to create a system group.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceGroupGetResourceData(d *schema.ResourceData) (*client.Group, diag.Diagnostics) {
	r := &client.Group{
		Name:   "",
		Gid:    intOrDefault(optional(d.GetOk(resourceGroupAttrGid)), -1),
		System: d.Get(resourceGroupAttrSystem).(bool),
	}

	if d.HasChange(resourceGroupAttrName) {
		r.Name = d.Get(resourceGroupAttrName).(string)
	}

	return r, nil
}

func resourceGroupSetResourceData(r *client.Group, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceGroupAttrName, r.Name)
	_ = d.Set(resourceGroupAttrGid, r.Gid)
	_ = d.Set(resourceGroupAttrSystem, r.System)

	return nil
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewGroupClient(p.System)

	r, diagErr := resourceGroupGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	id, err := c.Create(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(id))

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewGroupClient(p.System)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := c.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr = resourceGroupSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewGroupClient(p.System)

	r, diagErr := resourceGroupGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	err := c.Update(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewGroupClient(p.System)

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
