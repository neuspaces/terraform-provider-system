package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/source"
	"github.com/neuspaces/terraform-provider-system/internal/validate"
	"io"
	"path"
	"strings"
)

const resourceFileName = "system_file"

const (
	resourceFileAttrId               = "id"
	resourceFileAttrPath             = "path"
	resourceFileAttrMode             = "mode"
	resourceFileAttrUser             = "user"
	resourceFileAttrUid              = "uid"
	resourceFileAttrGroup            = "group"
	resourceFileAttrGid              = "gid"
	resourceFileAttrContent          = "content"
	resourceFileAttrContentSensitive = "content_sensitive"
	resourceFileAttrSource           = "source"
	resourceFileAttrMd5Sum           = "md5sum"
	resourceFileAttrBasename         = "basename"
)

func resourceFile() *schema.Resource {
	// Configure source registry
	sources, err := source.NewRegistry(
		source.WithClients(
			source.NewMetaCache(source.NewFileClient()),
			source.NewMetaCache(source.NewHttpClient()),
		),
		source.WithDefaultScheme(source.FileScheme),
	)
	if err != nil {
		panic(err)
	}

	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages a file on the remote system.", resourceFileName),

		CreateContext: resourceFileCreateFactory(sources),
		ReadContext:   resourceFileRead,
		UpdateContext: resourceFileUpdateFactory(sources),
		DeleteContext: resourceFileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceFileAttrId: {
				Description: "ID of the file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceFileAttrPath: {
				Description:      "Path to the file. Must be an absolute path.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.AbsolutePath(),
			},
			resourceFileAttrMode: {
				Description:      "Permissions of the file in octal format like `755`. Defaults to the umask of the system.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validate.FileMode(),
			},
			resourceFileAttrUser: {
				Description:   "Name of the user who owns the file",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFileAttrUid},
			},
			resourceFileAttrUid: {
				Description:   "ID of the user who owns the file",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFileAttrUser},
			},
			resourceFileAttrGroup: {
				Description:   "Name of the group that owns the file",
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFileAttrGid},
			},
			resourceFileAttrGid: {
				Description:   "ID of the group that owns the file",
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{resourceFileAttrGroup},
			},
			resourceFileAttrContent: {
				Description: fmt.Sprintf("Content of the file. Only recommended for small text-based payloads such as configuration files etc. The content will be stored in plain-text in the terraform state. Mutually exclusive with attributes `%[2]s` and `%[3]s`.", resourceFileAttrContent, resourceFileAttrContentSensitive, resourceFileAttrSource),
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				ConflictsWith: []string{
					resourceFileAttrContentSensitive,
					resourceFileAttrSource,
				},
			},
			resourceFileAttrContentSensitive: {
				Description: fmt.Sprintf("Content of the file similar to `%[1]s` attribute but with enabled sensitive flag. Prefer `%[2]s` to `%[1]s` to avoid leak of the content in the terraform log output. Mutually exclusive with attributes `%[1]s` and `%[3]s`.", resourceFileAttrContent, resourceFileAttrContentSensitive, resourceFileAttrSource),
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ConflictsWith: []string{
					resourceFileAttrContent,
					resourceFileAttrSource,
				},
			},
			resourceFileAttrSource: {
				Description: fmt.Sprintf("Path to a local file to upload as the file. Mutually exclusive with attributes `%[1]s` and `%[2]s`.", resourceFileAttrContent, resourceFileAttrContentSensitive, resourceFileAttrSource),
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				ForceNew:    true,
				ValidateDiagFunc: func(val interface{}, path cty.Path) diag.Diagnostics {
					valUrl, err := validate.ExpectUrl(val, path)
					if err != nil {
						return err
					}

					// Attempt to open
					s, openErr := sources.OpenUrl(valUrl)
					if openErr != nil {
						return []diag.Diagnostic{
							{
								Severity:      diag.Error,
								Summary:       fmt.Sprintf("failed to open url %q", valUrl.String()),
								Detail:        openErr.Error(),
								AttributePath: path,
							},
						}
					}

					_ = s.Close()

					return nil
				},
				// StateFunc stores the etag of the referenced source in the state in the form etag=[etag]
				StateFunc: func(val interface{}) string {
					// Expect a string
					valStr, valIsStr := val.(string)
					if !valIsStr {
						panic(fmt.Sprintf("[ERROR] StateFunc of attribute `%[2]s` in resource `%[1]s` expects a string but got %+v", resourceFileName, resourceFileAttrSource, val))
					}

					// Pass-through etag
					if strings.HasPrefix(valStr, "etag=") {
						return valStr
					}

					// Get etag from source meta struct
					s, err := sources.Open(valStr)
					if err != nil {
						panic(err)
					}
					defer func() {
						_ = s.Close()
					}()

					m, err := s.Meta()
					if err != nil {
						panic(err)
					}

					stateStr := fmt.Sprintf("etag=%s", m.ETag())

					return stateStr
				},
				ConflictsWith: []string{
					resourceFileAttrContent,
					resourceFileAttrContentSensitive,
				},
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

func resourceFileGetResourceData(sources *source.Registry, d *schema.ResourceData) (*client.File, diag.Diagnostics) {
	r := &client.File{
		Path:    d.Get(resourceFileAttrPath).(string),
		Mode:    0,
		User:    "",
		Uid:     -1,
		Group:   "",
		Gid:     -1,
		Content: nil,
		Source:  nil,
		Md5Sum:  "",
	}

	if d.HasChange(resourceFileAttrMode) {
		r.Mode = filemode.MustParse(d.Get(resourceFileAttrMode).(string))
	}

	if d.HasChange(resourceFileAttrUser) {
		r.User = d.Get(resourceFileAttrUser).(string)
	}

	if d.HasChange(resourceFileAttrUid) {
		r.Uid = intOrDefault(optional(d.GetOk(resourceFileAttrUid)), -1)
	}

	if d.HasChange(resourceFileAttrGroup) {
		r.Group = d.Get(resourceFileAttrGroup).(string)
	}

	if d.HasChange(resourceFileAttrGid) {
		r.Gid = intOrDefault(optional(d.GetOk(resourceFileAttrGid)), -1)
	}

	if d.HasChange(resourceFileAttrContent) {
		content := d.Get(resourceFileAttrContent).(string)
		if content != "" {
			r.Source = bytes.NewReader([]byte(content))
		}
	} else if d.HasChange(resourceFileAttrContentSensitive) {
		content := d.Get(resourceFileAttrContentSensitive).(string)
		if content != "" {
			r.Source = bytes.NewReader([]byte(content))
		}
	} else if d.HasChange(resourceFileAttrSource) {
		sourceUrlStr := d.Get(resourceFileAttrSource).(string)
		s, err := sources.Open(sourceUrlStr)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		r.Source = s
	}

	return r, nil
}

func resourceFileSetResourceData(r *client.File, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceFileAttrPath, r.Path)
	_ = d.Set(resourceFileAttrMode, filemode.Mode(r.Mode).String())
	_ = d.Set(resourceFileAttrUser, r.User)
	_ = d.Set(resourceFileAttrUid, r.Uid)
	_ = d.Set(resourceFileAttrGroup, r.Group)
	_ = d.Set(resourceFileAttrGid, r.Gid)

	_ = d.Set(resourceFileAttrMd5Sum, r.Md5Sum)
	_ = d.Set(resourceFileAttrBasename, path.Base(r.Path))

	if r.Content != nil {
		// Decide whether to store the retrieved content in "content" or "content_sensitive" attribute
		if _, hasContent := d.GetOk(resourceFileAttrContent); hasContent {
			_ = d.Set(resourceFileAttrContent, string(r.Content))
		} else if _, hasContentSensitive := d.GetOk(resourceFileAttrContentSensitive); hasContentSensitive {
			_ = d.Set(resourceFileAttrContentSensitive, string(r.Content))
		} else {
			return newDetailedDiagnostic(diag.Error, "inconsistent configuration", fmt.Sprintf(`cannot decide between "%s" and "%s" attribute`, resourceFileAttrContent, resourceFileAttrContentSensitive), nil)
		}
	} else {
		_ = d.Set(resourceFileAttrContent, nil)
		_ = d.Set(resourceFileAttrContentSensitive, nil)
	}

	return nil
}

func resourceFileCreateFactory(sources *source.Registry) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		p, diagErr := providerFromMeta(meta)
		if diagErr != nil {
			return diagErr
		}

		c := client.NewFileClient(p.System, client.FileClientCompression(true))

		r, diagErr := resourceFileGetResourceData(sources, d)
		if diagErr != nil {
			return diagErr
		}

		err := c.Create(ctx, *r)
		if err != nil {
			return diag.FromErr(err)
		}

		// Close source if source is an io.Closer
		if sourceCloser, isCloser := r.Source.(io.Closer); isCloser {
			err := sourceCloser.Close()
			if err != nil {
				return diag.FromErr(err)
			}
		}

		d.SetId(r.Path)

		diagErr = resourceFileRead(ctx, d, meta)
		if diagErr != nil {
			return diagErr
		}

		return nil
	}
}

func resourceFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	_, hasContent := d.GetOk(resourceFileAttrContent)
	_, hasContentSensitive := d.GetOk(resourceFileAttrContentSensitive)

	// Include content when attributes content or content_sensitive are used
	includeContentOpt := client.FileClientIncludeContent(hasContent || hasContentSensitive)
	c := client.NewFileClient(p.System, includeContentOpt, client.FileClientCompression(true))

	id := d.Id()

	r, err := c.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	diagErr = resourceFileSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceFileUpdateFactory(sources *source.Registry) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		p, diagErr := providerFromMeta(meta)
		if diagErr != nil {
			return diagErr
		}

		c := client.NewFileClient(p.System)

		r, diagErr := resourceFileGetResourceData(sources, d)
		if diagErr != nil {
			return diagErr
		}

		err := c.Update(ctx, *r)
		if err != nil {
			return diag.FromErr(err)
		}

		return resourceFileRead(ctx, d, meta)
	}
}

func resourceFileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewFileClient(p.System)

	id := d.Id()

	err := c.Delete(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
