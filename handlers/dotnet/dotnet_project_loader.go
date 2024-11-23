package dotnet

import (
	"fmt"
	"log"
	"path/filepath"
	"redun-pendancy/helpers"

	"github.com/beevik/etree"
)

type DotNetProjectLoader struct {
	packageContainer  *PackageContainer
	loadedProjects    map[string]*DotNetProjectFile //Tracks loaded projects to ensure idempotency
	globalPackages    map[string]string
	hasGlobalPackages bool
}

func NewDotNetProjectLoader(packageContainer *PackageContainer, globalPackages map[string]string) *DotNetProjectLoader {
	return &DotNetProjectLoader{
		packageContainer:  packageContainer,
		loadedProjects:    make(map[string]*DotNetProjectFile),
		globalPackages:    globalPackages,
		hasGlobalPackages: len(globalPackages) != 0,
	}
}

func (projectLoader *DotNetProjectLoader) GetOrLoad(projectPath string) (*DotNetProjectFile, error) {
	normalizedPath := filepath.Clean(projectPath)
	projectFile, exists := projectLoader.loadedProjects[normalizedPath]
	if exists {
		return projectFile, nil
	}

	projectFile, err := projectLoader.load(normalizedPath)
	return projectFile, err
}

func (projectLoader *DotNetProjectLoader) load(projectPath string) (*DotNetProjectFile, error) {
	projectName := filepath.Base(projectPath)
	log.Println("Reading:", projectName)

	xmlFile, err := helpers.NewXMLFile(projectPath)
	if err != nil {
		return nil, err
	}

	err = xmlFile.LoadOrSkip()
	if err != nil {
		return nil, err
	}

	projectNode := xmlFile.Document.SelectElement("Project")
	if projectNode == nil {
		return nil, fmt.Errorf("<Project> element not found in file: %s", projectPath)
	}

	targetFramework, isExeProject := extractFrameworkAndExeFlag(projectNode)
	if targetFramework == nil {
		return nil, fmt.Errorf("<TargetFramework> element not found in file: %s", projectPath)
	}

	project := projectLoader.createProjectPackage(projectPath, targetFramework.Text())
	projectFile := NewDotNetProjectFile(project, xmlFile)
	projectLoader.loadedProjects[projectPath] = projectFile
	if isExeProject || isWebProject(projectNode) {
		project.MarkAsExeProject()
	}

	projectFolderPath := filepath.Dir(projectPath)
	for _, itemGroup := range projectNode.SelectElements("ItemGroup") {
		err := projectLoader.processItemGroup(itemGroup, projectFile, projectFolderPath)
		if err != nil {
			return nil, err
		}
	}
	return projectFile, nil
}

func extractFrameworkAndExeFlag(projectNode *etree.Element) (*etree.Element, bool) {
	var targetFramework *etree.Element
	hasExeOutputType := false
	for _, propertyGroup := range projectNode.SelectElements("PropertyGroup") {
		if targetFramework == nil {
			targetFramework = propertyGroup.SelectElement("TargetFramework")
		}

		if !hasExeOutputType && checkOutputTypeExe(propertyGroup) {
			hasExeOutputType = true
		}

		if targetFramework != nil && hasExeOutputType {
			break
		}
	}
	return targetFramework, hasExeOutputType
}

func checkOutputTypeExe(propertyGroup *etree.Element) bool {
	outputType := propertyGroup.SelectElement("OutputType")
	return outputType != nil && outputType.Text() == "Exe"
}

func isWebProject(projectNode *etree.Element) bool {
	sdk := projectNode.SelectAttrValue("Sdk", "")
	return sdk == "Microsoft.NET.Sdk.Web"
}

func (projectLoader *DotNetProjectLoader) createProjectPackage(projectPath string, framework string) *PackageInfo {
	//In theory this method SHOULD only be called ONCE per project...
	fileName := filepath.Base(projectPath) //This will serve as the package name
	projectPackage := projectLoader.packageContainer.GetOrCreatePackage(fileName, "", framework, projectPath)
	return projectPackage
}

func (projectLoader *DotNetProjectLoader) processItemGroup(itemGroup *etree.Element, projectFile *DotNetProjectFile, projectFolderPath string) error {
	for _, item := range itemGroup.ChildElements() {
		err := projectLoader.processProjectItem(item, projectFile, projectFolderPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (projectLoader *DotNetProjectLoader) processProjectItem(projectItem *etree.Element, projectFile *DotNetProjectFile, projectFolderPath string) error {
	tag := projectItem.Tag
	if tag == "PackageReference" {
		return projectLoader.processPackageReference(projectItem, projectFile)
	}
	if tag == "ProjectReference" {
		return projectLoader.processProjectReference(projectItem, projectFile, projectFolderPath)
	}
	return nil
}

func (projectLoader *DotNetProjectLoader) processPackageReference(packageReference *etree.Element, projectFile *DotNetProjectFile) error {
	packageName := packageReference.SelectAttrValue("Include", "")
	if packageName == "" {
		return fmt.Errorf(`"Include=" attribute is missing in "<PackageReference>" element`)
	}

	version := packageReference.SelectAttrValue("Version", "")
	if projectLoader.hasGlobalPackages {
		localVersion := version

		var exists bool
		version, exists = projectLoader.globalPackages[packageName]
		if !exists {
			return fmt.Errorf(`"%s" package version not found in "Directory.Packages.props"`, packageName)
		}

		if localVersion != "" {
			log.Printf(`[Warning] Found local version "%s" for package "%s", but solution has a "Directory.Packages.props" file\n`, localVersion, packageName)
		}
	}
	project := projectFile.GetProject()
	projectFile.AddPackageRefNode(packageName, packageReference)
	dependency := projectLoader.packageContainer.GetOrCreatePackage(packageName, version, project.Framework, "NOT_LOADED")
	project.AddDependency(dependency)
	return nil
}

func (projectLoader *DotNetProjectLoader) processProjectReference(projectReference *etree.Element, projectFile *DotNetProjectFile, projectFolderPath string) error {
	relativePath := projectReference.SelectAttrValue("Include", "")
	if relativePath == "" {
		return fmt.Errorf(`"Include=" attribute is missing in "<ProjectReference>" element`)
	}

	projectReferencePath := filepath.Join(projectFolderPath, relativePath)
	referencedProject, err := projectLoader.GetOrLoad(projectReferencePath)
	if err != nil {
		return fmt.Errorf(`"%s" failed to load: %w`, projectReferencePath, err)
	}

	project := projectFile.GetProject()
	dependency := referencedProject.GetProject()
	projectFile.AddProjectRefNode(dependency.Name, projectReference)
	project.AddDependency(dependency)
	return nil
}
