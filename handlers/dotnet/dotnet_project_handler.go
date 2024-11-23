package dotnet

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"redun-pendancy/handlers/base"
	"redun-pendancy/helpers"
	"redun-pendancy/utils"
	"regexp"
	"strings"
)

type DotNetProjectHandler struct {
	packageContainer *PackageContainer
	packageManager   *DotNetPackageManager
	projectFiles     map[string]*DotNetProjectFile

	projects     []*PackageInfo
	solutionName string

	globalPackages    map[string]string //PackageName => Version
	hasGlobalPackages bool

	globalPackagesFile  *helpers.LazyBufferedFile
	hasGlobalPkgChanges bool
}

func NewDotNetProjectHandler(userHomePath string) *DotNetProjectHandler {
	packageContainer := base.NewPackageContainer()
	return &DotNetProjectHandler{
		packageContainer: packageContainer,
		packageManager:   NewNuGetPackageManager(userHomePath, packageContainer),
		projectFiles:     make(map[string]*DotNetProjectFile),
	}
}

func (projectHandler *DotNetProjectHandler) GetWorkspaceName() string {
	return projectHandler.solutionName
}

func (projectHandler *DotNetProjectHandler) GetProjects() []*PackageInfo {
	return projectHandler.projects
}

func (projectHandler *DotNetProjectHandler) GetPackageContainer() *PackageContainer {
	return projectHandler.packageContainer
}

func (projectHandler *DotNetProjectHandler) Initialize(solutionFilePath string) error {
	projectPaths, err := extractProjectPaths(solutionFilePath)
	if err != nil {
		return err
	}

	globalPackages, err := projectHandler.loadGlobalPackages(solutionFilePath)
	if err != nil {
		return err
	}

	projectLoader := NewDotNetProjectLoader(projectHandler.packageContainer, globalPackages)
	for _, projectFilePath := range projectPaths {
		projectFile, err := projectLoader.GetOrLoad(projectFilePath)
		if err != nil {
			return err
		}

		project := projectFile.GetProject()
		projectHandler.projectFiles[project.Name] = projectFile
		projectHandler.projects = append(projectHandler.projects, project)
	}

	fmt.Println()
	err = projectHandler.initProjects()
	if err != nil {
		return err
	}

	projectHandler.globalPackages = globalPackages
	projectHandler.solutionName = filepath.Base(solutionFilePath)
	projectHandler.hasGlobalPackages = len(globalPackages) != 0
	return nil
}

func (projectHandler *DotNetProjectHandler) loadGlobalPackages(solutionFilePath string) (map[string]string, error) {
	folderPath := filepath.Dir(solutionFilePath)
	packagePropsFilePath := path.Join(folderPath, "Directory.Packages.props")
	globalPackagesFile, err := helpers.NewLazyBufferedFile(packagePropsFilePath)
	if err != nil {
		return nil, err
	}

	packageVersions := make(map[string]string)
	lines, err := globalPackagesFile.GetLines()
	if err == nil {
		for _, line := range lines {
			processPackagePropsLine(line, packageVersions)
		}
		projectHandler.globalPackagesFile = globalPackagesFile
		return packageVersions, nil
	}

	if os.IsNotExist(err) {
		//Optional file, safe to continue without it.
		fmt.Printf("Solution has no \"Directory.Packages.props\".\n\n")
		return packageVersions, nil
	}
	return nil, err
}

func processPackagePropsLine(line string, packageVersions map[string]string) {
	if !strings.Contains(line, `<PackageVersion Include="`) {
		return
	}
	parts := strings.Split(line, "\"")
	packageName := parts[1]
	version := parts[3]
	packageVersions[packageName] = version
}

func (projectHandler *DotNetProjectHandler) initProjects() error {
	for _, project := range projectHandler.projects {
		err := projectHandler.packageContainer.Load(project, projectHandler.packageManager, project.Framework)
		if err != nil {
			return err
		}
	}
	return nil
}

func (projectHandler *DotNetProjectHandler) GetProject(projectName string) *PackageInfo {
	project, _ := utils.FirstOrDefault(projectHandler.projects, func(p *PackageInfo) bool {
		return p.Name == projectName
	})
	return project
}

func (projectHandler *DotNetProjectHandler) AddDependency(projectName string, dependency *PackageInfo) {
	projectFile := projectHandler.projectFiles[projectName]
	projectFile.AddDependency(dependency, projectHandler.hasGlobalPackages)
}

func (projectHandler *DotNetProjectHandler) RemoveDependency(projectName string, dependency *PackageInfo) {
	projectFile := projectHandler.projectFiles[projectName]
	projectFile.RemoveDependency(dependency)
}

func (projectHandler *DotNetProjectHandler) DividesProjectsAndPackages(projectName string) bool {
	projectFile := projectHandler.projectFiles[projectName]
	return projectFile.DividesProjectsAndPackages()
}

func (projectHandler *DotNetProjectHandler) SortDependencies(projectName string, dependencies []*PackageInfo) {
	project := projectHandler.GetProject(projectName)
	projectFile := projectHandler.projectFiles[projectName]
	for _, dependency := range dependencies {
		project.RemoveDependency(dependency.Name)
		project.AddDependencySorted(dependency)
		projectFile.SortDependency(dependency)
	}
}

func (projectHandler *DotNetProjectHandler) GetGlobalPackages() map[string]string {
	return projectHandler.globalPackages
}

func (projectHandler *DotNetProjectHandler) RemoveGlobalPackage(packageName string) bool {
	_, exists := projectHandler.globalPackages[packageName]
	if !exists {
		return false
	}

	globalPackagesFile := projectHandler.globalPackagesFile
	pattern := fmt.Sprintf(`<PackageVersion Include="%s"`, packageName)
	regex := regexp.MustCompile(pattern)
	index, _ := globalPackagesFile.FindLineIndex(regex, 0) //Ignoring error, as the file should already be loaded
	if index == -1 {
		return false
	}

	projectHandler.hasGlobalPkgChanges = true
	globalPackagesFile.RemoveLine(index)
	return true
}

func (projectHandler *DotNetProjectHandler) CommitChanges() error {
	for _, projectFile := range projectHandler.projectFiles {
		err := projectFile.Commit()
		if err != nil {
			return err
		}
	}
	if projectHandler.hasGlobalPkgChanges {
		err := commitGlobalPackages(projectHandler.globalPackagesFile)
		if err != nil {
			return err
		}
		projectHandler.hasGlobalPkgChanges = false
	}
	fmt.Println()
	return nil
}

func commitGlobalPackages(globalPackagesFile *helpers.LazyBufferedFile) error {
	log.Println("Writing:", globalPackagesFile.FilePath)
	err := globalPackagesFile.Commit()
	return err
}

func (projectHandler *DotNetProjectHandler) RevertChanges() {
	for _, projectFile := range projectHandler.projectFiles {
		projectFile.RevertChanges(projectHandler.hasGlobalPackages)
	}
	if projectHandler.hasGlobalPkgChanges {
		projectHandler.revertGlobalPackages()
	}
}

func (projectHandler *DotNetProjectHandler) revertGlobalPackages() {
	globalPackagesFile := projectHandler.globalPackagesFile
	err := globalPackagesFile.Reload()
	if err != nil {
		log.Printf(`[Warning] Failed to reload "%s": %v`, globalPackagesFile.FilePath, err)
		return
	}
	projectHandler.hasGlobalPkgChanges = false
}
