package actions

import "fmt"

type RemovePackageAction struct {
	projectName string
	packageInfo *PackageInfo
	recommended bool
	reason      string
}

func NewRemovePackageAction(projectName string, packageInfo *PackageInfo, recommended bool, reason string) *RemovePackageAction {
	return &RemovePackageAction{
		projectName: projectName,
		packageInfo: packageInfo,
		recommended: recommended,
		reason:      reason,
	}
}

func (action *RemovePackageAction) GetReason() string {
	return action.reason
}

func (action *RemovePackageAction) GetDescription() string {
	description := fmt.Sprintf(`Remove package "%s" from "%s"`, action.packageInfo.Name, action.projectName)
	return description
}

func (action *RemovePackageAction) IsRecommended() bool {
	return action.recommended
}

func (action *RemovePackageAction) Execute(projectHandler ProjectHandler) error {
	projectHandler.RemoveDependency(action.projectName, action.packageInfo)
	return nil
}
