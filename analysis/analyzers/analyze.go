package analyzers

import (
	"fmt"
	"redun-pendancy/analysis"
)

type Analyzer interface {
	Analyze(projects []*PackageInfo, packages map[string]*PackageInfo)
}

func AnalyzeProject(projectHandler ProjectHandler) *AnalysisResults {
	results := analysis.NewAnalysisResults()
	analyzers := []Analyzer{
		NewUnusedGlobalPackagesAnalyzer(results, projectHandler),
		NewUnsortedDependenciesAnalyzer(results, projectHandler),
		NewRedundancyAnalyzer(results),
		NewBubbleUpAnalyzer(results),
		NewUpgradeAnalyzer(results),
	}

	projects := projectHandler.GetProjects()
	packages := projectHandler.GetPackageContainer().GetPackages()
	for _, analyzer := range analyzers {
		analyzer.Analyze(projects, packages)
	}
	fmt.Println()
	return results
}
