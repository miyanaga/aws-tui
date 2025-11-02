package internal

import (
	"fmt"
	"path/filepath"

	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/settings"
	"github.com/bporter816/aws-tui/internal/view"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type S3DownloadForm struct {
	*tview.Form
	view.S3
	repo       *repo.S3
	bucket     string
	key        string
	settings   *settings.Settings
	app        *Application
	onComplete func()
}

func NewS3DownloadForm(repo *repo.S3, bucket, key string, settings *settings.Settings, app *Application, onComplete func()) *S3DownloadForm {
	defaultFilename := filepath.Base(key)
	localDir := settings.GetLocalDirectory()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(" Download File ")
	form.SetTitleColor(tcell.ColorBlue)

	d := &S3DownloadForm{
		Form:       form,
		repo:       repo,
		bucket:     bucket,
		key:        key,
		settings:   settings,
		app:        app,
		onComplete: onComplete,
	}

	form.AddInputField("S3 Key", key, 0, nil, nil).
		AddInputField("Local Directory", localDir, 0, nil, nil).
		AddInputField("Filename", defaultFilename, 0, nil, nil)

	form.AddButton("Download", d.downloadHandler)
	form.AddButton("Cancel", d.cancelHandler)

	// Make read-only fields non-editable
	form.GetFormItem(0).(*tview.InputField).SetDisabled(true)
	form.GetFormItem(1).(*tview.InputField).SetDisabled(true)

	// Set focus styling
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorYellow)
	form.SetButtonBackgroundColor(tcell.ColorBlue)
	form.SetButtonTextColor(tcell.ColorWhite)

	return d
}

func (d *S3DownloadForm) downloadHandler() {
	dirPath := d.GetFormItem(1).(*tview.InputField).GetText()
	filename := d.GetFormItem(2).(*tview.InputField).GetText()
	destPath := filepath.Join(dirPath, filename)

	err := d.repo.DownloadObject(d.bucket, d.key, destPath)
	if err != nil {
		d.showError(fmt.Sprintf("Download failed: %v", err))
		return
	}

	d.app.Close()
	if d.onComplete != nil {
		d.onComplete()
	}
}

func (d *S3DownloadForm) cancelHandler() {
	d.app.Close()
}

func (d *S3DownloadForm) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			d.app.Close()
		})
	d.app.AddAndSwitch(&ComponentWrapper{Primitive: modal, service: "S3", labels: []string{"Error"}})
}

func (d S3DownloadForm) GetService() string {
	return "S3"
}

func (d S3DownloadForm) GetLabels() []string {
	return []string{d.bucket, "Download"}
}

func (d S3DownloadForm) GetKeyActions() []KeyAction {
	return []KeyAction{}
}

func (d S3DownloadForm) Render() {
}
