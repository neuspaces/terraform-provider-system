package provider

import (
	"encoding/base64"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const internalDataSchemaKey = "internal"

func internalDataSchema() *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeString,
		Computed:  true,
		Default:   nil,
		Sensitive: true,
	}
}

// getInternalData returns internal data stored in the resource state
func getInternalData(d *schema.ResourceData, v interface{}) (bool, diag.Diagnostics) {
	internalData, hasInternalData := d.GetOk(internalDataSchemaKey)
	if !hasInternalData {
		return false, nil
	}

	internalEncoded, ok := internalData.(string)
	if !ok {
		return false, newShortDiagnostic(diag.Error, "failed to decode internal resource data")
	}

	internalJson, err := base64.StdEncoding.DecodeString(internalEncoded)
	if err != nil {
		return false, newDetailedDiagnostic(diag.Error, "failed to decode internal resource data", err.Error(), nil)
	}

	err = json.Unmarshal(internalJson, v)
	if err != nil {
		return false, newDetailedDiagnostic(diag.Error, "failed to unmarshal internal resource data", err.Error(), nil)
	}

	return true, nil
}

func setInternalData(d *schema.ResourceData, v interface{}) diag.Diagnostics {
	internalJson, err := json.Marshal(v)
	if err != nil {
		return newDetailedDiagnostic(diag.Error, "failed to marshal internal resource data", err.Error(), nil)
	}

	internalEncoded := base64.StdEncoding.EncodeToString(internalJson)

	err = d.Set(internalDataSchemaKey, internalEncoded)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
