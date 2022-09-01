package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/neuspaces/terraform-provider-system/internal/client"
	"regexp"
	"sort"
	"strings"
)

const resourcePackagesApkName = "system_packages_apk"

const (
	resourcePackagesApkAttrId      = "id"
	resourcePackagesApkAttrPackage = "package"

	resourcePackagesApkAttrPackageName    = "name"
	resourcePackagesApkAttrPackageVersion = "version"

	resourcePackagesApkAttrPackageVersions          = "versions"
	resourcePackagesApkAttrPackageVersionsInstalled = "installed"
	resourcePackagesApkAttrPackageVersionsAvailable = "available"
)

func resourcePackagesApk() *schema.Resource {
	sr := &SyncResource{
		CreateContext: resourcePackagesApkCreate,
		ReadContext:   resourcePackagesApkRead,
		UpdateContext: resourcePackagesApkUpdate,
		DeleteContext: resourcePackagesApkDelete,
	}

	return &schema.Resource{
		Description: fmt.Sprintf("`%s` manages one or more apk packages on the remote system.", resourcePackagesApkName),

		CreateContext: sr.CreateContextSync,
		ReadContext:   sr.ReadContextSync,
		UpdateContext: sr.UpdateContextSync,
		DeleteContext: sr.DeleteContextSync,

		// Importer is intentionally not configured
		// Read will not fail if the one or more packages is not installed
		// Create will implicitly import the one or more packages in the state

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			resourcePackagesApkAttrId: {
				Description: "ID of the apk packages",
				Type:        schema.TypeString,
				Computed:    true,
			},
			resourcePackagesApkAttrPackage: {
				Description: "List of packages",
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem:        resourcePackagesApkPackageSchema(),
			},
			internalDataSchemaKey: internalDataSchema(),
		},
	}
}

func resourcePackagesApkPackageSchema() *schema.Resource {
	return &schema.Resource{
		Description: "Package description",
		Schema: map[string]*schema.Schema{
			resourcePackagesApkAttrPackageName: {
				Description: "Name of the package",
				Type:        schema.TypeString,
				Required:    true,
			},
			resourcePackagesApkAttrPackageVersion: {
				Description: "Sticky version of the installed package. Supported values consist of a constraint component and a version component. Supported constraints are `=` (strictly equal), `<` (strictly lower), `<=` (lower or equal), `~` (fuzzy), `>=` (greater or equal), `>` (strictly greater). Example values are `=~1.1` to pin the major/minor version or `=1.1.1n-r0` to pin the exact version. For details on the semantics and more examples, refer to the Alpine wiki on [Holding a specific package back](https://wiki.alpinelinux.org/wiki/Package_management#Holding_a_specific_package_back).",

				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(<|<=|=|~|>=|>)`), "version must begin with a constraint operator"),
			},
			resourcePackagesApkAttrPackageVersions: {
				Description: "Computed version information of the package",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						resourcePackagesApkAttrPackageVersionsInstalled: {
							Description: "Installed version of the package.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						resourcePackagesApkAttrPackageVersionsAvailable: {
							Description: "Available version of the package.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

type resourcePackagesApkInternalData struct {
	// PreInstalled is map with package names as key and true as value if the package was already installed before managed by the resource
	PreInstalled map[string]bool `json:"pre_installed,omitempty"`
}

func resourcePackagesApkIdFromPackages(pkgs client.Packages) string {
	return strings.Join(pkgs.Names(), "|")
}

func packageNamesFromResourcePackagesApkId(id string) []string {
	return strings.Split(id, "|")
}

func expandPackagesApkPackageSet(v interface{}) (map[string]map[string]interface{}, error) {
	pkgsSet, ok := v.(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("expected *schema.Set, got unexpected type %T", v)
	}

	// Use a map with package names as key to detect duplicate package names in the set
	// ValidateDiagFunc on TypeSet in schema structure is not supported in SDK v2.8.0
	pkgsDataMap := make(map[string]map[string]interface{})

	for _, pkgData := range pkgsSet.List() {
		pkgDataMap, ok := pkgData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map[string]interface{}, got unexpected type %T", pkgData)
		}

		pkgName := pkgDataMap[resourcePackagesApkAttrPackageName].(string)
		if pkgName == "" {
			return nil, errors.New("empty package name not allowed")
		}

		if _, duplicatePkgName := pkgDataMap[pkgName]; duplicatePkgName {
			return nil, fmt.Errorf("duplicate package name %s", pkgName)
		}

		pkgsDataMap[pkgName] = pkgDataMap
	}

	return pkgsDataMap, nil
}

// expandPackagesApkPackage takes a *schema.Set and returns a map of client.Package with package name as key
// expandPackagesApkPackage raises an error on duplicate package names
func expandPackagesApkPackage(v interface{}) (map[string]*client.Package, error) {
	pkgsDataMap, err := expandPackagesApkPackageSet(v)
	if err != nil {
		return nil, err
	}

	pkgsMap := make(map[string]*client.Package)

	for pkgName, pkgData := range pkgsDataMap {
		pkg := &client.Package{
			Manager: client.ApkPackageManager,
			Name:    pkgName,
		}

		if pkgVersion, ok := pkgData[resourcePackagesApkAttrPackageVersion].(string); ok {
			pkg.Version.Required = pkgVersion
		}

		pkgsMap[pkgName] = pkg
	}

	return pkgsMap, nil
}

func flattenPackagesApkPackage(v client.Packages) *schema.Set {
	if v == nil {
		return nil
	}

	// Set for attribute `package`
	packageSet := schema.NewSet(schema.HashResource(resourcePackagesApkPackageSchema()), []interface{}{})

	// Packages
	for _, clientPkg := range v {
		pkg := map[string]interface{}{}

		// Name
		pkg[resourcePackagesApkAttrPackageName] = clientPkg.Name

		// Version
		if clientPkg.Version.Required != "" {
			pkg[resourcePackagesApkAttrPackageVersion] = clientPkg.Version.Required
		}

		// Computed versions
		pkgVersions := make(map[string]interface{})

		if clientPkg.Version.Installed != "" {
			pkgVersions[resourcePackagesApkAttrPackageVersionsInstalled] = clientPkg.Version.Installed
		}

		if clientPkg.Version.Available != "" {
			pkgVersions[resourcePackagesApkAttrPackageVersionsAvailable] = clientPkg.Version.Available
		}

		pkg[resourcePackagesApkAttrPackageVersions] = []interface{}{pkgVersions}

		packageSet.Add(pkg)
	}

	return packageSet
}

func resourcePackagesApkGetResourceData(d *schema.ResourceData) (client.Packages, diag.Diagnostics) {
	prevPackageSet, packageSet := d.GetChange(resourcePackagesApkAttrPackage)

	packageMap, err := expandPackagesApkPackage(packageSet)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	prevPackageMap, err := expandPackagesApkPackage(prevPackageSet)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// Result package list
	pkgs := client.Packages{}

	// Packages which should be installed
	for _, pkg := range packageMap {
		installedPkg := *pkg

		// State should be installed
		installedPkg.State = client.PackageInstalled

		pkgs = append(pkgs, &installedPkg)
	}

	// Packages which should be uninstalled, i.e. are in prevPackageMap but not in packageMap
	for prevPackageName, prevPkg := range prevPackageMap {
		if _, keep := packageMap[prevPackageName]; !keep {
			uninstalledPkg := *prevPkg

			// State should be not installed
			uninstalledPkg.State = client.PackageNotInstalled

			pkgs = append(pkgs, &uninstalledPkg)
		}
	}

	// Sort packages
	sort.SliceStable(pkgs, pkgs.ByName())

	return pkgs, nil
}

func resourcePackagesApkSetResourceData(r client.Packages, d *schema.ResourceData) diag.Diagnostics {
	_ = d.Set(resourcePackagesApkAttrPackage, flattenPackagesApkPackage(r))

	return nil
}

func resourcePackagesApkNewClient(ctx context.Context, meta interface{}) (client.PackageClient, diag.Diagnostics) {
	p, diagErr := providerFromMeta(meta)
	if diagErr != nil {
		return nil, diagErr
	}

	c := client.NewApkPackageClient(p.System)

	return c, nil
}

func resourcePackagesApkApply(ctx context.Context, d *schema.ResourceData, meta interface{}) (client.Packages, diag.Diagnostics) {
	c, diagErr := resourcePackagesApkNewClient(ctx, meta)
	if diagErr != nil {
		return nil, diagErr
	}

	// Get packages to determine installation state before apply
	preApplyPackages, err := c.Get(ctx)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	preApplyPackagesMap := preApplyPackages.ToMap()

	r, diagErr := resourcePackagesApkGetResourceData(d)
	if diagErr != nil {
		return nil, diagErr
	}

	err = c.Apply(ctx, r)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// Remember installation state of the package before apply in the internal state
	var internalData resourcePackagesApkInternalData
	_, diagErr = getInternalData(d, &internalData)
	if diagErr != nil {
		return nil, diagErr
	}

	preInstalled := internalData.PreInstalled
	if preInstalled == nil {
		preInstalled = map[string]bool{}
	}

	for _, pkg := range r {
		if pkg.State == client.PackageInstalled {
			if _, inPreInstalled := preInstalled[pkg.Name]; !inPreInstalled {
				// If package is installed and not yet recorded in internal data remember the pre apply state
				preApplyPkg, inPreApplyPkg := preApplyPackagesMap[pkg.Name]
				preInstalled[pkg.Name] = inPreApplyPkg && preApplyPkg.State == client.PackageInstalled
			}
		} else if pkg.State == client.PackageNotInstalled {
			// If package is uninstalled remove from internal data
			delete(preInstalled, pkg.Name)
		}
	}

	internalData.PreInstalled = preInstalled
	diagErr = setInternalData(d, &internalData)
	if diagErr != nil {
		return nil, diagErr
	}

	return r, nil
}

func resourcePackagesApkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourcePackagesApkNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	packageNames := packageNamesFromResourcePackagesApkId(id)

	r, err := c.Get(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Filter for relevant packages
	r = r.Filter(client.PackageNameFilter(packageNames...))

	// Filter for installed packages
	r = r.Filter(client.PackageStateFiler(client.PackageInstalled))

	diagErr = resourcePackagesApkSetResourceData(r, d)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func resourcePackagesApkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	r, diagErr := resourcePackagesApkApply(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	id := resourcePackagesApkIdFromPackages(r)
	d.SetId(id)

	return resourcePackagesApkRead(ctx, d, meta)
}

func resourcePackagesApkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	r, diagErr := resourcePackagesApkApply(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	id := resourcePackagesApkIdFromPackages(r)
	d.SetId(id)

	return resourcePackagesApkRead(ctx, d, meta)
}

func resourcePackagesApkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := resourcePackagesApkNewClient(ctx, meta)
	if diagErr != nil {
		return diagErr
	}

	// In Delete func, GetChange does not provide information that the resource is deleted
	r, diagErr := resourcePackagesApkGetResourceData(d)
	if diagErr != nil {
		return diagErr
	}

	// Restore the installation state of the packages before create
	var internalData resourcePackagesApkInternalData
	_, diagErr = getInternalData(d, &internalData)
	if diagErr != nil {
		return diagErr
	}

	for _, pkg := range r {
		if internalData.PreInstalled != nil {
			if preApplyPkg, inPreApplyPkg := internalData.PreInstalled[pkg.Name]; inPreApplyPkg {
				if preApplyPkg {
					pkg.State = client.PackageInstalled
					continue
				}
			}
		}

		pkg.State = client.PackageNotInstalled
	}

	err := c.Apply(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
