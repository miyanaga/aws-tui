package internal

import (
	"fmt"
	"os"

	"github.com/bporter816/aws-tui/internal/settings"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChangeDirectoryForm struct {
	*tview.Form
	settings   *settings.Settings
	app        *Application
	onComplete func()
}

func NewChangeDirectoryForm(settings *settings.Settings, app *Application, onComplete func()) *ChangeDirectoryForm {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(" Change Local Directory ")
	form.SetTitleColor(tcell.ColorYellow)

	cd := &ChangeDirectoryForm{
		Form:       form,
		settings:   settings,
		app:        app,
		onComplete: onComplete,
	}

	currentDir := settings.GetLocalDirectory()
	form.AddInputField("Directory Path", currentDir, 0, nil, nil)
	form.AddButton("Change", cd.changeHandler)
	form.AddButton("Cancel", cd.cancelHandler)

	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorYellow)
	form.SetButtonBackgroundColor(tcell.ColorYellow)
	form.SetButtonTextColor(tcell.ColorBlack)

	return cd
}

func (cd *ChangeDirectoryForm) changeHandler() {
	dirPath := cd.GetFormItem(0).(*tview.InputField).GetText()

	// Expand ~ to home directory
	if dirPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			dirPath = home + dirPath[1:]
		}
	}

	// Validate directory
	info, err := os.Stat(dirPath)
	if err != nil {
		cd.showError(fmt.Sprintf("Directory not found: %v", err))
		return
	}
	if !info.IsDir() {
		cd.showError("Path is not a directory")
		return
	}

	// Save to settings
	if err := cd.settings.SetLocalDirectory(dirPath); err != nil {
		cd.showError(fmt.Sprintf("Failed to save: %v", err))
		return
	}

	cd.app.Close()
	if cd.onComplete != nil {
		cd.onComplete()
	}
}

func (cd *ChangeDirectoryForm) cancelHandler() {
	cd.app.Close()
}

func (cd *ChangeDirectoryForm) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			cd.app.Close()
		})
	cd.app.AddAndSwitch(&ComponentWrapper{Primitive: modal, service: "Settings", labels: []string{"Error"}})
}

func (cd ChangeDirectoryForm) GetService() string {
	return "Settings"
}

func (cd ChangeDirectoryForm) GetLabels() []string {
	return []string{"Change Directory"}
}

func (cd ChangeDirectoryForm) GetKeyActions() []KeyAction {
	return []KeyAction{}
}

func (cd ChangeDirectoryForm) Render() {
}
