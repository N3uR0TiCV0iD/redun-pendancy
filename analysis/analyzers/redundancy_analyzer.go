package analyzers

import (
	"fmt"
	"redun-pendancy/analysis/actions"
	"redun-pendancy/utils"
)

type RedundancyAnalyzer struct {
	results *AnalysisResults
}

func NewRedundancyAnalyzer(collector *AnalysisResults) *RedundancyAnalyzer {
	return &RedundancyAnalyzer{
		results: collector,
	}
}

func (analyzer *RedundancyAnalyzer) Analyze(projects []*PackageInfo, packages map[string]*PackageInfo) {
	for _, project := range projects {
		analyzer.findRedundantProjectDependencies(project)
	}
}

func (analyzer *RedundancyAnalyzer) findRedundantProjectDependencies(project *PackageInfo) {
	for _, dependency := range project.Dependencies {
		if dependency.IsTool() {
			continue
		}
		analyzer.checkProjectDependencyForRedundancy(project, dependency)
	}
}

func (analyzer *RedundancyAnalyzer) checkProjectDependencyForRedundancy(project *PackageInfo, targetDependency *PackageInfo) {
	for _, dependency := range project.Dependencies {
		if dependency == targetDependency {
			continue
		}

		seenPackages := utils.NewSet[*PackageInfo]()
		redundantPath, redundantDependency := analyzer.getDependencyPath(dependency, targetDependency.Name, seenPackages)
		if redundantPath != "" {
			analyzer.addRemovePackageAction(project, targetDependency, redundantPath, redundantDependency)
			break
		}
	}
}

func (analyzer *RedundancyAnalyzer) getDependencyPath(packageInfo *PackageInfo, packageName string, seenPackages utils.Set[*PackageInfo]) (string, *PackageInfo) {
	for _, dependency := range packageInfo.Dependencies {
		if dependency.Name == packageName {
			return packageInfo.Name, dependency
		}

		if !seenPackages.Add(dependency) {
			continue
		}

		dependencyPath, foundDependency := analyzer.getDependencyPath(dependency, packageName, seenPackages)
		if foundDependency != nil {
			return packageInfo.Name + "/" + dependencyPath, foundDependency
		}
	}
	return "", nil
}

func (analyzer *RedundancyAnalyzer) addRemovePackageAction(project *PackageInfo, targetDependency *PackageInfo, redundantPath string, redundantDependency *PackageInfo) {
	reason := fmt.Sprintf(`Dependency "%s" (%s) is already included in: "%s" (uses: %s)`,
		targetDependency.Name, formatDependencyVersion(targetDependency),
		redundantPath, formatDependencyVersion(redundantDependency),
	)
	couldKeep := utils.IsVersionHigher(targetDependency.Version, redundantDependency.Version)
	action := actions.NewRemovePackageAction(project.Name, targetDependency, !couldKeep, reason)
	analyzer.results.AddAction(action)
}

func formatDependencyVersion(dependency *PackageInfo) string {
	if dependency.IsProject() {
		return dependency.Framework
	}
	return fmt.Sprintf("%s [%s]", dependency.Version, dependency.Framework)
}
