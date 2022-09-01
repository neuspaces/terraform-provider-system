package etag_test

import (
	"github.com/neuspaces/terraform-provider-system/internal/lib/etag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Desc           string
		Header         string
		Expect         *etag.Header
		ExpectErr      bool
		ExpectExactErr error
	}

	tcs := []testCase{
		{
			Desc:   "valid strong etag",
			Header: `"xyzzy"`,
			Expect: &etag.Header{
				ETag: "xyzzy",
				Weak: false,
			},
		},
		{
			Desc:   "valid weak etag",
			Header: `W/"xyzzy"`,
			Expect: &etag.Header{
				ETag: "xyzzy",
				Weak: true,
			},
		},
		{
			Desc:   "valid empty etag",
			Header: `""`,
			Expect: &etag.Header{
				ETag: "",
				Weak: false,
			},
		},
		{
			Desc:      "invalid: empty header",
			Header:    ``,
			ExpectErr: true,
		},
		{
			Desc:      "invalid: single quote",
			Header:    `xyzzy"`,
			ExpectErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.Desc, func(t *testing.T) {
			actual, err := etag.Parse(tc.Header)

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
		})
	}
}
