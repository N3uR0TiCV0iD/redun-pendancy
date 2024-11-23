package base

import (
	"fmt"
	"log"
	"redun-pendancy/models"
	"redun-pendancy/utils"
)

type PackageContainer struct {
	packages        map[string]*PackageInfo
	highestVersions map[string]string //PackageName => Version
}

func NewPackageContainer() *PackageContainer {
	return &PackageContainer{
		packages:        make(map[string]*PackageInfo),
		highestVersions: make(map[string]string),
	}
}

func (packageContainer *PackageContainer) GetPackages() map[string]*PackageInfo {
	return packageContainer.packages
}

func (packageContainer *PackageContainer) HasPackage(name string, version string, framework string) bool {
	key := buildPackageKey(name, version, framework)
	_, exists := packageContainer.packages[key]
	return exists
}

func (packageContainer *PackageContainer) GetOrCreatePackage(name string, version string, framework string, filePath string) *PackageInfo {
	key := buildPackageKey(name, version, framework)
	packageInfo, exists := packageContainer.packages[key]
	if !exists {
		packageInfo = models.NewPackageInfo(name, version, framework, filePath)
		packageContainer.packages[key] = packageInfo

		versionKey := name
		highestVersion, exists := packageContainer.highestVersions[versionKey]
		if !exists || utils.IsVersionHigher(version, highestVersion) {
			packageContainer.highestVersions[versionKey] = version
		}
	}
	return packageInfo
}

func (packageContainer *PackageContainer) Load(packageInfo *PackageInfo, packageManager PackageManager, rootFramework string) error {
	if packageInfo.LoadStatus != models.LoadStatus_None {
		return nil
	}

	if packageContainer.shouldSkipPackage(packageInfo) {
		log.Println("Skipping:", packageInfo.ToString())
		packageInfo.MarkAsSkipped()
		return nil
	}

	log.Println("Loading:", packageInfo.ToString())
	packageInfo.MarkAsLoaded()
	if !packageInfo.IsProject() {
		err := packageManager.FetchDependencies(packageInfo, rootFramework)
		if err != nil {
			return err
		}
	}

	for _, dependency := range packageInfo.Dependencies {
		fmt.Printf("- %s %s\n", dependency.Name, dependency.Version)
	}
	fmt.Println()

	for _, dependency := range packageInfo.Dependencies {
		err := packageContainer.Load(dependency, packageManager, rootFramework)
		if err != nil {
			return err
		}
	}
	return nil
}

func (packageContainer *PackageContainer) shouldSkipPackage(packageInfo *PackageInfo) bool {
	key := packageInfo.Name
	highestVersion, exists := packageContainer.highestVersions[key]
	if !exists {
		//Technically this should never happen (unless the package is created from the outside)
		return false
	}

	version := packageInfo.Version
	if version == highestVersion || utils.IsVersionHigher(version, highestVersion) {
		return false
	}
	return true
}

func buildPackageKey(name string, version string, framework string) string {
	return utils.BuildPackageKey(name, version, framework)
}
