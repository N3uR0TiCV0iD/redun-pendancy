package actions

type ProjectAction interface {
	GetReason() string
	GetDescription() string

	IsRecommended() bool

	Execute(projectHandler ProjectHandler) error
}
