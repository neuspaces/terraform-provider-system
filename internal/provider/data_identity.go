package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/neuspaces/terraform-provider-system/internal/client"
)

const dataIdentityName = "system_identity"

const (
	dataIdentityAttrUser = "user"

	dataIdentityAttrUid = "uid"

	dataIdentityAttrGroup = "group"

	dataIdentityAttrGid = "gid"
)

func dataIdentity() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` retrieves information about the identity of the user on the remote system.", dataIdentityName),
		ReadContext: dataIdentityRead,
		Schema: map[string]*schema.Schema{
			dataIdentityAttrUser: {
				Description: "Name of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataIdentityAttrUid: {
				Description: "ID of the user.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			dataIdentityAttrGroup: {
				Description: "Name of the primary group of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			dataIdentityAttrGid: {
				Description: "ID of the primary group of the user.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	c := client.NewInfoClient(p.System)

	userInfo, err := c.GetIdentity(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Terraform requires an id: Use the hex encoded sha1 sum of a string concat of all attributes
	id, err := dataIdFromAttrValues(userInfo.Name, userInfo.Uid, userInfo.Group, userInfo.Gid)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	_ = d.Set(dataIdentityAttrUser, userInfo.Name)
	_ = d.Set(dataIdentityAttrUid, userInfo.Uid)
	_ = d.Set(dataIdentityAttrGroup, userInfo.Group)
	_ = d.Set(dataIdentityAttrGid, userInfo.Gid)

	return nil
}
