package helpers

import (
	"log"
	"path/filepath"
	"redun-pendancy/utils"
)

type CircularDependencyTracker struct {
	visitedProjects utils.Set[string]
	detectedCycles  utils.Set[string]
}

func NewCircularDependencyTracker() *CircularDependencyTracker {
	return &CircularDependencyTracker{
		visitedProjects: utils.NewSet[string](),
		detectedCycles:  utils.NewSet[string](),
	}
}

func (tracker *CircularDependencyTracker) RegisterProject(projectName string, parentPath string) bool {
	if tracker.visitedProjects.Add(projectName) {
		//First time we see this project
		return true
	}

	parentProject := filepath.Base(parentPath)
	circularDependency := parentProject + " <=> " + projectName
	newCycle := tracker.detectedCycles.Add(circularDependency)
	if newCycle {
		//Only log this once
		log.Println("[Warning] Circular dependency detected:", circularDependency)
		reverseDependency := projectName + " <=> " + parentProject
		tracker.detectedCycles.Add(reverseDependency)
	}
	return false
}

func (tracker *CircularDependencyTracker) RemoveProject(packageName string) {
	tracker.visitedProjects.Remove(packageName)
}
