package dotnet

import (
	"fmt"
	"log"
	"path/filepath"
	"redun-pendancy/utils"
	"sort"
	"strings"

	"github.com/beevik/etree"
)

type DotNetPackageManager struct {
	packageContainer  *PackageContainer
	nugetPackagesPath string
}

func NewNuGetPackageManager(userHomePath string, packageContainer *PackageContainer) *DotNetPackageManager {
	return &DotNetPackageManager{
		packageContainer:  packageContainer,
		nugetPackagesPath: filepath.Join(userHomePath, ".nuget", "packages"),
	}
}

func (packageManager *DotNetPackageManager) FetchDependencies(packageInfo *PackageInfo, rootFramework string) error {
	packageNameLower := strings.ToLower(packageInfo.Name)
	packageSpecPath := filepath.Join(
		packageManager.nugetPackagesPath,
		packageNameLower,
		packageInfo.Version,
		packageNameLower+".nuspec",
	)

	packageInfo.FilePath = packageSpecPath
	err := packageManager.load(packageInfo, packageSpecPath, rootFramework)
	return err
}

func (packageManager *DotNetPackageManager) load(packageInfo *PackageInfo, packageSpecPath string, rootFramework string) error {
	document := etree.NewDocument()
	err := document.ReadFromFile(packageSpecPath)
	if err != nil {
		return err
	}

	if isToolPackage(document) {
		//Package is a tool. Let's mark it as such without checking its dependencies
		packageInfo.MarkAsTool()
		return nil
	}

	dependencies := document.FindElement("//package/metadata/dependencies")
	if dependencies == nil || len(dependencies.Child) == 0 {
		//Package has no dependencies, nothing to do
		return nil
	}

	frameworkGroups := extractFrameworkGroups(dependencies)
	if len(frameworkGroups) == 0 {
		packageManager.processDependencies(dependencies, packageInfo, packageInfo.Framework)
		return nil
	}

	handled, rootFamily, err := packageManager.loadWithFrameworkOrFamily(frameworkGroups, packageInfo, rootFramework)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	var packageFamily *DotNetFrameworkFamily
	framework := packageInfo.Framework
	if !rootFamily.ContainsFramework(framework) {
		//Package framework family differs from root framework family
		handled, packageFamily, err = packageManager.loadWithFrameworkOrFamily(frameworkGroups, packageInfo, framework)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	} else {
		packageFamily = rootFamily
	}

	// Fallback to ".NETStandard" (if the frameworks used weren't already ".NETStandard")
	netstandardFamily := &frameworkFamilies[2]
	if rootFamily != netstandardFamily && packageFamily != netstandardFamily {
		if packageManager.tryLoadWithFamilyFramework(frameworkGroups, packageInfo, netstandardFamily) {
			return nil
		}
	}

	log.Printf(`[Warning] Could not resolve dependencies for package "%s"`, packageInfo.ToString())
	return nil
}

func isToolPackage(document *etree.Document) bool {
	developmentDependency := document.FindElement("//package/metadata/developmentDependency")
	if developmentDependency == nil {
		return false
	}

	isTool := developmentDependency.Text() == "true"
	return isTool
}

func (packageManager *DotNetPackageManager) processDependencies(parentNode *etree.Element, packageInfo *PackageInfo, framework string) {
	for _, dependencyNode := range parentNode.SelectElements("dependency") {
		dependency := packageManager.processDependency(dependencyNode, framework)
		if dependency != nil {
			packageInfo.AddDependency(dependency)
		}
	}
}

func (packageManager *DotNetPackageManager) loadWithFrameworkOrFamily(frameworkGroups map[string]*etree.Element, packageInfo *PackageInfo, framework string) (bool, *DotNetFrameworkFamily, error) {
	if packageManager.loadWithFramework(frameworkGroups, packageInfo, framework) {
		//Handled with exact matching framework
		return true, nil, nil
	}

	frameworkFamily := getFrameworkFamily(framework)
	if frameworkFamily == nil {
		return false, nil, fmt.Errorf("failed to determine target framework family of: %s", framework)
	}

	if packageManager.tryLoadWithFamilyFramework(frameworkGroups, packageInfo, frameworkFamily) {
		return true, frameworkFamily, nil
	}
	return false, frameworkFamily, nil
}

func getFrameworkFamily(framework string) *DotNetFrameworkFamily {
	for _, frameworkFamily := range frameworkFamilies {
		if frameworkFamily.ContainsFramework(framework) {
			return &frameworkFamily
		}
	}
	return nil
}

func extractFrameworkGroups(dependencies *etree.Element) map[string]*etree.Element {
	frameworkGroups := make(map[string]*etree.Element)
	for _, groupNode := range dependencies.SelectElements("group") {
		targetFramework := groupNode.SelectAttrValue("targetFramework", "")
		if targetFramework != "" {
			frameworkGroups[targetFramework] = groupNode
		}
	}
	return frameworkGroups
}

func (packageManager *DotNetPackageManager) loadWithFramework(frameworkGroups map[string]*etree.Element, packageInfo *PackageInfo, framework string) bool {
	groupNode, exists := frameworkGroups[framework]
	if !exists {
		return false
	}
	packageManager.processDependencies(groupNode, packageInfo, framework)
	return true
}

func (packageManager *DotNetPackageManager) processDependency(dependencyNode *etree.Element, framework string) *PackageInfo {
	packageName := dependencyNode.SelectAttrValue("id", "")
	if packageName == "" {
		//Super unlikely to happen though
		log.Print("[Warning] Skipped dependency node due to missing 'id' attribute.")
		return nil
	}

	version := dependencyNode.SelectAttrValue("version", "")
	if version == "" {
		//Super unlikely to happen though
		log.Print("[Warning] Skipped dependency node due to missing 'version' attribute.")
		return nil
	}

	normalizedVersion := normalizeVersion(version)
	dependency := packageManager.packageContainer.GetOrCreatePackage(packageName, normalizedVersion, framework, "NOT_LOADED")
	return dependency
}

func normalizeVersion(version string) string {
	splitIndex := strings.Index(version, ",")
	if splitIndex != -1 {
		//Keep only the first version number before the comma
		version = version[:splitIndex]
	}
	return utils.NormalizeVersion(version)
}

func (packageManager *DotNetPackageManager) tryLoadWithFamilyFramework(frameworkGroups map[string]*etree.Element, packageInfo *PackageInfo, frameworkFamily *DotNetFrameworkFamily) bool {
	familyFrameworks := getSortedFamilyFrameworks(frameworkGroups, frameworkFamily)
	for _, framework := range familyFrameworks {
		if packageManager.loadWithFramework(frameworkGroups, packageInfo, framework) {
			return true
		}
	}
	return false
}

func getSortedFamilyFrameworks(frameworkGroups map[string]*etree.Element, frameworkFamily *DotNetFrameworkFamily) []string {
	var familyFrameworks []string
	for framework := range frameworkGroups {
		if frameworkFamily.ContainsFramework(framework) {
			familyFrameworks = append(familyFrameworks, framework)
		}
	}

	if len(familyFrameworks) <= 1 {
		//Return an empty array or a single element array
		return familyFrameworks
	}

	versionMap := make(map[string]string)
	for _, framework := range familyFrameworks {
		versionMap[framework] = frameworkFamily.VersionFunc(framework)
	}

	//Sort "familyFrameworks" by __descending__ framework version
	sort.Slice(familyFrameworks, func(i, j int) bool {
		//NOTE: We compare __versions__ here (eg: 8.0 < 5.0)
		left := versionMap[familyFrameworks[i]]
		right := versionMap[familyFrameworks[j]]
		return left > right
	})
	return familyFrameworks
}
