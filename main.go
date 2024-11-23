package main

import (
	"fmt"
	"log"
	"os"
	"redun-pendancy/utils"

	"fyne.io/fyne/v2/app"
)

// Gets updated by the pipeline
var APP_VERSION = "DEV"

func main() {
	log.Println("Starting the application...")
	userHomePath, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get home path: %v\n", err)
		return
	}

	filePath, _, err := parseArguments()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse arguments: %v\n", err)
		return
	}

	application := app.New()
	mainWindow := NewMainWindow(application, userHomePath)
	mainWindow.ShowAndRun(filePath)
}

func parseArguments() (string, bool, error) {
	namedArgsConfig := map[string]bool{
		"-gui": true,
	}
	namedArgs, args, err := utils.ParseOSArgs(namedArgsConfig)
	if err != nil {
		return "", false, err
	}
	if len(args) > 1 {
		return "", false, fmt.Errorf("unexpected extra arguments %v", args[1:])
	}

	_, gui := namedArgs["-gui"] //NOTE: gui => exists
	if len(args) == 0 {
		return "", gui, nil
	}
	return args[0], gui, nil
}
