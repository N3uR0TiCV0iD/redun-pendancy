package actions

import (
	"fmt"
	"redun-pendancy/models"
)

type SortDependenciesAction struct {
	projectName      string
	packageType      models.PackageType
	dependencies     []*PackageInfo
	dependenciesType string
}

func NewSortDependenciesAction(projectName string, packageType models.PackageType, dependencies []*PackageInfo) *SortDependenciesAction {
	return &SortDependenciesAction{
		projectName:      projectName,
		packageType:      packageType,
		dependencies:     dependencies,
		dependenciesType: getDependenciesType(packageType),
	}
}

func getDependenciesType(packageType models.PackageType) string {
	if packageType == models.PackageType_Any {
		return "dependencies"
	}
	if packageType == models.PackageType_Package {
		return "package dependencies"
	}
	return "project dependencies"
}

func (action *SortDependenciesAction) GetReason() string {
	return fmt.Sprintf(`The project "%s" has unsorted %s`, action.projectName, action.dependenciesType)
}

func (action *SortDependenciesAction) GetDescription() string {
	return fmt.Sprintf(`Sort %s in "%s"`, action.dependenciesType, action.projectName)
}

func (action *SortDependenciesAction) IsRecommended() bool {
	return true
}

func (action *SortDependenciesAction) Execute(projectHandler ProjectHandler) error {
	projectHandler.SortDependencies(action.projectName, action.dependencies)
	return nil
}
