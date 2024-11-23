package analyzers

import (
	"fmt"
	"redun-pendancy/analysis"
	"redun-pendancy/analysis/actions"
	"redun-pendancy/utils"
)

type BubbleUpAnalyzer struct {
	results *AnalysisResults
}

func NewBubbleUpAnalyzer(collector *AnalysisResults) *BubbleUpAnalyzer {
	return &BubbleUpAnalyzer{
		results: collector,
	}
}

func (analyzer *BubbleUpAnalyzer) Analyze(projects []*PackageInfo, packages map[string]*PackageInfo) {
	if len(projects) <= 2 {
		//There must be 3 or more projects in order to bubble up dependencies
		return
	}

	ancestryMap := analysis.NewAncestryMap(projects)
	if ancestryMap.AreAllAncestorsEmpty() {
		return
	}

	//Build a map of "Dependency => []Project"
	dependencyProjects := make(map[*PackageInfo][]*PackageInfo)
	for _, project := range projects {
		analyzer.buildDependencyProjects(dependencyProjects, project)
	}

	for dependency, projectList := range dependencyProjects {
		if len(projectList) <= 1 {
			//Dependency is only used in 1 project, next please
			continue
		}

		commonAncestors := ancestryMap.GetCommonAncestors(projectList)
		if len(commonAncestors) == 0 {
			fmt.Printf("No common ancestors found for dependency \"%s\" among projects\n", dependency.ToString())
			continue
		}

		analyzer.addBubbleUpActions(dependency, commonAncestors, projectList)
	}
}

func (analyzer *BubbleUpAnalyzer) buildDependencyProjects(dependencyProjects map[*PackageInfo][]*PackageInfo, project *PackageInfo) {
	for _, dependency := range project.Dependencies {
		if dependency.IsTool() || dependency.IsExeProject() {
			//Do not consider tools or executable projects
			continue
		}
		projectList := dependencyProjects[dependency]
		dependencyProjects[dependency] = append(projectList, project)
	}
}

func (analyzer *BubbleUpAnalyzer) addBubbleUpActions(dependency *PackageInfo, commonAncestors utils.Set[*PackageInfo], projectList []*PackageInfo) {
	for ancestor := range commonAncestors {
		if ancestor == dependency {
			//We can't bubble up the project dependency to itself. That would just be silly, wouldn't it? ;)
			continue
		}

		if ancestor.ContainsDependency(dependency.Name) {
			//The ancestor already contains the dependency. Don't add an action to bubble it there
			continue
		}

		if dependency.ContainsDependency(ancestor.Name) {
			//The dependency contains a reference to the ancestor. Don't add an action to bubble it there
			continue
		}

		projectNames := extractProjectNames(projectList)
		action := actions.NewBubbleUpAction(projectNames, ancestor.Name, dependency)
		analyzer.results.AddAction(action)
	}
}

func extractProjectNames(projects []*PackageInfo) []string {
	projectNames := utils.Map(projects, func(project *PackageInfo) string {
		return project.Name
	})
	return projectNames
}
