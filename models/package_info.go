package models

import (
	"fmt"
	"redun-pendancy/utils"
)

type PackageInfo struct {
	Name      string
	Version   string
	Framework string
	FilePath  string

	Parents      []*PackageInfo
	Dependencies []*PackageInfo

	PackageType PackageType
	LoadStatus  LoadStatus
}

func NewPackageInfo(name string, version string, framework string, filePath string) *PackageInfo {
	packageType := utils.TernarySelect[PackageType](version == "", PackageType_Project, PackageType_Package)
	return &PackageInfo{
		Name:      name,
		Version:   version,
		Framework: framework,
		FilePath:  filePath,

		Parents:      []*PackageInfo{},
		Dependencies: []*PackageInfo{},

		PackageType: packageType,
		LoadStatus:  LoadStatus_None,
	}
}

func (packageInfo *PackageInfo) AddDependency(dependency *PackageInfo) {
	packageInfo.Dependencies = append(packageInfo.Dependencies, dependency)
	dependency.Parents = append(dependency.Parents, packageInfo)
}

func (packageInfo *PackageInfo) AddDependencySorted(dependency *PackageInfo) {
	dependency.Parents = append(dependency.Parents, packageInfo)
	index := utils.IndexOf(packageInfo.Dependencies, 0, func(currDependency *PackageInfo) bool {
		return currDependency.Name > dependency.Name
	})
	if index != -1 {
		packageInfo.Dependencies = utils.InsertAt(packageInfo.Dependencies, index, dependency)
		return
	}
	packageInfo.Dependencies = append(packageInfo.Dependencies, dependency)
}

func (packageInfo *PackageInfo) ContainsDependency(packageName string) bool {
	for _, dependency := range packageInfo.Dependencies {
		if dependency.Name == packageName {
			return true
		}
	}
	return false
}

func (packageInfo *PackageInfo) ContainsTransientDependency(packageName string) bool {
	visited := make(map[string]bool)
	return packageInfo.containsTransientDependency(packageName, visited)
}

func (packageInfo *PackageInfo) containsTransientDependency(packageName string, visited map[string]bool) bool {
	if visited[packageInfo.Name] {
		return false
	}
	visited[packageInfo.Name] = true

	for _, dependency := range packageInfo.Dependencies {
		if dependency.Name == packageName {
			return true
		}

		isInDependency := dependency.containsTransientDependency(packageName, visited)
		if isInDependency {
			return true
		}
	}
	return false
}

func (packageInfo *PackageInfo) IsPackage() bool {
	return packageInfo.MatchesType(PackageType_Package)
}

func (packageInfo *PackageInfo) IsProject() bool {
	return packageInfo.MatchesType(PackageType_Project)
}

func (packageInfo *PackageInfo) MatchesType(packageType PackageType) bool {
	result := packageInfo.PackageType & packageType
	return result == packageType
}

func (packageInfo *PackageInfo) IsExeProject() bool {
	return packageInfo.PackageType == PackageType_ExeProject
}

func (packageInfo *PackageInfo) IsTool() bool {
	return packageInfo.PackageType == PackageType_Tool
}

func (packageInfo *PackageInfo) MarkAsLoaded() {
	packageInfo.LoadStatus = LoadStatus_Loaded
}

func (packageInfo *PackageInfo) MarkAsSkipped() {
	packageInfo.LoadStatus = LoadStatus_Skipped
}

func (packageInfo *PackageInfo) MarkAsExeProject() {
	packageInfo.PackageType = PackageType_ExeProject
}

func (packageInfo *PackageInfo) MarkAsTool() {
	packageInfo.PackageType = PackageType_Tool
}

func (packageInfo *PackageInfo) RemoveDependency(packageName string) bool {
	index := utils.IndexOf(packageInfo.Dependencies, 0, func(depedency *PackageInfo) bool {
		return depedency.Name == packageName
	})
	if index == -1 {
		return false
	}

	dependency := packageInfo.Dependencies[index]
	dependency.Parents = utils.RemoveIf(dependency.Parents, func(parent *PackageInfo) bool {
		//Remove ourselves as the parent
		return parent == packageInfo
	})
	packageInfo.Dependencies = utils.RemoveAt(packageInfo.Dependencies, index)
	return true
}

func (packageInfo *PackageInfo) ToString() string {
	if packageInfo.IsProject() {
		return fmt.Sprintf("%s [%s]", packageInfo.Name, packageInfo.Framework)
	}
	return fmt.Sprintf("%s (%s) [%s]", packageInfo.Name, packageInfo.Version, packageInfo.Framework)
}
