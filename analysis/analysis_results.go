package analysis

type AnalysisResults struct {
	actions     []ProjectAction
	suggestions map[string][]string
}

func NewAnalysisResults() *AnalysisResults {
	return &AnalysisResults{
		suggestions: make(map[string][]string),
	}
}

func (results *AnalysisResults) AddAction(action ProjectAction) {
	results.actions = append(results.actions, action)
}

func (results *AnalysisResults) AddSuggestion(projectName string, suggestion string) {
	results.suggestions[projectName] = append(results.suggestions[projectName], suggestion)
}

func (results *AnalysisResults) GetActions() []ProjectAction {
	return results.actions
}

func (results *AnalysisResults) GetSuggestions() map[string][]string {
	return results.suggestions
}
