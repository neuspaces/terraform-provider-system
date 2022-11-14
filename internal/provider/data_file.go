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

const dataFileName = "system_file"

func dataFile() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` retrieves information about a file on the remote system.", dataFileName),

		ReadContext: dataFileRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceFileAttrId: {
				Description: "ID of the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceFileAttrPath: {
				Description: "Absolute path to the file.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			resourceFileAttrMode: {
				Description:      "Permissions of the file in octal format like `755`.",
				Type:             schema.TypeString,
				Computed:         true,
				ValidateDiagFunc: validate.FileMode(),
			},
			resourceFileAttrUser: {
				Description: "Name of the user who owns the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceFileAttrUid: {
				Description: "ID of the user who owns the file",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			resourceFileAttrGroup: {
				Description: "Name of the group that owns the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceFileAttrGid: {
				Description: "ID of the group that owns the file",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			resourceFileAttrContent: {
				Description: fmt.Sprintf("Content of the file. Do not use for sensitive content! Only recommended for small text-based payloads such as configuration files etc. In a terraform plan, The content will be stored in plain-text in the terraform state."),
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   false,
			},
			resourceFileAttrMd5Sum: {
				Description: "MD5 checksum of the remote file contents on the system in base64 encoding.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceFileAttrBasename: {
				Description: fmt.Sprintf("Base name of the file. Returns the last element of path. Example: Given the attribute `%[1]s` is `/path/to/file.txt`, the `%[2]s` is `file.txt`.", resourceFileAttrPath, resourceFileAttrBasename),
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataFileSetResourceData(r *client.File, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceFileAttrPath, r.Path)
	_ = d.Set(resourceFileAttrMode, Mode(r.Mode).String())
	_ = d.Set(resourceFileAttrUser, r.User)
	_ = d.Set(resourceFileAttrUid, r.Uid)
	_ = d.Set(resourceFileAttrGroup, r.Group)
	_ = d.Set(resourceFileAttrGid, r.Gid)

	_ = d.Set(resourceFileAttrMd5Sum, r.Md5Sum)
	_ = d.Set(resourceFileAttrBasename, path.Base(r.Path))

	if r.Content != nil {
		_ = d.Set(resourceFileAttrContent, string(r.Content))
	} else {
		_ = d.Set(resourceFileAttrContent, nil)
	}

	return nil
}

func dataFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	includeContentOpt := client.FileClientIncludeContent(true)
	c := client.NewFileClient(p.System, includeContentOpt, client.FileClientCompression(true))

	id := d.Id()

	r, err := c.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr = dataFileSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}
