package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
	"path"
)

const resourceFolderName = "system_folder"

const (
	resourceFolderAttrId       = "id"
	resourceFolderAttrPath     = "path"
	resourceFolderAttrMode     = "mode"
	resourceFolderAttrUser     = "user"
	resourceFolderAttrUid      = "uid"
	resourceFolderAttrGroup    = "group"
	resourceFolderAttrGid      = "gid"
	resourceFolderAttrBasename = "basename"
)

func resourceFolder() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages a folder on the remote system.", resourceFolderName),

		CreateContext: resourceFolderCreate,
		ReadContext:   resourceFolderRead,
		UpdateContext: resourceFolderUpdate,
		DeleteContext: resourceFolderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceFolderAttrId: {
				Description: "ID of the folder",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceFolderAttrPath: {
				Description:      "Path to the folder",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.AbsolutePath(),
			},
			resourceFolderAttrMode: {
				Description:      "Permissions of the folder in octal format like `755`. Defaults to the umask of the system.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validate.FileMode(),
			},
			resourceFolderAttrUser: {
				Description:   "Name of the user who owns the folder",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFolderAttrUid},
			},
			resourceFolderAttrUid: {
				Description:   "ID of the user who owns the folder",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFolderAttrUser},
			},
			resourceFolderAttrGroup: {
				Description:   "Name of the group that owns the folder",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFolderAttrGid},
			},
			resourceFolderAttrGid: {
				Description:   "ID of the group that owns the folder",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFolderAttrGroup},
			},
			resourceFolderAttrBasename: {
				Description: fmt.Sprintf("Base name of the folder. Returns the last element of path. Example: Given the attribute `%[1]s` is `/path/to/folder`, the `%[2]s` is `folder`.", resourceFolderAttrPath, resourceFolderAttrBasename),
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceFolderGetResourceData(d *schema.ResourceData) (*client.Folder, diag.Diagnostics) {
	r := &client.Folder{
		Path:  d.Get(resourceFolderAttrPath).(string),
		Mode:  0,
		User:  "",
		Uid:   -1,
		Group: "",
		Gid:   -1,
	}

	if d.HasChange(resourceFolderAttrMode) {
		r.Mode = mustParseMode(d.Get(resourceFolderAttrMode).(string))
	}

	if d.HasChange(resourceFolderAttrUser) {
		r.User = d.Get(resourceFolderAttrUser).(string)
	}

	if d.HasChange(resourceFolderAttrUid) {
		r.Uid = intOrDefault(optional(d.GetOk(resourceFolderAttrUid)), -1)
	}

	if d.HasChange(resourceFolderAttrGroup) {
		r.Group = d.Get(resourceFolderAttrGroup).(string)
	}

	if d.HasChange(resourceFolderAttrGid) {
		r.Gid = intOrDefault(optional(d.GetOk(resourceFolderAttrGid)), -1)
	}

	return r, nil
}

func resourceFolderSetResourceData(r *client.Folder, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceFolderAttrPath, r.Path)
	_ = d.Set(resourceFolderAttrMode, Mode(r.Mode).String())
	_ = d.Set(resourceFolderAttrUser, r.User)
	_ = d.Set(resourceFolderAttrUid, r.Uid)
	_ = d.Set(resourceFolderAttrGroup, r.Group)
	_ = d.Set(resourceFolderAttrGid, r.Gid)
	_ = d.Set(resourceFolderAttrBasename, path.Base(r.Path))

	return nil
}

func resourceFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewFolderClient(p.System)

	r, diagErr := resourceFolderGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	err := c.Create(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Path)

	return resourceFolderRead(ctx, d, meta)
}

func resourceFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewFolderClient(p.System)

	id := d.Id()

	r, err := c.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr = resourceFolderSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceFolderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewFolderClient(p.System)

	r, diagErr := resourceFolderGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	err := c.Update(ctx, *r)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFolderRead(ctx, d, meta)
}

func resourceFolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewFolderClient(p.System)

	id := d.Id()

	err := c.Delete(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
