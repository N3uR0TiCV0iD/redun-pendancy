package actions

import "fmt"

type RemoveGlobalPackageAction struct {
	packageName string
}

func NewRemoveGlobalPackageAction(packageName string) *RemoveGlobalPackageAction {
	return &RemoveGlobalPackageAction{
		packageName: packageName,
	}
}

func (action *RemoveGlobalPackageAction) GetReason() string {
	return fmt.Sprintf(`The global package "%s" is not used in any project`, action.packageName)
}

func (action *RemoveGlobalPackageAction) GetDescription() string {
	return fmt.Sprintf(`Remove unused global package "%s"`, action.packageName)
}

func (action *RemoveGlobalPackageAction) IsRecommended() bool {
	return true
}

func (action *RemoveGlobalPackageAction) Execute(projectHandler ProjectHandler) error {
	projectHandler.RemoveGlobalPackage(action.packageName)
	return nil
}
