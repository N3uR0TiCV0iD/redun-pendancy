package gui

import (
	"redun-pendancy/helpers"
	"redun-pendancy/models"
	"redun-pendancy/utils"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TreeNode struct {
	PackageInfo *PackageInfo
	ChildPaths  []string
}

type DependencyTree struct {
	widget            *widget.Tree
	nodes             map[string]TreeNode //nodePath => TreeNode
	filteredPaths     map[string][]string //nodePath => childPaths (filtered)
	selectionCallback func(*PackageInfo)
}

func NewDependencyTree() *DependencyTree {
	dt := &DependencyTree{
		nodes: make(map[string]TreeNode),
	}
	dt.widget = &widget.Tree{
		ChildUIDs:  dt.getChildUIDs,
		IsBranch:   dt.isBranch,
		CreateNode: dt.createNode,
		UpdateNode: dt.updateNode,
		OnSelected: dt.onNodeSelected,
	}
	return dt
}

func (dt *DependencyTree) GetWidget() *widget.Tree {
	return dt.widget
}

func (dt *DependencyTree) SetSelectionCallback(callback func(*PackageInfo)) {
	dt.selectionCallback = callback
}

func (dt *DependencyTree) Load(projectHandler ProjectHandler) {
	tracker := helpers.NewCircularDependencyTracker()
	projects := projectHandler.GetProjects()
	dt.nodes[""] = TreeNode{
		PackageInfo: nil, //Root node has no PackageInfo
		ChildPaths:  dt.buildPackagePaths("", projects, tracker),
	}
}

func (dt *DependencyTree) CollapseAll() {
	dt.widget.CloseAllBranches()
}

func (dt *DependencyTree) SearchNodes(searchTerm string) []string {
	var matchingPaths []string
	searchTerm = strings.ToLower(searchTerm)
	for nodePath, node := range dt.nodes {
		if node.PackageInfo == nil {
			//Technically shouldn't happen...
			continue
		}

		packageName := strings.ToLower(node.PackageInfo.Name)
		if strings.Contains(packageName, searchTerm) {
			matchingPaths = append(matchingPaths, nodePath)
		}
	}
	sort.Strings(matchingPaths)
	return matchingPaths
}

func (dt *DependencyTree) ExpandNode(nodePath string) {
	dt.widget.OpenBranch(nodePath)
}

func (dt *DependencyTree) ExpandPath(nodePath string) {
	widget := dt.widget
	processPathSegments(nodePath, func(pathSegment string) {
		widget.OpenBranch(pathSegment)
	})
}

func processPathSegments(nodePath string, action func(string)) {
	currentPath := ""
	pathSegments := strings.Split(nodePath, "/")
	for _, segment := range pathSegments {
		//Skip empty segments (eg: handling the root or a leading slash)
		if segment == "" {
			continue
		}
		currentPath += "/" + segment
		action(currentPath)
	}
}

func (dt *DependencyTree) SelectPath(nodePath string) {
	dt.widget.Select(nodePath)
	dt.widget.ScrollTo(nodePath)
}

func (dt *DependencyTree) GetFilter() []string {
	filters := utils.GetMapKeys(dt.filteredPaths)
	return filters
}

func (dt *DependencyTree) SetFilter(nodePaths []string) {
	if len(nodePaths) == 0 {
		dt.filteredPaths = nil
		return
	}

	seenPaths := utils.NewSet[string]()
	filteredPaths := make(map[string][]string)
	for _, nodePath := range nodePaths {
		parentPath := ""
		processPathSegments(nodePath, func(pathSegment string) {
			if seenPaths.Add(pathSegment) {
				filteredPaths[parentPath] = append(filteredPaths[parentPath], pathSegment)
			}
			parentPath = pathSegment
		})
	}
	dt.filteredPaths = filteredPaths
}

// NOTE: Also populates the nodes in the dependency tree
func (dt *DependencyTree) buildPackagePaths(parentPath string, packages []*PackageInfo, tracker *helpers.CircularDependencyTracker) []string {
	packagePaths := make([]string, 0, len(packages))
	for _, packageInfo := range packages {
		isProject := packageInfo.IsProject()
		if isProject && !tracker.RegisterProject(packageInfo.Name, parentPath) {
			continue
		}

		packagePath := parentPath + "/" + packageInfo.Name
		dt.nodes[packagePath] = TreeNode{
			PackageInfo: packageInfo,
			ChildPaths:  dt.buildPackagePaths(packagePath, packageInfo.Dependencies, tracker),
		}
		packagePaths = append(packagePaths, packagePath)

		if isProject {
			tracker.RemoveProject(packageInfo.Name)
		}
	}
	sort.Strings(packagePaths)
	return packagePaths
}

// =================================
// Implementations for "widget.Tree"
// =================================
func (dt *DependencyTree) getChildUIDs(nodePath string) []string {
	if dt.filteredPaths == nil {
		return dt.nodes[nodePath].ChildPaths
	}
	return dt.filteredPaths[nodePath]
}

func (dt *DependencyTree) isBranch(nodePath string) bool {
	node, exists := dt.nodes[nodePath]
	return exists && len(node.ChildPaths) != 0
}

func (dt *DependencyTree) createNode(isBranch bool) fyne.CanvasObject {
	return widget.NewLabel("Node")
}

func (dt *DependencyTree) updateNode(nodePath string, branch bool, node fyne.CanvasObject) {
	packageInfo := dt.nodes[nodePath].PackageInfo
	if packageInfo != nil {
		label := node.(*widget.Label)
		if packageInfo.LoadStatus == models.LoadStatus_Skipped {
			updateSkippedNode(label, packageInfo)
			return
		}
		updateNormalNode(label, packageInfo)
	}
}

func updateNormalNode(label *widget.Label, packageInfo *PackageInfo) {
	label.SetText(packageInfo.ToString())
	label.Importance = widget.MediumImportance
	label.TextStyle = fyne.TextStyle{}
}

func updateSkippedNode(label *widget.Label, packageInfo *PackageInfo) {
	label.SetText("(~) " + packageInfo.ToString())
	label.Importance = widget.WarningImportance
	label.TextStyle = fyne.TextStyle{
		Italic: true,
	}
}

func (dt *DependencyTree) onNodeSelected(nodePath string) {
	if dt.selectionCallback != nil {
		packageInfo := dt.nodes[nodePath].PackageInfo
		dt.selectionCallback(packageInfo)
	}
}
