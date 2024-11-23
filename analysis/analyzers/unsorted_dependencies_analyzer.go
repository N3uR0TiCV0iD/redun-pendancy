package analyzers

import (
	"redun-pendancy/analysis/actions"
	"redun-pendancy/models"
)

type UnsortedDependenciesAnalyzer struct {
	results        *AnalysisResults
	projectHandler ProjectHandler
}

func NewUnsortedDependenciesAnalyzer(collector *AnalysisResults, projectHandler ProjectHandler) *UnsortedDependenciesAnalyzer {
	return &UnsortedDependenciesAnalyzer{
		results:        collector,
		projectHandler: projectHandler,
	}
}

func (analyzer *UnsortedDependenciesAnalyzer) Analyze(projects []*PackageInfo, packages map[string]*PackageInfo) {
	for _, project := range projects {
		if len(project.Dependencies) == 0 {
			//Project has no dependencies, next please
			continue
		}

		dividesDependencies := analyzer.projectHandler.DividesProjectsAndPackages(project.Name)
		if !dividesDependencies {
			analyzer.addSortIfNeeded(project, models.PackageType_Any)
			continue
		}
		analyzer.addSortIfNeeded(project, models.PackageType_Package)
		analyzer.addSortIfNeeded(project, models.PackageType_Project)
	}
}

func (analyzer *UnsortedDependenciesAnalyzer) addSortIfNeeded(project *PackageInfo, packageType models.PackageType) {
	unsortedDependencies := analyzer.getUnsortedDependencies(project, packageType)
	if len(unsortedDependencies) != 0 {
		action := actions.NewSortDependenciesAction(project.Name, packageType, unsortedDependencies)
		analyzer.results.AddAction(action)
	}
}

func (analyzer *UnsortedDependenciesAnalyzer) getUnsortedDependencies(project *PackageInfo, packageType models.PackageType) []*PackageInfo {
	var unsortedDependencies []*PackageInfo
	var lastDependency *PackageInfo
	lastPackageName := ""
	for _, dependency := range project.Dependencies {
		if !dependency.MatchesType(packageType) {
			continue
		}

		if dependency.Name < lastPackageName {
			unsortedDependencies = append(unsortedDependencies, lastDependency)
		}
		lastPackageName = dependency.Name
		lastDependency = dependency
	}
	return unsortedDependencies
}
