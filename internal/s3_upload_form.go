package internal

import (
	"fmt"
	"mime"
	"path/filepath"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/view"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type S3UploadForm struct {
	*tview.Flex
	view.S3
	repo       *repo.S3
	bucket     string
	prefix     string
	filePath   string
	app        *Application
	form       *tview.Form
	onComplete func()
}

func NewS3UploadForm(repo *repo.S3, bucket, prefix, filePath string, app *Application, onComplete func()) *S3UploadForm {
	filename := filepath.Base(filePath)
	key := prefix + filename

	// Use standard library mime.TypeByExtension
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(" Upload File ")
	form.SetTitleColor(tcell.ColorGreen)

	u := &S3UploadForm{
		Flex:       tview.NewFlex(),
		repo:       repo,
		bucket:     bucket,
		prefix:     prefix,
		filePath:   filePath,
		app:        app,
		form:       form,
		onComplete: onComplete,
	}

	// Add form fields
	form.AddInputField("Local File", filePath, 0, nil, nil).
		AddInputField("S3 Key", key, 0, nil, nil).
		AddInputField("Content-Type", contentType, 0, nil, nil).
		AddDropDown("ACL", []string{
			"private",
			"public-read",
			"public-read-write",
			"authenticated-read",
			"aws-exec-read",
			"bucket-owner-read",
			"bucket-owner-full-control",
		}, 0, nil)

	form.AddButton("Upload", u.uploadHandler)
	form.AddButton("Cancel", u.cancelHandler)

	// Make read-only field (file path) non-editable
	form.GetFormItem(0).(*tview.InputField).SetDisabled(true)

	// Set focus styling
	form.SetFieldBackgroundColor(tcell.ColorBlack)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorYellow)
	form.SetButtonBackgroundColor(tcell.ColorGreen)
	form.SetButtonTextColor(tcell.ColorBlack)

	u.Flex.SetDirection(tview.FlexRow)
	u.Flex.AddItem(form, 0, 1, true)

	return u
}

func (u *S3UploadForm) uploadHandler() {
	key := u.form.GetFormItem(1).(*tview.InputField).GetText()
	contentType := u.form.GetFormItem(2).(*tview.InputField).GetText()
	_, aclText := u.form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()

	var acl s3Types.ObjectCannedACL
	switch aclText {
	case "private":
		acl = s3Types.ObjectCannedACLPrivate
	case "public-read":
		acl = s3Types.ObjectCannedACLPublicRead
	case "public-read-write":
		acl = s3Types.ObjectCannedACLPublicReadWrite
	case "authenticated-read":
		acl = s3Types.ObjectCannedACLAuthenticatedRead
	case "aws-exec-read":
		acl = s3Types.ObjectCannedACLAwsExecRead
	case "bucket-owner-read":
		acl = s3Types.ObjectCannedACLBucketOwnerRead
	case "bucket-owner-full-control":
		acl = s3Types.ObjectCannedACLBucketOwnerFullControl
	}

	err := u.repo.UploadObject(u.bucket, key, u.filePath, contentType, acl)
	if err != nil {
		u.showError(fmt.Sprintf("Upload failed: %v", err))
		return
	}

	u.app.Close()
	if u.onComplete != nil {
		u.onComplete()
	}
}

func (u *S3UploadForm) cancelHandler() {
	u.app.Close()
}

func (u *S3UploadForm) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			u.app.Close()
		})
	u.app.AddAndSwitch(&ComponentWrapper{Primitive: modal, service: "S3", labels: []string{"Error"}})
}

func (u S3UploadForm) GetService() string {
	return "S3"
}

func (u S3UploadForm) GetLabels() []string {
	return []string{u.bucket, "Upload"}
}

func (u S3UploadForm) GetKeyActions() []KeyAction {
	return []KeyAction{}
}

func (u S3UploadForm) Render() {
}

// ComponentWrapper wraps a tview.Primitive to implement Component interface
type ComponentWrapper struct {
	tview.Primitive
	service string
	labels  []string
}

func (c ComponentWrapper) GetService() string {
	return c.service
}

func (c ComponentWrapper) GetLabels() []string {
	return c.labels
}

func (c ComponentWrapper) GetKeyActions() []KeyAction {
	return []KeyAction{}
}

func (c ComponentWrapper) Render() {
}
