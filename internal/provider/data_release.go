package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
)

const dataReleaseName = "system_release"

const (
	dataReleaseAttrName    = "name"
	dataReleaseAttrVendor  = "vendor"
	dataReleaseAttrVersion = "version"
	dataReleaseAttrRelease = "release"
)

func dataRelease() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` retrieves information about the release and distribution of operating system.", dataReleaseName),
		ReadContext: dataReleaseRead,
		Schema: map[string]*schema.Schema{
			dataReleaseAttrName: {
				Description: "Name of the operating system distribution. The value is derived from variable `PRETTY_NAME` in file `/etc/os-release`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataReleaseAttrVendor: {
				Description: "Vendor of the operating system distribution. The value is derived from variable `ID` in file `/etc/os-release`. Example values are `alpine`, `debian`, `ubuntu`, `centos`, `rhel`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataReleaseAttrVersion: {
				Description: "Version of the operating system distribution.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataReleaseAttrRelease: {
				Description: "Vendor-specific release of the operating system distribution.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataReleaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewInfoClient(p.System)

	osInfo, err := c.GetRelease(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Terraform requires an id: Use the hex encoded sha1 sum of a string concat of all attributes
	id, err := dataIdFromAttrValues(osInfo.Name, osInfo.Vendor, osInfo.Version, osInfo.Release)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	_ = d.Set(dataReleaseAttrName, osInfo.Name)
	_ = d.Set(dataReleaseAttrVendor, osInfo.Vendor)
	_ = d.Set(dataReleaseAttrVersion, osInfo.Version)
	_ = d.Set(dataReleaseAttrRelease, osInfo.Release)

	return nil
}
