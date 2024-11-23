package analyzers

import (
	"redun-pendancy/models"
	"redun-pendancy/utils"
	"sort"
	"strings"
)

type UpgradeAnalyzer struct {
	results            *AnalysisResults
	upgradeSuggestions map[*PackageInfo][]UpgradeSuggestion
}

type UpgradeSuggestion struct {
	Project        *PackageInfo
	Dependency     *PackageInfo
	SkippedPackage *PackageInfo
}

func NewUpgradeAnalyzer(collector *AnalysisResults) *UpgradeAnalyzer {
	return &UpgradeAnalyzer{
		results:            collector,
		upgradeSuggestions: make(map[*models.PackageInfo][]UpgradeSuggestion),
	}
}

func (analyzer *UpgradeAnalyzer) Analyze(projects []*PackageInfo, packages map[string]*PackageInfo) {
	skippedPackages := analyzer.collectSkippedPackages(packages)

	for _, skippedPackage := range skippedPackages {
		seenPackages := utils.NewSet[*PackageInfo]()
		seenPackages.Add(skippedPackage) //Might not be necesary (since it should be a "leaf node"), but just in case :)
		analyzer.traverseUpForSuggestions(skippedPackage, skippedPackage, seenPackages)
	}

	for project, suggestedUpgrades := range analyzer.upgradeSuggestions {
		analyzer.addUpgradeSuggestions(project, suggestedUpgrades)
	}
}

func (analyzer *UpgradeAnalyzer) collectSkippedPackages(packages map[string]*PackageInfo) []*PackageInfo {
	var skippedPackages []*PackageInfo
	for _, packageInfo := range packages {
		if packageInfo.LoadStatus == models.LoadStatus_Skipped {
			skippedPackages = append(skippedPackages, packageInfo)
		}
	}
	return skippedPackages
}

func (analyzer *UpgradeAnalyzer) traverseUpForSuggestions(packageInfo *PackageInfo, skippedPackage *PackageInfo, seenPackages utils.Set[*PackageInfo]) {
	for _, parent := range packageInfo.Parents {
		if !seenPackages.Add(parent) {
			continue
		}
		if parent.IsProject() {
			suggestionList := analyzer.upgradeSuggestions[parent]
			analyzer.upgradeSuggestions[parent] = append(suggestionList, UpgradeSuggestion{
				Project:        parent,
				Dependency:     packageInfo,
				SkippedPackage: skippedPackage,
			})
			continue
		}
		analyzer.traverseUpForSuggestions(parent, skippedPackage, seenPackages)
	}
}

func (analyzer *UpgradeAnalyzer) addUpgradeSuggestions(project *PackageInfo, suggestedUpgrades []UpgradeSuggestion) {
	projectName := project.Name
	dependencyGroups := utils.GroupBy(suggestedUpgrades, func(suggestion UpgradeSuggestion) *PackageInfo {
		return suggestion.Dependency
	})
	for dependency, suggestions := range dependencyGroups {
		suggestion := formatUpgradeSuggestion(dependency, suggestions)
		analyzer.results.AddSuggestion(projectName, suggestion)
	}
}

func formatUpgradeSuggestion(dependency *PackageInfo, suggestions []UpgradeSuggestion) string {
	builder := strings.Builder{}
	builder.WriteString("Check updates for \"")
	builder.WriteString(dependency.Name)
	builder.WriteString("\". It contains older package versions:\n")

	skippedPackages := utils.Map(suggestions, func(suggestion UpgradeSuggestion) string {
		return suggestion.SkippedPackage.ToString()
	})
	sort.Strings(skippedPackages)

	for _, skippedPackage := range skippedPackages {
		builder.WriteString("- ")
		builder.WriteString(skippedPackage)
		builder.WriteString("\n")
	}
	return builder.String()
}
