package client

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"io"
	"sort"
	"strings"
)

const AptPackageManager PackageManager = "apt"

var (
	ErrAptPackage = errors.New("apt package resource")

	ErrAptPackageManagerNotAvailable = errors.Join(ErrAptPackage, errors.New("apt not available"))

	ErrAptPackageManager = errors.Join(ErrAptPackage, errors.New("apt error"))

	ErrAptPackageUnexpected = errors.Join(ErrAptPackage, errors.New("unexpected error"))
)

func NewAptPackageClient(s system.System) PackageClient {
	return &aptPackageClient{
		s: s,
	}
}

type aptPackageClient struct {
	s system.System
}

// Get returns a list of Packages which contain all installed packages. Each Package contains the available version. The caller of Get may further filter the returned Packages.
func (c *aptPackageClient) Get(ctx context.Context) (Packages, error) {
	cmd := NewCommand(`_do() { which dpkg-query >/dev/null 2>&1; which_dpkg_query_rc=$?; if [ $which_dpkg_query_rc -eq 0 ]; then dpkg-query --show --no-pager --showformat='"${Package}","${Version}","${db:Status-Abbrev}","${Status}"\n'; else echo "which_dpkg_query_rc=${which_dpkg_query_rc}"; fi }; _do;`)
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, errors.Join(ErrAptPackage, err)
	}

	if strings.HasPrefix(res.StdoutString(), "which_dpkg_query_rc=") && res.StdoutString() != "which_dpkg_query_rc=0" {
		return nil, ErrAptPackageManagerNotAvailable
	}

	if res.ExitCode != 0 {
		return nil, ErrAptPackageUnexpected
	}

	// Parse output of `dpkg-query --show`
	// Expect CSV with double-quoted fields
	dpkgQueryReader := csv.NewReader(bytes.NewReader(res.Stdout))

	// Construct result
	var pkgs Packages

	for {
		dpkgQueryPackage, err := dpkgQueryReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Join(ErrAptPackage, err)
		}
		if len(dpkgQueryPackage) != 4 {
			return nil, ErrAptPackageUnexpected
		}

		dpkgQueryPackageName := dpkgQueryPackage[0]
		dpkgQueryPackageVersion := dpkgQueryPackage[1]

		dpkgQueryPackageState := dpkgQueryPackage[2]

		// `dpkg-query --show` field ${db:Status-Abbrev} must be exactly 3 characters
		if len(dpkgQueryPackageState) != 3 {
			return nil, ErrAptPackageUnexpected
		}

		// first character: desired state
		// second character: actual state

		// Consider the package as installed if and only if current state character is `i`
		var pkgState PackageState
		switch dpkgQueryPackageState[1] {
		case 'i':
			pkgState = PackageInstalled
		default:
			pkgState = PackageNotInstalled
		}

		pkg := &Package{
			Manager: AptPackageManager,
			Name:    dpkgQueryPackageName,
			Version: PackageVersion{
				Installed: dpkgQueryPackageVersion,
			},
			State: pkgState,
		}

		pkgs = append(pkgs, pkg)
	}

	// Sort packages by name
	sort.SliceStable(pkgs, pkgs.ByName())

	return pkgs, nil
}

func (c *aptPackageClient) Apply(ctx context.Context, pkgs Packages) error {
	if len(pkgs) == 0 {
		// Nothing to apply
		return nil
	}

	// Construct package install/remove arguments
	var aptInstallPkgs []string

	for _, pkg := range pkgs {
		if pkg.State == PackageInstalled {
			aptInstallPkgs = append(aptInstallPkgs, fmt.Sprintf(`'%s+'`, pkg.Name))
		} else if pkg.State == PackageNotInstalled {
			aptInstallPkgs = append(aptInstallPkgs, fmt.Sprintf(`'%s-'`, pkg.Name))
		}
	}

	cmd := NewCommand(fmt.Sprintf(`_do() { export DEBIAN_FRONTEND=noninteractive DEBIAN_PRIORITY=critical LANGUAGE=C LANG=C LC_ALL=C LC_MESSAGES=C LC_CTYPE=C; apt-get update >/dev/null 2>&1; apt_update_rc=$?; if [ $apt_update_rc -ne 0 ]; then echo "apt_update_rc=${apt_update_rc}"; fi; apt-get install --no-install-recommends %[1]s -y -q; }; _do;`, strings.Join(aptInstallPkgs, " ")))
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return errors.Join(ErrAptPackageManager, err)
	}

	if strings.HasPrefix(res.StdoutString(), "apt_update_rc=") && res.StdoutString() != "which_dpkg_query_rc=0" {
		return errors.Join(ErrAptPackageManager, errors.New(res.StderrString()))
	}

	if res.ExitCode != 0 {
		return errors.Join(ErrAptPackageManager, errors.New(res.StderrString()))
	}

	return nil
}
