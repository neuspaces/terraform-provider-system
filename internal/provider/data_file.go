package provider

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
)

const dataFileName = "system_file"

const (
	dataFileAttrId       = "id"
	dataFileAttrPath     = resourceFileAttrPath
	dataFileAttrMode     = resourceFileAttrMode
	dataFileAttrUser     = resourceFileAttrUser
	dataFileAttrUid      = resourceFileAttrUid
	dataFileAttrGroup    = resourceFileAttrGroup
	dataFileAttrGid      = resourceFileAttrGid
	dataFileAttrContent  = resourceFileAttrContent
	dataFileAttrMd5Sum   = resourceFileAttrMd5Sum
	dataFileAttrBasename = resourceFileAttrBasename
)

func dataFile() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` retrieves meta information about and content of a file on the remote system.", dataFileName),

		ReadContext: dataFileRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			dataFileAttrId: {
				Description: "ID of the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileAttrPath: {
				Description:      "Absolute path to the file.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.AbsolutePath(),
			},
			dataFileAttrMode: {
				Description: "Permissions of the file in octal format like `755`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileAttrUser: {
				Description: "Name of the user who owns the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileAttrUid: {
				Description: "ID of the user who owns the file",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			dataFileAttrGroup: {
				Description: "Name of the group that owns the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileAttrGid: {
				Description: "ID of the group that owns the file",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			dataFileAttrContent: {
				Description: "Content of the file",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			dataFileAttrMd5Sum: {
				Description: "MD5 checksum of the remote file contents on the system in base64 encoding.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileAttrBasename: {
				Description: fmt.Sprintf("Base name of the file. Returns the last element of path. Example: Given the attribute `%[1]s` is `/path/to/file.txt`, the `%[2]s` is `file.txt`.", dataFileAttrPath, dataFileAttrBasename),
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataFileSetResourceData(r *client.File, d *schema.ResourceData) diag.Diagnostics {
	d.SetId(r.Path)

	_ = d.Set(dataFileAttrPath, r.Path)
	_ = d.Set(dataFileAttrMode, Mode(r.Mode).String())
	_ = d.Set(dataFileAttrUser, r.User)
	_ = d.Set(dataFileAttrUid, r.Uid)
	_ = d.Set(dataFileAttrGroup, r.Group)
	_ = d.Set(dataFileAttrGid, r.Gid)

	_ = d.Set(dataFileAttrMd5Sum, r.Md5Sum)
	_ = d.Set(dataFileAttrBasename, path.Base(r.Path))

	if r.Content != nil {
		_ = d.Set(dataFileAttrContent, string(r.Content))
	} else {
		_ = d.Set(dataFileAttrContent, nil)
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

	filePath := d.Get(dataFileAttrPath).(string)

	r, err := c.Get(ctx, filePath)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Path)

	diagErr = dataFileSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}
