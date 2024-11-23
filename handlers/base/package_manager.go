package base

type PackageManager interface {
	FetchDependencies(packageInfo *PackageInfo, rootFramework string) error
}
