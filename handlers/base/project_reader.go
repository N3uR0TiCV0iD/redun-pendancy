package base

type ProjectReader interface {
	Load(projectPath string) *PackageInfo
}
