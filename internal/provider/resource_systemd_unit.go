package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"github.com/neuspaces/terraform-provider-system/internal/client/systemd"
	"github.com/neuspaces/terraform-provider-system/internal/extlib/to"
	"github.com/sethvargo/go-retry"
	"time"
)

const resourceSystemdUnitName = "system_systemd_unit"

const (
	resourceSystemdUnitAttrId            = "id"
	resourceSystemdUnitAttrType          = "type"
	resourceSystemdUnitAttrName          = "name"
	resourceSystemdUnitAttrStatus        = "status"
	resourceSystemdUnitAttrStatusStarted = "started"
	resourceSystemdUnitAttrStatusStopped = "stopped"
	resourceSystemdUnitAttrEnabled       = "enabled"
	resourceSystemdUnitAttrScope         = "scope"
	resourceSystemdUnitAttrScopeSystem   = "system"
	resourceSystemdUnitAttrScopeUser     = "user"
	resourceSystemdUnitAttrScopeGlobal   = "global"
	resourceSystemdUnitAttrRestartOn     = "restart_on"
	resourceSystemdUnitAttrReloadOn      = "reload_on"
)

func resourceSystemdUnit() *schema.Resource {
	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages a systemd unit on the remote system.", resourceSystemdUnitName),

		CreateContext: resourceSystemdUnitCreate,
		ReadContext:   resourceSystemdUnitRead,
		UpdateContext: resourceSystemdUnitUpdate,
		DeleteContext: resourceSystemdUnitDelete,

		// Importer is not required; resourceSystemdUnitRead does not fail when systemd unit does not exist; idempotent create in resourceSystemdUnitCreate;

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourceSystemdUnitAttrId: {
				Description: "ID of the systemd unit",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourceSystemdUnitAttrType: {
				Description: fmt.Sprintf("Unit type. Supported unit types are `%[1]s`, `%[2]s`, `%[3]s`, `%[4]s`, `%[5]s`, `%[6]s`, `%[7]s`, `%[8]s`, `%[9]s`, `%[10]s`, and `%[11]s`.",
					string(systemd.UnitTypeService),
					string(systemd.UnitTypeSocket),
					string(systemd.UnitTypeDevice),
					string(systemd.UnitTypeMount),
					string(systemd.UnitTypeAutoMount),
					string(systemd.UnitTypeSwap),
					string(systemd.UnitTypeTarget),
					string(systemd.UnitTypePath),
					string(systemd.UnitTypeTimer),
					string(systemd.UnitTypeSlide),
					string(systemd.UnitTypeScope),
				),
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(systemd.UnitTypeService),
					string(systemd.UnitTypeSocket),
					string(systemd.UnitTypeDevice),
					string(systemd.UnitTypeMount),
					string(systemd.UnitTypeAutoMount),
					string(systemd.UnitTypeSwap),
					string(systemd.UnitTypeTarget),
					string(systemd.UnitTypePath),
					string(systemd.UnitTypeTimer),
					string(systemd.UnitTypeSlide),
					string(systemd.UnitTypeScope),
				}, false),
			},
			resourceSystemdUnitAttrName: {
				Description: "Name of the unit without the type suffix. The unit must exist.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			resourceSystemdUnitAttrStatus: {
				Description: fmt.Sprintf("Status of the unit. If `%[1]s`, the unit will be started. If `%[2]s`, the unit will be stopped.", resourceSystemdUnitAttrStatusStarted, resourceSystemdUnitAttrStatusStopped),
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice([]string{
					resourceSystemdUnitAttrStatusStarted,
					resourceSystemdUnitAttrStatusStopped,
				}, false),
			},
			resourceSystemdUnitAttrEnabled: {
				Description: "If `true`, the unit will be enabled. If not provided, the unit will not be changed.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Default:     nil,
			},
			resourceSystemdUnitAttrScope: {
				Description: fmt.Sprintf("Scope in which the unit is managed. In the current iteration, the only supported scope is `%[1]s`. In future iterations, the scopes `%[2]s` and `%[3]s` may be added. Defaults to `%[1]s`", resourceSystemdUnitAttrScopeSystem, resourceSystemdUnitAttrScopeUser, resourceSystemdUnitAttrScopeGlobal),
				Type:        schema.TypeString,
				Optional:    true,
				Default:     resourceSystemdUnitAttrScopeSystem,
				ValidateFunc: validation.StringInSlice([]string{
					resourceSystemdUnitAttrScopeSystem,
					// resourceSystemdUnitAttrScopeUser,
					// resourceSystemdUnitAttrScopeGlobal,
				}, false),
			},
			resourceSystemdUnitAttrRestartOn: {
				Description: "Set of arbitrary strings which when changed will trigger a restart of the unit when changed.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			resourceSystemdUnitAttrReloadOn: {
				Description: "Set of arbitrary strings which when changed will trigger a reload of the unit.",
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

type resourceSystemdUnitInternalData struct {
	// PreStatus is the original status of the unit before managed by the resource. This status will be applied when the resource is destroyed.
	PreStatus string `json:"pre_status,omitempty"`

	// PreEnabled is true if the unit was enabled before managed by the resource. This activation will be applied when the resource is destroyed.
	PreEnabled *bool `json:"pre_enabled,omitempty"`
}

func resourceSystemdUnitGetResourceData(d *schema.ResourceData) (*client.SystemdUnit, diag.Diagnostics) {
	r := &client.SystemdUnit{
		Type: d.Get(resourceSystemdUnitAttrType).(string),
		Name: d.Get(resourceSystemdUnitAttrName).(string),
	}

	if val, ok := d.GetOk(resourceSystemdUnitAttrStatus); ok {
		r.Status = client.SystemdUnitStatusPtr(resourceSystemdUnitStatusToClientStatus(val.(string)))
	}

	// Use deprecated GetOkExists instead of HasChange because HasChange does not support optional bool attributes
	// https://github.com/hashicorp/terraform-plugin-sdk/issues/817
	if val, exists := d.GetOkExists(resourceSystemdUnitAttrEnabled); exists {
		r.Enabled = to.BoolPtr(val.(bool))
	}

	return r, nil
}

func resourceSystemdUnitSetResourceData(r *client.SystemdUnit, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourceSystemdUnitAttrType, r.Type)
	_ = d.Set(resourceSystemdUnitAttrName, r.Name)

	if r.Status != nil {
		_ = d.Set(resourceSystemdUnitAttrStatus, resourceSystemdUnitStatusFromClientStatus(*r.Status))
	}

	if r.Enabled != nil {
		_ = d.Set(resourceSystemdUnitAttrEnabled, to.Bool(r.Enabled))
	}

	return nil
}

func resourceSystemdUnitNewClient(ctx context.Context, meta interface{}) (client.SystemdUnitClient, diag.Diagnostics) {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return nil, diagErr
	}

	c := client.NewSystemdUnitClient(p.System)

	return c, nil
}

func resourceSystemdUnitGet(ctx context.Context, c client.SystemdUnitClient, args client.SystemdUnitGetArgs) (*client.SystemdUnit, error) {
	var r *client.SystemdUnit

	err := retry.Do(ctx, retry.NewConstant(5*time.Second), func(ctx context.Context) error {
		var err error

		r, err = c.Get(ctx, args)
		if err != nil {
			return err
		}

		if r.Status != nil && r.Status.IsPending() {
			return retry.RetryableError(fmt.Errorf("pending unit status: %s", *r.Status))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func resourceSystemdUnitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceSystemdUnitNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceSystemdUnitGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	preR, err := resourceSystemdUnitGet(ctx, c, client.SystemdUnitGetArgs{Type: r.Type, Name: r.Name})
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
	var applyOpts []client.SystemdUnitApplyOption

	// Handle reload trigger
	if d.HasChange(resourceSystemdUnitAttrReloadOn) {
		applyOpts = append(applyOpts, client.SystemdUnitReload())
	}

	// Handle restart trigger
	if d.HasChange(resourceSystemdUnitAttrRestartOn) {
		applyOpts = append(applyOpts, client.SystemdUnitRestart())
	}

	err = c.Apply(ctx, *r, applyOpts...)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Name)

	// Store status and activation before create in internal data
	internalData := resourceSystemdUnitInternalData{
		PreStatus:  resourceSystemdUnitStatusFromClientStatus(*preR.Status),
		PreEnabled: preR.Enabled,
	}

	diagErr = setInternalData(d, &internalData)
	if diagErr != nil {
		return diagErr
	}

	return resourceSystemdUnitRead(ctx, d, meta)
}

func resourceSystemdUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceSystemdUnitNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceSystemdUnitGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	r, err := resourceSystemdUnitGet(ctx, c, client.SystemdUnitGetArgs{Type: r.Type, Name: r.Name})
	if err != nil {
		if errors.Is(err, client.ErrSystemdUnitNotFound) {
			// Ignore if unit does not exist in get phase
		} else {
			return diag.FromErr(err)
		}
	}

	diagErr = resourceSystemdUnitSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceSystemdUnitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceSystemdUnitNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceSystemdUnitGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	// Apply options
	var applyOpts []client.SystemdUnitApplyOption

	// Handle reload trigger
	if d.HasChange(resourceSystemdUnitAttrReloadOn) {
		applyOpts = append(applyOpts, client.SystemdUnitReload())
	}

	// Handle restart trigger
	if d.HasChange(resourceSystemdUnitAttrRestartOn) {
		applyOpts = append(applyOpts, client.SystemdUnitRestart())
	}

	err := c.Apply(ctx, *r, applyOpts...)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSystemdUnitRead(ctx, d, meta)
}

func resourceSystemdUnitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourceSystemdUnitNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	r, diagErr := resourceSystemdUnitGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	// Apply the unit status and activation to values before the resource has been created
	preR := &client.SystemdUnit{
		Type: r.Type,
		Name: r.Name,
	}

	var internalData resourceSystemdUnitInternalData
	_, diagErr = getInternalData(d, &internalData)
	if diagErr != nil {
		return diagErr
	}

	if internalData.PreStatus != "" {
		preR.Status = client.SystemdUnitStatusPtr(resourceSystemdUnitStatusToClientStatus(internalData.PreStatus))
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

func resourceSystemdUnitStatusToClientStatus(s string) client.SystemdUnitStatus {
	switch s {
	case resourceSystemdUnitAttrStatusStarted:
		return client.SystemdUnitStatusStarted
	case resourceSystemdUnitAttrStatusStopped:
		return client.SystemdUnitStatusStopped
	}

	return client.SystemdUnitStatusUndefined
}

func resourceSystemdUnitStatusFromClientStatus(s client.SystemdUnitStatus) string {
	switch s {
	case client.SystemdUnitStatusStarted:
		return resourceSystemdUnitAttrStatusStarted
	case client.SystemdUnitStatusStopped:
		return resourceSystemdUnitAttrStatusStopped
	}

	return ""
}
