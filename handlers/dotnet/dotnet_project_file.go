package dotnet

import (
	"log"
	"os"
	"path/filepath"
	"redun-pendancy/helpers"
	"redun-pendancy/utils"

	"github.com/beevik/etree"
)

type DotNetProjectFile struct {
	project             *PackageInfo
	xmlFile             *helpers.XMLFileHelper
	folderPath          string
	packageRefNodes     *helpers.OrderedMap[string, *etree.Element] //PackageName => Node
	projectRefNodes     *helpers.OrderedMap[string, *etree.Element] //ProjectName => Node
	addedDependencies   []*PackageInfo
	removedDependencies []*PackageInfo
	isDirty             bool
}

func NewDotNetProjectFile(project *PackageInfo, xmlFile *helpers.XMLFileHelper) *DotNetProjectFile {
	return &DotNetProjectFile{
		project:         project,
		xmlFile:         xmlFile,
		folderPath:      filepath.Dir(project.FilePath),
		packageRefNodes: helpers.NewOrderedMap[string, *etree.Element](),
		projectRefNodes: helpers.NewOrderedMap[string, *etree.Element](),
	}
}

func (projectFile *DotNetProjectFile) GetProject() *PackageInfo {
	return projectFile.project
}

func (projectFile *DotNetProjectFile) HasChanges() bool {
	return projectFile.isDirty
}

func (projectFile *DotNetProjectFile) AddPackageRefNode(packageName string, node *etree.Element) {
	projectFile.packageRefNodes.Set(packageName, node)
}

func (projectFile *DotNetProjectFile) AddProjectRefNode(projectName string, node *etree.Element) {
	projectFile.projectRefNodes.Set(projectName, node)
}

func (projectFile *DotNetProjectFile) DividesProjectsAndPackages() bool {
	packageRefNodes := projectFile.packageRefNodes
	projectRefNodes := projectFile.projectRefNodes
	if packageRefNodes.GetCount() == 0 || projectRefNodes.GetCount() == 0 {
		return true
	}

	packageNode := packageRefNodes.GetAt(0)
	projectNode := projectRefNodes.GetAt(0)
	return packageNode.Parent() != projectNode.Parent()
}

func (projectFile *DotNetProjectFile) SortDependency(dependency *PackageInfo) bool {
	nodesMap := projectFile.getRefNodesMap(dependency)
	node, exists := nodesMap.Get(dependency.Name)
	if !exists {
		return false
	}

	xmlFile := projectFile.xmlFile
	xmlFile.RemoveNode(node, true, false)

	oldIndex, _ := nodesMap.GetIndex(dependency.Name)
	newIndex := projectFile.addDependency(node, dependency.Name, nodesMap)
	if newIndex != -1 {
		nodesMap.MoveIndex(oldIndex, newIndex)
	}
	projectFile.isDirty = true
	return true
}

func (projectFile *DotNetProjectFile) getRefNodesMap(dependency *PackageInfo) *helpers.OrderedMap[string, *etree.Element] {
	if dependency.IsProject() {
		return projectFile.projectRefNodes
	}
	return projectFile.packageRefNodes
}

func (projectFile *DotNetProjectFile) AddDependency(dependency *PackageInfo, hasGlobalPackages bool) bool {
	node, nodesMap := projectFile.createDependencyNode(dependency, hasGlobalPackages)
	_, exists := nodesMap.Get(dependency.Name)
	if exists {
		return false
	}

	projectFile.addedDependencies = append(projectFile.addedDependencies, dependency)
	projectFile.addDependency(node, dependency.Name, nodesMap)
	projectFile.project.AddDependency(dependency)
	nodesMap.Set(dependency.Name, node)
	projectFile.isDirty = true
	return true
}

func (projectFile *DotNetProjectFile) createDependencyNode(dependency *PackageInfo, hasGlobalPackages bool) (*etree.Element, *helpers.OrderedMap[string, *etree.Element]) {
	if dependency.IsProject() {
		node := createProjectReference(dependency, projectFile.folderPath)
		return node, projectFile.projectRefNodes
	}
	node := createPackageReference(dependency, hasGlobalPackages)
	return node, projectFile.packageRefNodes
}

func createProjectReference(dependency *PackageInfo, projectDirPath string) *etree.Element {
	relativePath, err := filepath.Rel(projectDirPath, dependency.FilePath)
	if err != nil {
		log.Printf(`[Warning] Failed to compute relative path from "%s" to "%s": %v`, projectDirPath, dependency.FilePath, err)
		relativePath = dependency.FilePath //Fallback to absolute path
	}
	projectRef := etree.NewElement("ProjectReference")
	projectRef.CreateAttr("Include", relativePath)
	return projectRef
}

func createPackageReference(dependency *PackageInfo, hasGlobalPackages bool) *etree.Element {
	packageRef := etree.NewElement("PackageReference")
	packageRef.CreateAttr("Include", dependency.Name)
	if !hasGlobalPackages {
		packageRef.CreateAttr("Version", dependency.Version)
	}
	return packageRef
}

// Adds the node to the document AND returns the insert index (or -1 if it was appended)
func (projectFile *DotNetProjectFile) addDependency(node *etree.Element, dependencyName string, nodesMap *helpers.OrderedMap[string, *etree.Element]) int {
	if nodesMap.GetCount() == 0 {
		//Project has NO references of this "type"! Add a new "ItemGroup" section to the file
		projectFile.createNewItemGroup(node)
		return -1
	}

	xmlFile := projectFile.xmlFile
	newLineNode := createNewLineNode()
	orderedKeys := nodesMap.GetOrderedKeys()
	index := utils.IndexOf(orderedKeys, 0, func(packageName string) bool {
		return packageName > dependencyName
	})
	if index != -1 {
		siblingNode := nodesMap.GetAt(index)
		xmlFile.InsertNode(newLineNode, siblingNode)
		xmlFile.InsertNode(node, newLineNode)
		return index
	}

	lastIndex := nodesMap.GetCount() - 1
	lastNode := nodesMap.GetAt(lastIndex)
	parent := lastNode.Parent()
	xmlFile.AppendNode(node, parent)
	xmlFile.AppendNode(newLineNode, parent)
	return -1
}

func (projectFile *DotNetProjectFile) createNewItemGroup(node *etree.Element) {
	itemGroup := etree.NewElement("ItemGroup")
	itemGroup.AddChild(createNewLineNode())
	itemGroup.AddChild(node)
	itemGroup.AddChild(createNewLineNode())

	projectNode := projectFile.xmlFile.Document.SelectElement("Project")
	projectNode.AddChild(createNewLineNode())
	projectNode.AddChild(itemGroup)
	projectNode.AddChild(createNewLineNode())
}

func createNewLineNode() *etree.CharData {
	return etree.NewText("\n")
}

func (projectFile *DotNetProjectFile) RemoveDependency(dependency *PackageInfo) bool {
	targetMap := utils.TernarySelect(dependency.IsProject(),
		projectFile.projectRefNodes,
		projectFile.packageRefNodes,
	)
	node, exists := targetMap.Get(dependency.Name)
	if !exists {
		return false
	}

	projectFile.removedDependencies = append(projectFile.removedDependencies, dependency)
	projectFile.project.RemoveDependency(dependency.Name)
	projectFile.xmlFile.RemoveNode(node, true, true)
	targetMap.Remove(dependency.Name)
	projectFile.isDirty = true
	return true
}

func (projectFile *DotNetProjectFile) RevertChanges(hasGlobalPackages bool) {
	//Back up removed dependencies to avoid re-removal afterwards :)
	removedDependencies := projectFile.removedDependencies
	projectFile.removedDependencies = nil
	for _, dependency := range projectFile.addedDependencies {
		projectFile.RemoveDependency(dependency)
	}
	for _, dependency := range removedDependencies {
		projectFile.AddDependency(dependency, hasGlobalPackages)
	}
	projectFile.resetTracking()
}

func (projectFile *DotNetProjectFile) Commit() error {
	if !projectFile.isDirty {
		return nil
	}

	filePath := projectFile.xmlFile.FilePath
	log.Println("Writing:", filePath)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	xmlWriter := helpers.NewCustomXMLWriter(file, "  ", true)
	err = projectFile.xmlFile.Commit(xmlWriter)
	if err != nil {
		return err
	}

	projectFile.resetTracking()
	return nil
}

func (projectFile *DotNetProjectFile) resetTracking() {
	projectFile.removedDependencies = nil
	projectFile.addedDependencies = nil
	projectFile.isDirty = false
}
