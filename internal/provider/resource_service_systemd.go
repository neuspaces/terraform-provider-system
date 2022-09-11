package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
	"regexp"
)

const resourceServiceSystemdName = "system_service_systemd"

const (
	resourceServiceSystemdAttrId            = "id"
	resourceServiceSystemdAttrName          = "name"
	resourceServiceSystemdAttrStatus        = "status"
	resourceServiceSystemdAttrStatusStarted = "started"
	resourceServiceSystemdAttrStatusStopped = "stopped"
	resourceServiceSystemdAttrEnabled       = "enabled"
	resourceServiceSystemdAttrScope         = "scope"
	resourceServiceSystemdAttrScopeSystem   = "system"
	resourceServiceSystemdAttrScopeUser     = "user"
	resourceServiceSystemdAttrScopeGlobal   = "global"
	resourceServiceSystemdAttrRestartOn     = "restart_on"
	resourceServiceSystemdAttrReloadOn      = "reload_on"
)

func resourceServiceSystemd() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages a systemd service on the remote system.", resourceServiceSystemdName),

		CreateContext: resourceServiceSystemdCreate,
		ReadContext:   resourceServiceSystemdRead,
		UpdateContext: resourceServiceSystemdUpdate,
		DeleteContext: resourceServiceSystemdDelete,

		// Importer is not required; resourceServiceSystemdRead does not fail when service does not exist; idempotent create in resourceServiceSystemdCreate;

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceServiceSystemdAttrId: {
				Description: "ID of the service",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceServiceSystemdAttrName: {
				Description:  "Name of the service without the suffix `.service`. The service unit must exist.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.service$`), "name of the service must not have the suffix `.service`"),
			},
			resourceServiceSystemdAttrStatus: {
				Description: fmt.Sprintf("Status of the service. If `%[1]s`, the service will be started. If `%[2]s`, the service will be stopped.", resourceServiceSystemdAttrStatusStarted, resourceServiceSystemdAttrStatusStopped),
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice([]string{
					resourceServiceSystemdAttrStatusStarted,
					resourceServiceSystemdAttrStatusStopped,
				}, false),
			},
			resourceServiceSystemdAttrEnabled: {
				Description: "If `true`, the service will be enabled. If not provided, the service will not be changed.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			resourceServiceSystemdAttrScope: {
				Description: fmt.Sprintf("Scope in which the service is managed. In the current iteration, the only supported scope is `%[1]s`. In future iterations, the scopes `%[2]s` and `%[3]s` may be added. Defaults to `%[1]s`", resourceServiceSystemdAttrScopeSystem, resourceServiceSystemdAttrScopeUser, resourceServiceSystemdAttrScopeGlobal),
				Type:        schema.TypeString,
				Optional:    true,
				Default:     resourceServiceSystemdAttrScopeSystem,
				ValidateFunc: validation.StringInSlice([]string{
					resourceServiceSystemdAttrScopeSystem,
					// resourceServiceSystemdAttrScopeUser,
					// resourceServiceSystemdAttrScopeGlobal,
				}, false),
			},
			resourceServiceSystemdAttrRestartOn: {
				Description: "Set of arbitrary strings which will trigger a restart of the service.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			resourceServiceSystemdAttrReloadOn: {
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

type resourceServiceSystemdInternalData struct {
	// PreStatus is the original status of the service before managed by the resource. This status will be applied when the resource is destroyed.
	PreStatus string `json:"pre_status,omitempty"`

	// PreEnabled is true if the service was enabled before managed by the resource. This activation will be applied when the resource is destroyed.
	PreEnabled *bool `json:"pre_enabled,omitempty"`
}

func resourceServiceSystemdGetResourceData(d *schema.ResourceData) (*client.Service, diag.Diagnostics) {
	r := &client.Service{
		Name: d.Get(resourceServiceSystemdAttrName).(string),
	}

	if val, ok := d.GetOk(resourceServiceSystemdAttrStatus); ok {
		r.Status = client.ServiceStatusPtr(resourceServiceSystemdStatusToClientStatus(val.(string)))
	}

	// Use deprecated GetOkExists instead of HasChange because HasChange does not support optional bool attributes
	// https://github.com/hashicorp/terraform-plugin-sdk/issues/817
	if val, exists := d.GetOkExists(resourceServiceSystemdAttrEnabled); exists {
		r.Enabled = to.BoolPtr(val.(bool))
	}

	return r, nil
}

func resourceServiceSystemdSetResourceData(r *client.Service, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceServiceSystemdAttrName, r.Name)

	if r.Status != nil {
		_ = d.Set(resourceServiceSystemdAttrStatus, resourceServiceSystemdStatusFromClientStatus(*r.Status))
	}

	if r.Enabled != nil {
		_ = d.Set(resourceServiceSystemdAttrEnabled, to.Bool(r.Enabled))
	}

	return nil
}

func resourceServiceSystemdNewClient(ctx context.Context, meta interface{}) (client.ServiceClient, diag.Diagnostics) {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return nil, diagErr
	}

	c := client.NewSystemdServiceClient(p.System)

	return c, nil
}

func resourceServiceSystemdCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceSystemdNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceSystemdGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	preR, err := resourceServiceClientGet(ctx, c, client.ServiceGetArgs{Name: r.Name})
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
	if d.HasChange(resourceServiceSystemdAttrReloadOn) {
		applyOpts = append(applyOpts, client.ServiceReload())
	}

	// Handle restart trigger
	if d.HasChange(resourceServiceSystemdAttrRestartOn) {
		applyOpts = append(applyOpts, client.ServiceRestart())
	}

	err = c.Apply(ctx, *r, applyOpts...)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Name)

	// Store status and activation before create in internal data
	internalData := resourceServiceSystemdInternalData{
		PreStatus:  resourceServiceSystemdStatusFromClientStatus(*preR.Status),
		PreEnabled: preR.Enabled,
	}

	diagErr = setInternalData(d, &internalData)
	if diagErr != nil {
		return diagErr
	}

	return resourceServiceSystemdRead(ctx, d, meta)
}

func resourceServiceSystemdRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceSystemdNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceSystemdGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	r, err := resourceServiceClientGet(ctx, c, client.ServiceGetArgs{Name: r.Name})
	if err != nil {
		if errors.Is(err, client.ErrServiceNotFound) {
			// Ignore if service does not exist in get
		} else {
			return diag.FromErr(err)
		}
	}

	diagErr = resourceServiceSystemdSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceServiceSystemdUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceSystemdNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceSystemdGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	// Apply options
	var applyOpts []client.ServiceApplyOption

	// Handle reload trigger
	if d.HasChange(resourceServiceSystemdAttrReloadOn) {
		applyOpts = append(applyOpts, client.ServiceReload())
	}

	// Handle restart trigger
	if d.HasChange(resourceServiceSystemdAttrRestartOn) {
		applyOpts = append(applyOpts, client.ServiceRestart())
	}

	err := c.Apply(ctx, *r, applyOpts...)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceSystemdRead(ctx, d, meta)
}

func resourceServiceSystemdDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceServiceSystemdNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceServiceSystemdGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	// Apply the service status and activation to values before the resource has been created
	preR := &client.Service{
		Name: r.Name,
	}

	var internalData resourceServiceSystemdInternalData
	_, diagErr = getInternalData(d, &internalData)
	if diagErr != nil {
		return diagErr
	}

	if internalData.PreStatus != "" {
		preR.Status = client.ServiceStatusPtr(resourceServiceSystemdStatusToClientStatus(internalData.PreStatus))
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

func resourceServiceSystemdStatusToClientStatus(s string) client.ServiceStatus {
	switch s {
	case resourceServiceSystemdAttrStatusStarted:
		return client.ServiceStatusStarted
	case resourceServiceSystemdAttrStatusStopped:
		return client.ServiceStatusStopped
	}

	return client.ServiceStatusUndefined
}

func resourceServiceSystemdStatusFromClientStatus(s client.ServiceStatus) string {
	switch s {
	case client.ServiceStatusStarted:
		return resourceServiceSystemdAttrStatusStarted
	case client.ServiceStatusStopped:
		return resourceServiceSystemdAttrStatusStopped
	}

	return ""
}
