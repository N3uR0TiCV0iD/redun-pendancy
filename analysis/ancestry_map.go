package analysis

import (
	"redun-pendancy/utils"
	"sort"
	"strings"
)

type AncestryMap struct {
	ancestors            map[*PackageInfo]utils.Set[*PackageInfo] //Project => Set[Project]
	commonAncestorsCache map[string]utils.Set[*PackageInfo]       //Memoization cache
}

func NewAncestryMap(projects []*PackageInfo) *AncestryMap {
	ancestryMap := &AncestryMap{
		ancestors:            make(map[*PackageInfo]utils.Set[*PackageInfo]),
		commonAncestorsCache: make(map[string]utils.Set[*PackageInfo]),
	}
	for _, project := range projects {
		ancestryMap.processAncestors(project)
	}
	return ancestryMap
}

func (ancestryMap *AncestryMap) processAncestors(project *PackageInfo) utils.Set[*PackageInfo] {
	projectAncestors, exists := ancestryMap.ancestors[project]
	if exists {
		return projectAncestors
	}

	projectAncestors = utils.NewSet[*PackageInfo]()
	for _, parent := range project.Dependencies {
		if !parent.IsProject() {
			continue
		}

		projectAncestors.Add(parent)
		parentAncestors := ancestryMap.processAncestors(parent)
		projectAncestors.UnionWith(parentAncestors)
	}
	ancestryMap.ancestors[project] = projectAncestors
	return projectAncestors
}

func (ancestryMap *AncestryMap) GetAncestors(project *PackageInfo) utils.Set[*PackageInfo] {
	return ancestryMap.ancestors[project]
}

func (ancestryMap *AncestryMap) GetCommonAncestors(projects []*PackageInfo) utils.Set[*PackageInfo] {
	if len(projects) == 0 {
		return nil
	}

	startingAncestors := ancestryMap.GetAncestors(projects[0])
	if len(projects) == 1 {
		return startingAncestors
	}

	cacheKey := getCacheKey(projects)
	cachedResult, exists := ancestryMap.commonAncestorsCache[cacheKey]
	if exists {
		return cachedResult
	}

	commonAncestors := utils.NewSet[*PackageInfo]()
	commonAncestors.UnionWith(startingAncestors)
	for _, project := range projects[1:] {
		projectAncestors := ancestryMap.GetAncestors(project)
		commonAncestors.IntersectWith(projectAncestors)
		if commonAncestors.IsEmpty() {
			break
		}
	}
	ancestryMap.commonAncestorsCache[cacheKey] = commonAncestors
	return commonAncestors
}

func getCacheKey(projects []*PackageInfo) string {
	names := make([]string, len(projects))
	for index, project := range projects {
		names[index] = project.Name
	}
	sort.Strings(names)
	return strings.Join(names, "|")
}

func (ancestryMap *AncestryMap) AreAllAncestorsEmpty() bool {
	for _, projectAncestors := range ancestryMap.ancestors {
		if !projectAncestors.IsEmpty() {
			return false
		}
	}
	return true
}
