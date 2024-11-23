package analyzers

import (
	"redun-pendancy/analysis/actions"
	"redun-pendancy/utils"
)

type UnusedGlobalPackagesAnalyzer struct {
	results        *AnalysisResults
	globalPackages map[string]string //PackageName => Version
}

func NewUnusedGlobalPackagesAnalyzer(results *AnalysisResults, projectHandler ProjectHandler) *UnusedGlobalPackagesAnalyzer {
	return &UnusedGlobalPackagesAnalyzer{
		results:        results,
		globalPackages: projectHandler.GetGlobalPackages(),
	}
}

func (analyzer *UnusedGlobalPackagesAnalyzer) Analyze(projects []*PackageInfo, packages map[string]*PackageInfo) {
	usedPackages := analyzer.collectUsedPackages(projects)
	analyzer.addRemoveActions(usedPackages)
}

func (analyzer *UnusedGlobalPackagesAnalyzer) collectUsedPackages(projects []*PackageInfo) utils.Set[string] {
	usedPackages := utils.NewSet[string]()
	for _, project := range projects {
		for _, dependency := range project.Dependencies {
			usedPackages.Add(dependency.Name)
		}
	}
	return usedPackages
}

func (analyzer *UnusedGlobalPackagesAnalyzer) addRemoveActions(usedPackages utils.Set[string]) {
	for packageName := range analyzer.globalPackages {
		if !usedPackages.Contains(packageName) {
			action := actions.NewRemoveGlobalPackageAction(packageName)
			analyzer.results.AddAction(action)
		}
	}
}
