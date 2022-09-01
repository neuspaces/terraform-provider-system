package osrelease_test

import (
	"github.com/neuspaces/terraform-provider-system/internal/extlib/heredoc"
	"github.com/neuspaces/terraform-provider-system/internal/lib/osrelease"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Desc           string
		Value          string
		Expect         *osrelease.Info
		ExpectErr      bool
		ExpectExactErr error
	}

	tcs := []testCase{
		{
			Desc: "alpine 3.14.1",
			Value: heredoc.String(`
				NAME="Alpine Linux"
				ID=alpine
				VERSION_ID=3.14.1
				PRETTY_NAME="Alpine Linux v3.14"
				HOME_URL="https://alpinelinux.org/"
				BUG_REPORT_URL="https://bugs.alpinelinux.org/"
			`),
			Expect: &osrelease.Info{
				Name:       "Alpine Linux",
				Id:         "alpine",
				PrettyName: "Alpine Linux v3.14",
				Version:    "",
				VersionId:  "3.14.1",
			},
		},
		{
			Desc: "debian bullseye",
			Value: heredoc.String(`
				PRETTY_NAME="Debian GNU/Linux 11 (bullseye)"
				NAME="Debian GNU/Linux"
				VERSION_ID="11"
				VERSION="11 (bullseye)"
				VERSION_CODENAME=bullseye
				ID=debian
				HOME_URL="https://www.debian.org/"
				SUPPORT_URL="https://www.debian.org/support"
				BUG_REPORT_URL="https://bugs.debian.org/"
			`),
			Expect: &osrelease.Info{
				Name:       "Debian GNU/Linux",
				Id:         "debian",
				PrettyName: "Debian GNU/Linux 11 (bullseye)",
				Version:    "11 (bullseye)",
				VersionId:  "11",
			},
		},
		{
			Desc: "fedora 35",
			Value: heredoc.String(`
				NAME="Fedora Linux"
				VERSION="35 (Container Image)"
				ID=fedora
				VERSION_ID=35
				VERSION_CODENAME=""
				PLATFORM_ID="platform:f35"
				PRETTY_NAME="Fedora Linux 35 (Container Image)"
				ANSI_COLOR="0;38;2;60;110;180"
				LOGO=fedora-logo-icon
				CPE_NAME="cpe:/o:fedoraproject:fedora:35"
				HOME_URL="https://fedoraproject.org/"
				DOCUMENTATION_URL="https://docs.fedoraproject.org/en-US/fedora/f35/system-administrators-guide/"
				SUPPORT_URL="https://ask.fedoraproject.org/"
				BUG_REPORT_URL="https://bugzilla.redhat.com/"
				REDHAT_BUGZILLA_PRODUCT="Fedora"
				REDHAT_BUGZILLA_PRODUCT_VERSION=35
				REDHAT_SUPPORT_PRODUCT="Fedora"
				REDHAT_SUPPORT_PRODUCT_VERSION=35
				PRIVACY_POLICY_URL="https://fedoraproject.org/wiki/Legal:PrivacyPolicy"
				VARIANT="Container Image"
				VARIANT_ID=container
			`),
			Expect: &osrelease.Info{
				Name:       "Fedora Linux",
				Id:         "fedora",
				PrettyName: "Fedora Linux 35 (Container Image)",
				Version:    "35 (Container Image)",
				VersionId:  "35",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.Desc, func(t *testing.T) {
			actual, err := osrelease.Parse(strings.NewReader(tc.Value))

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
