package provider

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

// TestCheckResourceAttrBase64 is a TestCheckFunc which validates
// the base64 encoded value in state for the given name/key combination.
func TestCheckResourceAttrBase64(name, key, decodedValue string) resource.TestCheckFunc {
	encodedValue := base64.StdEncoding.EncodeToString([]byte(decodedValue))
	return resource.TestCheckResourceAttr(name, key, encodedValue)
}

// TestLogResourceAttr returns a resource.TestCheckFunc which logs the attributes of the defined resource
// Usage: TestLogResourceAttr(t, "system_group.test")
func TestLogResourceAttr(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()

		is, err := modulePrimaryInstanceState(s, ms, name)
		if err != nil {
			return err
		}

		// String() returns the resource state in multiple lines
		resourceState := is.String()

		// Log should print a single line
		resourceState = strings.Join(strings.Split(strings.TrimSpace(resourceState), "\n"), ",")
		resourceState = fmt.Sprintf("%s{%s}", name, resourceState)

		t.Log(resourceState)

		return nil
	}
}

// TestLogString logs the provided s before returning it
func TestLogString(t *testing.T, s string) string {
	t.Log(s)
	return s
}

// modulePrimaryInstanceState returns the instance state for the given resource
// name in a ModuleState
func modulePrimaryInstanceState(s *terraform.State, ms *terraform.ModuleState, name string) (*terraform.InstanceState, error) {
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s in %s", name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
	}

	return is, nil
}
