package local_test

import (
	"bytes"
	"context"
	"github.com/neuspaces/terraform-provider-system/internal/cmd"
	"github.com/neuspaces/terraform-provider-system/internal/system/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSystem(t *testing.T) {
	s := local.NewSystem()

	stdout := &bytes.Buffer{}
	cmd := cmd.NewCommand(`whoami`, cmd.Stdout(stdout))

	ctx := context.Background()
	result, err := s.Execute(ctx, cmd)
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode())

	stdoutBytes := stdout.Bytes()
	assert.NotEmpty(t, stdoutBytes)
}
