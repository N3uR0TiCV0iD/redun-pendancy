package main

import (
	"errors"
	"fmt"
	"image"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"redun-pendancy/analysis/analyzers"
	"redun-pendancy/gui"
	"redun-pendancy/handlers/dotnet"
	"redun-pendancy/helpers"
	"redun-pendancy/models"
	"redun-pendancy/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	DropFilePrompt = "Drop a project file here"
)

type MainWindow struct {
	ui           MainWindowUI
	window       fyne.Window
	application  fyne.App
	userHomePath string

	projectHandler ProjectHandler

	projectActions      []ProjectAction
	analysisSuggestions string

	lastCursorPos   image.Point
	selectedActions int
}

type MainWindowUI struct {
	dropFileLabel *widget.Label

	workspaceLabel *widget.Label
	searchEntry    *widget.Entry
	dependencyTree *gui.DependencyTree

	dependencyName      *widget.Label
	dependencyVersion   *widget.Label
	dependencyFramework *widget.Label
	totalReferences     *widget.Label

	analyzeButton    *widget.Button
	openReportButton *widget.Button

	analysisResultsContainer *fyne.Container
	analysisSuggestions      *widget.Entry
	actionsScrollContainer   *container.Scroll
	actionsContainer         *fyne.Container
	applyActionsButton       *widget.Button

	statusLabel *widget.Label

	mainContainer *fyne.Container
}

func NewMainWindow(application fyne.App, userHomePath string) *MainWindow {
	window := application.NewWindow(getAppTitle())
	gui.TrySetWindowIcon(window, "assets/icon.png")

	mainWindow := MainWindow{
		window:       window,
		application:  application,
		userHomePath: userHomePath,
	}
	ui := MainWindowUI{
		dropFileLabel: widget.NewLabelWithStyle(DropFilePrompt, fyne.TextAlignCenter, gui.BoldTextStyle()),

		workspaceLabel: widget.NewLabelWithStyle("{workspaceName}", fyne.TextAlignCenter, gui.BoldTextStyle()),
		searchEntry:    widget.NewEntry(),
		dependencyTree: gui.NewDependencyTree(),

		dependencyName:      widget.NewLabel("Name:"),
		dependencyVersion:   widget.NewLabel("Version:"),
		dependencyFramework: widget.NewLabel("Framework:"),
		totalReferences:     widget.NewLabel("Total References:"),

		analyzeButton:    widget.NewButton("Analyze", mainWindow.analyzeButton_Click),
		openReportButton: widget.NewButton("Open report", mainWindow.openReportButton_Click),

		analysisSuggestions: widget.NewMultiLineEntry(),
		actionsContainer:    container.NewVBox(),
		applyActionsButton:  widget.NewButton("Apply Selected Actions", mainWindow.applyActionsButton_Click),

		statusLabel: widget.NewLabel(""),
	}
	ui.analysisResultsContainer = createAnalysisResultsContainer(&ui)
	ui.mainContainer = createMainContainer(&mainWindow, &ui)

	mainWindow.ui = ui
	mainWindow.initUI()
	return &mainWindow
}

func (mainWindow *MainWindow) ShowAndRun(filePath string) {
	if filePath != "" {
		mainWindow.loadFile(filePath)
	}
	mainWindow.window.ShowAndRun()
}

func getAppTitle() string {
	return "redun-pendancy " + APP_VERSION
}

func createMainContainer(mainWindow *MainWindow, ui *MainWindowUI) *fyne.Container {
	collapseButton := widget.NewButtonWithIcon("", theme.ViewRestoreIcon(), mainWindow.collapseButton_Click)
	searchButton := widget.NewButtonWithIcon("Search", theme.SearchIcon(), mainWindow.searchButton_Click)

	// Search Area Container
	searchContainer := container.NewBorder(
		nil,            //top
		nil,            //bottom
		collapseButton, //left
		searchButton,   //right
		ui.searchEntry, //center
	)

	// Scrollable TreeView Container
	treeScrollContainer := container.NewVScroll(ui.dependencyTree.GetWidget())
	treeScrollContainer.SetMinSize(fyne.NewSize(300, 400)) //Ensure it has some size

	// Dependency Info Container
	dependencyInfoContainer := container.NewVBox(
		ui.dependencyName,
		ui.dependencyVersion,
		ui.dependencyFramework,
		ui.totalReferences,
	)

	// Dependency View Split
	dependencyViewSplit := container.NewHSplit(treeScrollContainer, dependencyInfoContainer)
	dependencyViewSplit.Offset = 0.7

	// Main Content Container
	mainContent := container.NewVBox(
		ui.workspaceLabel,
		searchContainer,
		dependencyViewSplit,
		ui.analyzeButton,
		ui.openReportButton,
	)

	mainContainer := container.NewBorder(
		mainContent,                 //top
		ui.statusLabel,              //bottom
		nil,                         //left
		nil,                         //right
		ui.analysisResultsContainer, //center
	)
	return mainContainer
}

func createAnalysisResultsContainer(ui *MainWindowUI) *fyne.Container {
	suggestionsLabel := widget.NewLabelWithStyle("Suggestions:", fyne.TextAlignLeading, gui.BoldTextStyle())

	// Suggestions Panel Container
	suggestionsPanel := container.NewBorder(
		suggestionsLabel,       //top
		nil,                    //bottom
		nil,                    //left
		nil,                    //right
		ui.analysisSuggestions, //center
	)

	ui.actionsScrollContainer = container.NewVScroll(ui.actionsContainer)

	// Actions Panel Container
	actionsPanel := container.NewBorder(
		nil,                       //top
		ui.applyActionsButton,     //bottom
		nil,                       //left
		nil,                       //right
		ui.actionsScrollContainer, //center
	)

	// Analysis Results Container
	analysisResultsContainer := container.NewGridWithColumns(
		2,
		suggestionsPanel,
		actionsPanel,
	)
	return analysisResultsContainer
}

func (mainWindow *MainWindow) initUI() {
	ui := mainWindow.ui
	window := mainWindow.window

	ui.openReportButton.Hide()
	ui.analysisResultsContainer.Hide()

	ui.searchEntry.OnSubmitted = mainWindow.searchEntry_Submit
	ui.analysisSuggestions.OnChanged = mainWindow.analysisSuggestions_Changed
	ui.analysisSuggestions.OnCursorChanged = mainWindow.analysisSuggestions_CursorChanged

	ui.dependencyTree.SetSelectionCallback(mainWindow.dependencyTree_OnSelected)

	tempContent := container.NewCenter(mainWindow.ui.dropFileLabel)
	window.SetContent(tempContent)
	mainWindow.configureShortcuts()
	window.Resize(fyne.NewSize(1100, 800))
	window.SetOnDropped(mainWindow.onFileDropped)
}

func (mainWindow *MainWindow) configureShortcuts() {
	canvas := mainWindow.window.Canvas()

	ctrlF := &desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierControl}
	canvas.AddShortcut(ctrlF, mainWindow.focusSearchEntry)
}

func (mainWindow *MainWindow) onFileDropped(position fyne.Position, uriArray []fyne.URI) {
	if len(uriArray) != 1 {
		mainWindow.showErrorMessage("Please drop exactly one file.")
		return
	}

	fileURI := uriArray[0]
	filePath := fileURI.Path()
	mainWindow.loadFile(filePath)
}

func (mainWindow *MainWindow) loadFile(filePath string) {
	projectHandler := mainWindow.getProjectHandler(filePath)
	if projectHandler == nil {
		message := "No suitable project handler found for file: " + filePath
		mainWindow.showErrorMessage(message)
		return
	}

	log.Println("Loading:", filePath)
	mainWindow.setDropFileLabelText("Loading project...")
	err := projectHandler.Initialize(filePath)
	if err != nil {
		mainWindow.setDropFileLabelText("Drop a project file here")
		mainWindow.showError(err)
		return
	}

	log.Println("Project loaded successfully!")
	mainWindow.projectHandler = projectHandler
	mainWindow.refreshProjectView()
}

func (mainWindow *MainWindow) getProjectHandler(filePath string) ProjectHandler {
	fileName := filepath.Base(filePath)
	fileName = strings.ToLower(fileName)
	if strings.HasSuffix(fileName, ".sln") {
		return dotnet.NewDotNetProjectHandler(mainWindow.userHomePath)
	}
	if fileName == "package.json" {
		return nil //nodejs.NewNodeJSProjectHandler()
	}
	if fileName == "pom.xml" {
		return nil //maven.NewMavenProjectHandler()
	}
	return nil
}

func (mainWindow *MainWindow) setDropFileLabelText(text string) {
	mainWindow.ui.dropFileLabel.SetText(text)
}

func (mainWindow *MainWindow) refreshProjectView() {
	ui := mainWindow.ui
	window := mainWindow.window
	projectHandler := mainWindow.projectHandler

	workspaceName := projectHandler.GetWorkspaceName()
	ui.workspaceLabel.SetText(workspaceName)

	title := getAppTitle() + " - " + workspaceName
	window.SetTitle(title)

	ui.dependencyTree.Load(projectHandler)
	window.SetContent(ui.mainContainer)
	ui.mainContainer.Refresh()
}

func (mainWindow *MainWindow) collapseButton_Click() {
	mainWindow.ui.dependencyTree.CollapseAll()
}

func (mainWindow *MainWindow) focusSearchEntry(shortcut fyne.Shortcut) {
	gui.FocusWidget(mainWindow.window, mainWindow.ui.searchEntry)
}

func (mainWindow *MainWindow) searchButton_Click() {
	searchTerm := mainWindow.ui.searchEntry.Text
	mainWindow.searchEntry_Submit(searchTerm)
}

func (mainWindow *MainWindow) searchEntry_Submit(searchTerm string) {
	dependencyTree := mainWindow.ui.dependencyTree
	if searchTerm == "" {
		dependencyTree.SetFilter(nil)
		return
	}

	matchingPaths := dependencyTree.SearchNodes(searchTerm)
	if len(matchingPaths) == 0 {
		mainWindow.showErrorMessage(fmt.Sprintf(`No results found for "%s"`, searchTerm))
		return
	}

	dependencyTree.CollapseAll()
	dependencyTree.SetFilter(matchingPaths)
	nodePaths := dependencyTree.GetFilter()
	for _, nodePath := range nodePaths {
		dependencyTree.ExpandNode(nodePath)
	}
	dependencyTree.SelectPath(matchingPaths[0])
}

func (mainWindow *MainWindow) dependencyTree_OnSelected(packageInfo *models.PackageInfo) {
	ui := mainWindow.ui
	version := utils.ValueOrDefault(packageInfo.Version, "1.0")

	ui.dependencyName.SetText("Name: " + packageInfo.Name)
	ui.dependencyVersion.SetText("Version: " + version)
	ui.dependencyFramework.SetText("Framework: " + packageInfo.Framework)
	ui.totalReferences.SetText("Total References: " + fmt.Sprint(len(packageInfo.Parents)))
}

func (mainWindow *MainWindow) analyzeButton_Click() {
	ui := mainWindow.ui
	results := analyzers.AnalyzeProject(mainWindow.projectHandler)
	actions := results.GetActions()
	mainWindow.projectActions = actions

	formattedSuggestions := formatSuggestions(results.GetSuggestions())
	mainWindow.analysisSuggestions = formattedSuggestions
	ui.analysisSuggestions.SetText(formattedSuggestions)

	actionsContainer := ui.actionsContainer
	selectedActions := 0
	actionsContainer.RemoveAll()
	for _, action := range actions {
		checkbox := gui.NewHoverableCheckbox(action.GetDescription(), mainWindow.actionCheckbox_CheckChanged)
		if action.IsRecommended() {
			checkbox.SetChecked(true)
			selectedActions++
		}

		//Capture the current action in a new variable to avoid closure issues
		currAction := action
		checkbox.SetMouseInCallback(func() {
			mainWindow.actionCheckbox_MouseIn(currAction)
		})
		checkbox.SetMouseOutCallback(mainWindow.actionCheckbox_MouseOut)
		actionsContainer.Add(checkbox)
	}

	ui.analyzeButton.Hide()
	ui.openReportButton.Show()
	ui.analysisResultsContainer.Show()
	ui.actionsScrollContainer.ScrollToTop()
	mainWindow.selectedActions = selectedActions
	mainWindow.setApplyActionsEnabled(selectedActions != 0)
}

func formatSuggestions(suggestions map[string][]string) string {
	result := strings.Builder{}
	sortedProjects := utils.GetMapKeys(suggestions)
	sort.Strings(sortedProjects)
	for _, projectName := range sortedProjects {
		suggestionList := suggestions[projectName]
		result.WriteString(projectName)
		result.WriteString(":\n")
		for _, suggestion := range suggestionList {
			result.WriteString("* ")
			result.WriteString(suggestion)
			result.WriteString("\n")
		}
	}
	return result.String()
}

func (mainWindow *MainWindow) actionCheckbox_MouseIn(action ProjectAction) {
	statusText := "REASON: " + action.GetReason()
	statusLabel := mainWindow.ui.statusLabel
	statusLabel.SetText(statusText)
}

func (mainWindow *MainWindow) actionCheckbox_CheckChanged(checked bool) {
	if checked {
		mainWindow.selectedActions++
	} else {
		mainWindow.selectedActions--
	}
	mainWindow.setApplyActionsEnabled(mainWindow.selectedActions != 0)
}

func (mainWindow *MainWindow) setApplyActionsEnabled(enabled bool) {
	applyActionsButton := mainWindow.ui.applyActionsButton
	if enabled {
		applyActionsButton.Enable()
		return
	}
	applyActionsButton.Disable()
}

func (mainWindow *MainWindow) actionCheckbox_MouseOut() {
	statusLabel := mainWindow.ui.statusLabel
	statusLabel.SetText("")
}

func (mainWindow *MainWindow) openReportButton_Click() {
	dependenciesCount := helpers.NewHashBag[string]()
	for _, project := range mainWindow.projectHandler.GetProjects() {
		for _, dependency := range project.Dependencies {
			dependenciesCount.Add(dependency.Name)
		}
	}
	dependencyList := createDependencyList(dependenciesCount)
	scrollableContent := container.NewScroll(dependencyList)
	customDialog := dialog.NewCustom("Dependencies Report", "Close", scrollableContent, mainWindow.window)
	customDialog.Resize(fyne.NewSize(600, 400))
	customDialog.Show()
}

func createDependencyList(dependenciesCount *helpers.HashBag[string]) fyne.CanvasObject {
	dependencyList := container.NewVBox()
	sortedDependencies := dependenciesCount.GetSortedEntriesDesc()
	for _, entry := range sortedDependencies {
		label := fmt.Sprintf("• %s - %d reference(s)", entry.Key, entry.Value)
		dependencyList.Add(widget.NewLabel(label))
	}
	return dependencyList
}

func (mainWindow *MainWindow) analysisSuggestions_CursorChanged() {
	suggestionsBox := mainWindow.ui.analysisSuggestions
	mainWindow.lastCursorPos = image.Point{
		X: suggestionsBox.CursorColumn,
		Y: suggestionsBox.CursorRow,
	}
}

func (mainWindow *MainWindow) analysisSuggestions_Changed(newText string) {
	//Forces the text to always be set to what the analysis reported :) | Effectively "disabling" the textbox
	suggestions := mainWindow.analysisSuggestions
	if newText != suggestions {
		lastCursorPos := mainWindow.lastCursorPos
		suggestionsBox := mainWindow.ui.analysisSuggestions
		suggestionsBox.CursorColumn = lastCursorPos.X
		suggestionsBox.CursorRow = lastCursorPos.Y
		suggestionsBox.SetText(suggestions)
	}
}

func (mainWindow *MainWindow) applyActionsButton_Click() {
	err := mainWindow.applySelectedActions()
	if err != nil {
		mainWindow.projectHandler.RevertChanges()
		mainWindow.showError(err)
		return
	}

	log.Println("All actions applied!")
	dialog.ShowInformation("Success", "Actions applied successfully!", mainWindow.window)
	mainWindow.analyzeButton_Click()
}

func (mainWindow *MainWindow) applySelectedActions() error {
	actionsContainer := mainWindow.ui.actionsContainer
	for index, obj := range actionsContainer.Objects {
		checkbox, isCheckbox := obj.(*gui.HoverableCheckbox)
		if !isCheckbox || !checkbox.Checked {
			continue
		}

		err := mainWindow.executeAction(index)
		if err != nil {
			return err
		}
	}

	err := mainWindow.projectHandler.CommitChanges()
	if err != nil {
		return err
	}
	return nil
}

func (mainWindow *MainWindow) executeAction(index int) error {
	action := mainWindow.projectActions[index]
	err := action.Execute(mainWindow.projectHandler)
	return err
}

func (mainWindow *MainWindow) showErrorMessage(message string) {
	mainWindow.showError(errors.New(message))
}

func (mainWindow *MainWindow) showError(err error) {
	dialog.ShowError(err, mainWindow.window)
}
