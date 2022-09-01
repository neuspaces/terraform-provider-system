package client

import "context"

type Package struct {
	// Manager is an id of the responsible package management system
	// Supported values are "apk"
	Manager PackageManager

	// Name is the name of the package
	Name string

	Version PackageVersion

	State PackageState
}

type PackageVersion struct {
	Required  string
	Installed string
	Available string
}

type PackageManager string

type PackageState uint8

const (
	PackageInstalled    PackageState = 0
	PackageNotInstalled PackageState = 1
)

type PackageClient interface {
	Get(ctx context.Context) (Packages, error)
	Apply(ctx context.Context, pkgs Packages) error
}

// Packages is a list of *Package
type Packages []*Package

func (p Packages) Names() []string {
	var names []string
	for _, pkg := range p {
		names = append(names, pkg.Name)
	}
	return names
}

func (p Packages) ByName() func(i, j int) bool {
	return func(i, j int) bool {
		return p[i].Name < p[j].Name
	}
}

func (p Packages) Filter(f func(pkg *Package) bool) Packages {
	var filtered Packages
	for _, pkg := range p {
		if f(pkg) {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

func (p Packages) ToMap() PackageMap {
	pkgMap := PackageMap{}
	for _, pkg := range p {
		pkgMap[pkg.Name] = pkg
	}
	return pkgMap
}

// PackageMap is a map of *Package indexed by package name
type PackageMap map[string]*Package

// ToList returns Packages sorted by package name
func (p PackageMap) ToList() Packages {
	var pkgList Packages
	for _, pkg := range p {
		pkgList = append(pkgList, pkg)
	}
	return pkgList
}

func PackageNameFilter(names ...string) func(pkg *Package) bool {
	return func(pkg *Package) bool {
		for _, name := range names {
			if name == pkg.Name {
				return true
			}
		}

		return false
	}
}

func PackageStateFiler(state PackageState) func(pkg *Package) bool {
	return func(pkg *Package) bool {
		return pkg.State == state
	}
}
