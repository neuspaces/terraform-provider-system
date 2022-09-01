package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func intOrDefault(val interface{}, defaultVal int) int {
	if intVal, ok := val.(int); ok {
		return intVal
	}
	return defaultVal
}

func optional(val interface{}, ok bool) interface{} {
	if ok {
		return val
	}
	return nil
}

// expandListSingle expects a value of type schema.TypeList which has a single element and has been retrieved from
// schema.ResourceData and returns the attributes as a map[string]interface{}
func expandListSingle(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return nil, fmt.Errorf("expected []interface{}, got nil")
	}

	l, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected []interface{}, got unexpected type %T", v)
	}

	if len(l) != 1 {
		return nil, fmt.Errorf("expected single eleement, got %d", len(l))
	}

	d, ok := l[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{}, got unexpected type %T", v)
	}

	return d, nil
}

func schemaEnvDefaultFunc(schemaKey string, prefix string, dv interface{}) schema.SchemaDefaultFunc {
	return schema.EnvDefaultFunc(prefix+strings.ToUpper(schemaKey), dv)
}

type attrPath []string

func (p attrPath) String() string {
	return strings.Join(p, ".")
}

func (p attrPath) Extend(parts ...string) attrPath {
	if p == nil {
		return newAttrPath(parts...)
	}

	return append(p, parts...)
}

func newAttrPath(parts ...string) attrPath {
	return parts
}
