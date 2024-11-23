package dotnet

import (
	"path/filepath"
	"redun-pendancy/utils"
	"strings"
)

func extractProjectPaths(solutionFilePath string) ([]string, error) {
	projectPaths := []string{}
	folderPath := filepath.Dir(solutionFilePath)
	err := utils.OpenAndProcessFileLines(solutionFilePath, func(line string) error {
		relativePath := processSolutionLine(line)
		if relativePath != "" {
			projectPath := filepath.Join(folderPath, relativePath)
			projectPaths = append(projectPaths, projectPath)
		}
		return nil
	})
	return projectPaths, err
}

func processSolutionLine(line string) string {
	if !strings.HasPrefix(line, `Project("{`) {
		return ""
	}
	return extractProjectPath(line)
}

func extractProjectPath(line string) string {
	parts := strings.Split(line, "\"")
	projectPath := parts[5]
	if !isValidProjectPath(projectPath) {
		return ""
	}
	return projectPath
}

func isValidProjectPath(projectPath string) bool {
	return strings.HasSuffix(projectPath, ".csproj") ||
		strings.HasSuffix(projectPath, ".vbproj")
}
