package provider

import (
	"context"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/lib/filemode"
	"path"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
)

const dataFileMetaName = "system_file_meta"

const (
	dataFileMetaAttrId       = "id"
	dataFileMetaAttrPath     = resourceFileAttrPath
	dataFileMetaAttrMode     = resourceFileAttrMode
	dataFileMetaAttrUser     = resourceFileAttrUser
	dataFileMetaAttrUid      = resourceFileAttrUid
	dataFileMetaAttrGroup    = resourceFileAttrGroup
	dataFileMetaAttrGid      = resourceFileAttrGid
	dataFileMetaAttrMd5Sum   = resourceFileAttrMd5Sum
	dataFileMetaAttrBasename = resourceFileAttrBasename
)

func dataFileMeta() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` retrieves meta information about a file on the remote system.", dataFileMetaName),

		ReadContext: dataFileMetaRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			dataFileMetaAttrId: {
				Description: "ID of the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileMetaAttrPath: {
				Description:      "Absolute path to the file.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.AbsolutePath(),
			},
			dataFileMetaAttrMode: {
				Description: "Permissions of the file in octal format like `755`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileMetaAttrUser: {
				Description: "Name of the user who owns the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileMetaAttrUid: {
				Description: "ID of the user who owns the file",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			dataFileMetaAttrGroup: {
				Description: "Name of the group that owns the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileMetaAttrGid: {
				Description: "ID of the group that owns the file",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			dataFileMetaAttrMd5Sum: {
				Description: "MD5 checksum of the remote file contents on the system in base64 encoding.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataFileMetaAttrBasename: {
				Description: fmt.Sprintf("Base name of the file. Returns the last element of path. Example: Given the attribute `%[1]s` is `/path/to/file.txt`, the `%[2]s` is `file.txt`.", dataFileMetaAttrPath, dataFileMetaAttrBasename),
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataFileMetaSetResourceData(r *client.File, d *schema.ResourceData) diag.Diagnostics {
	d.SetId(r.Path)

	_ = d.Set(dataFileMetaAttrPath, r.Path)
	_ = d.Set(dataFileMetaAttrMode, filemode.Mode(r.Mode).String())
	_ = d.Set(dataFileMetaAttrUser, r.User)
	_ = d.Set(dataFileMetaAttrUid, r.Uid)
	_ = d.Set(dataFileMetaAttrGroup, r.Group)
	_ = d.Set(dataFileMetaAttrGid, r.Gid)

	_ = d.Set(dataFileMetaAttrMd5Sum, r.Md5Sum)
	_ = d.Set(dataFileMetaAttrBasename, path.Base(r.Path))

	return nil
}

func dataFileMetaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewFileClient(p.System)

	filePath := d.Get(dataFileMetaAttrPath).(string)

	r, err := c.Get(ctx, filePath)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Path)

	diagErr = dataFileMetaSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}
