package base

type ProjectHandler interface {
	Initialize(filePath string) error

	GetWorkspaceName() string
	GetPackageContainer() *PackageContainer

	GetProjects() []*PackageInfo

	AddDependency(projectName string, dependency *PackageInfo)
	RemoveDependency(projectName string, dependency *PackageInfo)

	DividesProjectsAndPackages(projectName string) bool
	SortDependencies(projectName string, dependencies []*PackageInfo)

	GetGlobalPackages() map[string]string //PackageName => Version
	RemoveGlobalPackage(packageName string) bool

	CommitChanges() error
	RevertChanges()
}
