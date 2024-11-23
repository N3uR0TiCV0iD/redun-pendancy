package actions

import (
	"fmt"
	"strings"
)

type BubbleUpAction struct {
	to             string
	from           []string
	dependency     *PackageInfo
	projectsString string
}

func NewBubbleUpAction(from []string, to string, dependency *PackageInfo) *BubbleUpAction {
	return &BubbleUpAction{
		to:             to,
		from:           from,
		dependency:     dependency,
		projectsString: strings.Join(from, ", "),
	}
}

func (action *BubbleUpAction) GetReason() string {
	reason := fmt.Sprintf(`Projects [%s] have the common ancestor "%s"`, action.projectsString, action.to)
	return reason
}

func (action *BubbleUpAction) GetDescription() string {
	description := fmt.Sprintf(`Bubble up package "%s" to "%s"`, action.dependency.Name, action.to)
	return description
}

func (action *BubbleUpAction) IsRecommended() bool {
	return false
}

func (action *BubbleUpAction) Execute(projectHandler ProjectHandler) error {
	for _, projectName := range action.from {
		projectHandler.RemoveDependency(projectName, action.dependency)
	}
	projectHandler.AddDependency(action.to, action.dependency)
	return nil
}
