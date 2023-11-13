package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/neuspaces/terraform-provider-system/internal/system"
	"regexp"
	"sort"
	"strings"
)

const ApkPackageManager PackageManager = "apk"

var (
	ErrApkPackage = errors.New("apk package resource")

	ErrApkPackageManagerNotAvailable = errors.Join(ErrApkPackage, errors.New("apk not available"))

	ErrApkPackageManager = errors.Join(ErrApkPackage, errors.New("apk error"))

	ErrApkPackageUnexpected = errors.Join(ErrApkPackage, errors.New("unexpected error"))
)

var (
	// apkPackageInfoRegexp matches lines emitted by `apk -v info`
	apkPackageInfoRegexp = regexp.MustCompile(`(?m)^(?P<name>[\S]+)-(?P<version>\d\S*)\s*$`)

	// apkPackageVersionRegexp matches lines emitted by `apk -v version`
	apkPackageVersionRegexp = regexp.MustCompile(`(?m)^(?P<name>[\S]+)-(?P<version_installed>\d\S*)\s*(=|<)\s*(?P<version_available>\d\S*)\s*$`)

	apkWorldRegexp = regexp.MustCompile(`(?m)^(?P<name>[\S]+?)(?P<version_spec>(?P<version_prefix>=|\<|\>|=~)(?P<version>[\S]+))?\s*$`)
)

func NewApkPackageClient(s system.System) PackageClient {
	return &apkPackageClient{
		s: s,
	}
}

type apkPackageClient struct {
	s system.System
}

// Get returns a list of Packages which contain all installed packages. Each Package contains the available version. The caller of Get may further filter the returned Packages.
func (c *apkPackageClient) Get(ctx context.Context) (Packages, error) {
	// Run command `apk version`
	// - returned in formation is used to determine installed and available versions
	cmd := NewCommand(`_do() { which apk >/dev/null 2>&1; which_apk_rc=$?; if [ $which_apk_rc -eq 0 ]; then apk -v version; else echo "which_apk_rc=${which_apk_rc}"; fi }; _do;`)
	res, err := ExecuteCommand(ctx, c.s, cmd)
	if err != nil {
		return nil, errors.Join(ErrApkPackage, err)
	}

	if strings.HasPrefix(res.StdoutString(), "which_apk_rc=") && res.StdoutString() != "which_apk_rc=0" {
		return nil, ErrApkPackageManagerNotAvailable
	}

	if res.ExitCode != 0 || len(res.Stdout) == 0 {
		return nil, ErrApkPackageUnexpected
	}

	// Parse output of from apk version
	apkVersionPackages, err := convertApkVersionToPackages(res.Stdout)
	if err != nil {
		return nil, errors.Join(ErrApkPackage, err)
	}
	apkVersionPackageMap := apkVersionPackages.ToMap()

	// Get /etc/apk/world
	// - /etc/apk/world contains the packages which have been explicitly installed by the user
	// - only packages which are in /etc/apk/world are returned by the client
	apkWorld, err := getApkWorld(ctx, c.s)
	if err != nil {
		return nil, errors.Join(ErrApkPackage, err)
	}

	// Get package list from /etc/apk/world
	apkWorldPackages, err := convertApkWorldToPackages(apkWorld)
	if err != nil {
		return nil, errors.Join(ErrApkPackage, err)
	}

	// Construct result
	var pkgs Packages

	for _, pkg := range apkWorldPackages {
		finalPkg := *pkg

		// Lookup package in apk version output
		apkVersionPkg, ok := apkVersionPackageMap[pkg.Name]
		if ok {
			// Skipped if apk version output does not contain version information; this might occur for compatibility packages
			finalPkg.Version.Available = apkVersionPkg.Version.Available
			finalPkg.Version.Installed = apkVersionPkg.Version.Installed
		}

		pkgs = append(pkgs, &finalPkg)
	}

	// Sort packages by name
	sort.SliceStable(pkgs, pkgs.ByName())

	return pkgs, nil
}

func (c *apkPackageClient) Apply(ctx context.Context, pkgs Packages) error {
	if len(pkgs) == 0 {
		// Nothing to apply
		return nil
	}

	// Get /etc/apk/world
	apkWorld, err := getApkWorld(ctx, c.s)
	if err != nil {
		return errors.Join(ErrApkPackage, err)
	}

	// Get package list from /etc/apk/world
	apkWorldPackages, err := convertApkWorldToPackages(apkWorld)
	if err != nil {
		return errors.Join(ErrApkPackage, err)
	}

	// Construct package map from /etc/apk/world
	apkWorldPackageMap := apkWorldPackages.ToMap()

	// Apply changes to a packages map
	for _, pkg := range pkgs {
		if pkg.State == PackageInstalled {
			// Add package to /etc/apk/world

			apkWorldPackageMap[pkg.Name] = &Package{
				Manager: ApkPackageManager,
				Name:    pkg.Name,
				Version: PackageVersion{
					Required: pkg.Version.Required,
				},
				State: PackageInstalled,
			}
		} else if pkg.State == PackageNotInstalled {
			// Remove package to /etc/apk/world
			delete(apkWorldPackageMap, pkg.Name)
		}
	}

	// Render modified /etc/apk/world
	newApkWorldPackages := apkWorldPackageMap.ToList()
	newApkWorld := convertPackagesToApkWorld(newApkWorldPackages)

	// Return early if no change in /etc/apk/world
	if string(newApkWorld) == string(apkWorld) {
		// Nothing to apply
		return nil
	}

	apkUpgradeCmd := NewInputCommand(`{ ! which apk >/dev/null 2>&1 && { >&2 echo '{"pre":1}'; }; } || { cat - > /etc/apk/world.new; mv /etc/apk/world /etc/apk/world.old; mv /etc/apk/world.new /etc/apk/world; apk upgrade; rm -f /etc/apk/world.old; } || { [ -f /etc/apk/world.new ] && rm -f /etc/apk/world.new; [ -f /etc/apk/world.old ] && { rm -f /etc/apk/world; mv /etc/apk/world.old /etc/apk/world; apk upgrade }; }; }`, bytes.NewReader(newApkWorld))

	apkUpgradeRes, err := ExecuteCommand(ctx, c.s, apkUpgradeCmd)
	if err != nil {
		return errors.Join(ErrApkPackage, err)
	}
	if apkUpgradeRes.ExitCode != 0 {
		return errors.Join(ErrApkPackageManager, errors.New(string(apkUpgradeRes.Stderr)))
	}

	return nil
}

func getApkWorld(ctx context.Context, s system.System) ([]byte, error) {
	// Get /etc/apk/world
	apkWorldCatRes, err := ExecuteCommand(ctx, s, &CatCommand{Path: "/etc/apk/world"})
	if err != nil {
		return nil, errors.Join(ErrApkPackage, err)
	}
	if apkWorldCatRes.ExitCode != 0 {
		return nil, ErrApkPackageUnexpected
	}

	return apkWorldCatRes.Stdout, nil
}

// convertApkVersionToPackages converts the output of `apk version` to Packages
func convertApkVersionToPackages(apkWorld []byte) (Packages, error) {
	var pkgs Packages

	pkgMatches := apkPackageVersionRegexp.FindAllStringSubmatch(strings.TrimSpace(string(apkWorld)), -1)

	for _, pkgMatch := range pkgMatches {
		// expect 5 matched groups
		if len(pkgMatch) != 5 {
			continue
		}

		pkgs = append(pkgs, &Package{
			Manager: ApkPackageManager,
			Name:    pkgMatch[1],
			Version: PackageVersion{
				Required:  "",
				Installed: pkgMatch[2],
				Available: pkgMatch[4],
			},
			State: PackageInstalled,
		})
	}

	return pkgs, nil
}

// convertApkWorldToPackages converts an /etc/apk/world to Packages
func convertApkWorldToPackages(apkWorld []byte) (Packages, error) {
	var pkgs Packages

	apkWorldMatches := apkWorldRegexp.FindAllSubmatch(apkWorld, -1)
	for _, apkWorldMatch := range apkWorldMatches {
		if len(apkWorldMatch) != 5 {
			continue
		}

		pkg := &Package{
			Manager: ApkPackageManager,
			Name:    string(apkWorldMatch[1]),
			Version: PackageVersion{
				Required: string(apkWorldMatch[2]),
			},
			State: PackageInstalled,
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

// convertApkWorldToPackages converts a list of Packages to /etc/apk/world
func convertPackagesToApkWorld(pkgs Packages) []byte {
	// Sort packages
	sort.SliceStable(pkgs, pkgs.ByName())

	// Render lines of /etc/apk/world
	apkWorld := new(bytes.Buffer)
	for _, pkg := range pkgs {
		version := pkg.Version.Required

		// Implicit operator = if not specified
		if version != "" && !hasApkVersionOperator(version) {
			version = "=" + version
		}

		_, _ = fmt.Fprintf(apkWorld, "%s%s\n", pkg.Name, version)
	}

	return apkWorld.Bytes()
}

// hasApkVersionOperator returns true if the provided version is prefixed with an apk version operator (=, <, <=, >, >=, ~)
func hasApkVersionOperator(version string) bool {
	return strings.HasPrefix(version, "=") ||
		strings.HasPrefix(version, "<") ||
		strings.HasPrefix(version, "<=") ||
		strings.HasPrefix(version, ">") ||
		strings.HasPrefix(version, ">=") ||
		strings.HasPrefix(version, "~")
}
