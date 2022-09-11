package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/client/openrc"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
)

const resourceServiceOpenrcName = "system_service_openrc"

const (
	resourceServiceOpenrcAttrId            = "id"
	resourceServiceOpenrcAttrName          = "name"
	resourceServiceOpenrcAttrStatus        = "status"
	resourceServiceOpenrcAttrStatusStarted = "started"
	resourceServiceOpenrcAttrStatusStopped = "stopped"
	resourceServiceOpenrcAttrEnabled       = "enabled"
	resourceServiceOpenrcAttrRunlevel      = "runlevel"
	resourceServiceOpenrcAttrRestartOn     = "restart_on"
	resourceServiceOpenrcAttrReloadOn      = "reload_on"
)

func resourceServiceOpenrc() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages an OpenRC service on the remote system.", resourceServiceOpenrcName),

		CreateContext: resourceServiceOpenrcCreate,
		ReadContext:   resourceServiceOpenrcRead,
		UpdateContext: resourceServiceOpenrcUpdate,
		DeleteContext: resourceServiceOpenrcDelete,

		// Importer is intentionally not configured
		// Read will not fail if the service does not exist
		// Create will implicitly import the service in the state

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceServiceOpenrcAttrId: {
				Description: "ID of the service",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceServiceOpenrcAttrName: {
				Description: "Name of the service. The service must exist.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			resourceServiceOpenrcAttrStatus: {
				Description: fmt.Sprintf("Status of the service. If `%[1]s`, the service will be started. If `%[2]s`, the service will be stopped.", resourceServiceOpenrcAttrStatusStarted, resourceServiceOpenrcAttrStatusStopped),
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice([]string{
					resourceServiceOpenrcAttrStatusStarted,
					resourceServiceOpenrcAttrStatusStopped,
				}, false),
			},
			resourceServiceOpenrcAttrEnabled: {
				Description: "If `true`, the service will be enabled on the provided runlevel. If not provided, the service will not be changed.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			resourceServiceOpenrcAttrRunlevel: {
				Description: fmt.Sprintf("Runlevel to which the `enabled` attribute refers to. Defaults to `%[1]s`.", openrc.DefaultRunlevel),
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				DefaultFunc: func() (interface{}, error) {
					return openrc.DefaultRunlevel, nil
				},
			},
			resourceServiceOpenrcAttrRestartOn: {
				Description: "Set of arbitrary strings which will trigger a restart of the service.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			resourceServiceOpenrcAttrReloadOn: {
				Description: "Set of arbitrary strings which will trigger a reload of the service.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			internalDataSchemaKey: internalDataSchema(),
		},
	}
}

type resourceServiceOpenrcInternalData struct {
	// PreStatus is the original status of the service before managed by the resource. This status will be applied when the resource is destroyed.
	PreStatus string `json:"pre_status,omitempty"`

	// PreEnabled is true if the service was enabled before managed by the resource. This activation will be applied when the resource is destroyed.
	PreEnabled *bool `json:"pre_enabled,omitempty"`
}

func resourceServiceOpenrcGetResourceData(d *schema.ResourceData) (*client.Service, diag.Diagnostics) {
	r := &client.Service{
		Name:     d.Get(resourceServiceOpenrcAttrName).(string),
		Runlevel: d.Get(resourceServiceOpenrcAttrRunlevel).(string),
	}

	if val, ok := d.GetOk(resourceServiceOpenrcAttrStatus); ok {
		r.Status = client.ServiceStatusPtr(resourceServiceOpenRcStatusToClientStatus(val.(string)))
	}

	// Use deprecated GetOkExists instead of HasChange because HasChange does not support optional bool attributes
	// https://github.com/hashicorp/terraform-plugin-sdk/issues/817
	if val, ok := d.GetOkExists(resourceServiceOpenrcAttrEnabled); ok {
		r.Enabled = to.BoolPtr(val.(bool))
	}

	return r, nil
}

func resourceServiceOpenrcSetResourceData(r *client.Service, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceServiceOpenrcAttrName, r.Name)

	if r.Status != nil {
		_ = d.Set(resourceServiceOpenrcAttrStatus, resourceServiceOpenRcStatusFromClientStatus(*r.Status))
	}

	if r.Enabled != nil {
		_ = d.Set(resourceServiceOpenrcAttrEnabled, to.Bool(r.Enabled))
	}

	_ = d.Set(resourceServiceOpenrcAttrRunlevel, r.Runlevel)

	return nil
}

func resourceServiceOpenrcNewClient(ctx context.Context, meta interface{}) (client.ServiceClient, diag.Diagnostics) {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return nil, diagErr
	}

	c := client.NewOpenRcServiceClient(p.System)

	return c, nil
}

func resourceServiceOpenrcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceOpenrcNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceOpenrcGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	preR, err := resourceServiceClientGet(ctx, c, client.ServiceGetArgs{Name: r.Name, Runlevel: r.Runlevel})
	if err != nil {
		return diag.FromErr(err)
	}

	// Require enabled and status properties
	if preR.Enabled == nil {
		return newDetailedDiagnostic(diag.Error, "unexpected enabled property", "enabled property could not be determined during create", nil)
	}

	if preR.Status == nil {
		return newDetailedDiagnostic(diag.Error, "unexpected status property", "status property could not be determined during create", nil)
	}

	// Apply options
	var applyOpts []client.ServiceApplyOption

	// Handle reload trigger
	if d.HasChange(resourceServiceOpenrcAttrReloadOn) {
		applyOpts = append(applyOpts, client.ServiceReload())
	}

	// Handle restart trigger
	if d.HasChange(resourceServiceOpenrcAttrRestartOn) {
		applyOpts = append(applyOpts, client.ServiceRestart())
	}

	err = c.Apply(ctx, *r, applyOpts...)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Name)

	// Store status and activation before create in internal data
	internalData := resourceServiceOpenrcInternalData{
		PreStatus:  resourceServiceOpenRcStatusFromClientStatus(*preR.Status),
		PreEnabled: preR.Enabled,
	}

	diagErr = setInternalData(d, &internalData)
	if diagErr != nil {
		return diagErr
	}

	return resourceServiceOpenrcRead(ctx, d, meta)
}

func resourceServiceOpenrcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceOpenrcNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceOpenrcGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	r, err := resourceServiceClientGet(ctx, c, client.ServiceGetArgs{Name: r.Name, Runlevel: r.Runlevel})
	if err != nil {
		if errors.Is(err, client.ErrServiceNotFound) {
			// Ignore if service does not exist in read operation
		} else {
			return diag.FromErr(err)
		}
	}

	diagErr = resourceServiceOpenrcSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceServiceOpenrcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceOpenrcNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceOpenrcGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	// Update of `runlevel` is handled by ForceNew

	// Apply options
	var applyOpts []client.ServiceApplyOption

	// Handle reload trigger
	if d.HasChange(resourceServiceOpenrcAttrReloadOn) {
		applyOpts = append(applyOpts, client.ServiceReload())
	}

	// Handle restart trigger
	if d.HasChange(resourceServiceOpenrcAttrRestartOn) {
		applyOpts = append(applyOpts, client.ServiceRestart())
	}

	err := c.Apply(ctx, *r, applyOpts...)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceOpenrcRead(ctx, d, meta)
}

func resourceServiceOpenrcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceOpenrcNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceOpenrcGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	// Apply the service status and activation to values before the resource has been created
	preR := &client.Service{
		Name: r.Name,
		// Assume the runlevel attribute has not changed due to the ForceNew flag
		Runlevel: r.Runlevel,
	}

	var internalData resourceServiceOpenrcInternalData
	_, diagErr = getInternalData(d, &internalData)
	if diagErr != nil {
		return diagErr
	}

	if internalData.PreStatus != "" {
		preR.Status = client.ServiceStatusPtr(resourceServiceOpenRcStatusToClientStatus(internalData.PreStatus))
	}

	if internalData.PreEnabled != nil {
		preR.Enabled = internalData.PreEnabled
	}

	err := c.Apply(ctx, *preR)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceOpenRcStatusToClientStatus(s string) client.ServiceStatus {
	switch s {
	case resourceServiceOpenrcAttrStatusStarted:
		return client.ServiceStatusStarted
	case resourceServiceOpenrcAttrStatusStopped:
		return client.ServiceStatusStopped
	}

	return client.ServiceStatusUndefined
}

func resourceServiceOpenRcStatusFromClientStatus(s client.ServiceStatus) string {
	switch s {
	case client.ServiceStatusStarted:
		return resourceServiceOpenrcAttrStatusStarted
	case client.ServiceStatusStopped:
		return resourceServiceOpenrcAttrStatusStopped
	}

	return ""
}
