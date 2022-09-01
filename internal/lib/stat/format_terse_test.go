package stat_test

import (
	"github.com/neuspaces/terraform-provider-system/internal/lib/stat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseTerseFormat(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Stat           string
		Name           string
		Desc           string
		Expect         *stat.Stat
		ExpectErr      bool
		ExpectExactErr error
		Assert         func(t *testing.T, s *stat.Stat)
	}

	tcs := []testCase{
		{
			Desc: "regular file",
			Name: "/root/regular-file",
			Stat: "/root/regular-file 6 8 81a4 0 0 9e 3416818 1 0 0 1628450121 1628448200 1628448202 4096",
			Assert: func(t *testing.T, s *stat.Stat) {
				assert.Equal(t, "/root/regular-file", s.Name)
				assert.Equal(t, int64(6), s.Size)
				assert.Equal(t, true, s.Mode.IsRegular())
				assert.Equal(t, false, s.Mode.IsDir())
				assert.Equal(t, "-rw-r--r--", s.Mode.ToFsFileMode().Perm().String())
				assert.Equal(t, "2021-08-08 18:43:20 +0000 UTC", s.ModifiedTime.UTC().String())
			},
			ExpectErr:      false,
			ExpectExactErr: nil,
		},
		{
			Desc: "regular file with spaces",
			Name: "/root/file with spaces.txt",
			Stat: "/root/file with spaces.txt 14 8 81a4 0 0 9a 1050294 1 0 0 1628888621 1628888616 1628888618 4096",
			Assert: func(t *testing.T, s *stat.Stat) {
				assert.Equal(t, "/root/file with spaces.txt", s.Name)
				assert.Equal(t, int64(14), s.Size)
				assert.Equal(t, true, s.Mode.IsRegular())
				assert.Equal(t, false, s.Mode.IsDir())
				assert.Equal(t, "-rw-r--r--", s.Mode.ToFsFileMode().Perm().String())
				assert.Equal(t, "2021-08-13 21:03:36 +0000 UTC", s.ModifiedTime.UTC().String())
			},
			ExpectErr:      false,
			ExpectExactErr: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.Desc, func(t *testing.T) {
			actual, err := stat.ParseTerseFormat([]byte(tc.Stat))

			if tc.ExpectExactErr != nil {
				assert.Equal(t, tc.ExpectExactErr, err)
			} else if tc.ExpectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.Expect != nil {
				assert.Equal(t, tc.Expect, actual)
			}

			if tc.Assert != nil {
				tc.Assert(t, actual)
			}
		})
	}
}
