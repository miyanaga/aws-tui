package internal

import (
	"strings"

	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/settings"
	"github.com/bporter816/aws-tui/internal/ui"
	"github.com/bporter816/aws-tui/internal/view"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type S3Objects struct {
	*ui.Tree
	view.S3
	repo     *repo.S3
	bucket   string
	app      *Application
	settings *settings.Settings
}

func NewS3Objects(repo *repo.S3, bucket string, app *Application) *S3Objects {
	root := tview.NewTreeNode(bucket + "/")
	root.SetReference("")

	// Load settings
	userSettings, err := settings.Load()
	if err != nil {
		userSettings = &settings.Settings{}
	}

	s := &S3Objects{
		Tree:     ui.NewTree(root),
		repo:     repo,
		bucket:   bucket,
		app:      app,
		settings: userSettings,
	}
	s.SetSelectedFunc(s.selectHandler)
	return s
}

func (s S3Objects) GetLabels() []string {
	return []string{s.bucket, "Objects"}
}

func (s S3Objects) selectHandler(n *tview.TreeNode) {
	s.expandDir(n)
}

func (s S3Objects) objectHandler() {
	if node := s.GetCurrentNode(); node != nil {
		key := node.GetReference().(string)
		if strings.HasSuffix(key, "/") {
			return
		}
		objectView := NewS3Object(s.repo, s.bucket, key, s.app)
		s.app.AddAndSwitch(objectView)
	}
}

func (s S3Objects) metadataHandler() {
	if node := s.GetCurrentNode(); node != nil {
		key := node.GetReference().(string)
		if strings.HasSuffix(key, "/") {
			return
		}
		metadataView := NewS3ObjectMetadata(s.repo, s.bucket, key, s.app)
		s.app.AddAndSwitch(metadataView)
	}
}

func (s S3Objects) tagsHandler() {
	if node := s.GetCurrentNode(); node != nil {
		key := node.GetReference().(string)
		if strings.HasSuffix(key, "/") {
			return
		}
		tagsView := NewTags(s.repo, s.GetService(), "object:"+s.bucket+":"+key, s.app)
		s.app.AddAndSwitch(tagsView)
	}
}

func (s *S3Objects) uploadHandler() {
	// Get current prefix (directory)
	prefix := ""
	if node := s.GetCurrentNode(); node != nil {
		ref := node.GetReference().(string)
		if strings.HasSuffix(ref, "/") {
			prefix = ref
		} else {
			// Use parent directory
			lastSlash := strings.LastIndex(ref, "/")
			if lastSlash >= 0 {
				prefix = ref[:lastSlash+1]
			}
		}
	}

	// Show file selector from current local directory
	localDir := s.settings.GetLocalDirectory()
	fileSelector := ui.NewFileSelector(localDir, func(filePath string) {
		// File selected, show upload form
		uploadForm := NewS3UploadForm(s.repo, s.bucket, prefix, filePath, s.app, func() {
			s.Render()
		})
		s.app.AddAndSwitch(uploadForm)
	})

	s.app.AddAndSwitch(&ComponentWrapper{
		Primitive: fileSelector,
		service:   "S3",
		labels:    []string{s.bucket, "Select File"},
	})
}

func (s *S3Objects) downloadHandler() {
	if node := s.GetCurrentNode(); node != nil {
		key := node.GetReference().(string)
		if strings.HasSuffix(key, "/") {
			return
		}
		downloadForm := NewS3DownloadForm(s.repo, s.bucket, key, s.settings, s.app, func() {
			// No need to refresh on download
		})
		s.app.AddAndSwitch(downloadForm)
	}
}

func (s *S3Objects) changeDirectoryHandler() {
	changeDirForm := NewChangeDirectoryForm(s.settings, s.app, func() {
		// Directory changed, settings updated
	})
	s.app.AddAndSwitch(changeDirForm)
}

func (s S3Objects) GetKeyActions() []KeyAction {
	return []KeyAction{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),
			Description: "View Object",
			Action:      s.objectHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone),
			Description: "Metadata",
			Action:      s.metadataHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'T', tcell.ModNone),
			Description: "Tags",
			Action:      s.tagsHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),
			Description: "Upload",
			Action:      s.uploadHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
			Description: "Download",
			Action:      s.downloadHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
			Description: "Change Local Dir",
			Action:      s.changeDirectoryHandler,
		},
	}
}

func (s S3Objects) expandDir(n *tview.TreeNode) {
	if strings.HasSuffix(n.GetText(), "/") {
		if len(n.GetChildren()) > 0 {
			n.SetExpanded(!n.IsExpanded())
			return
		}

		ref := n.GetReference().(string)
		prefixes, objects, err := s.repo.ListObjects(s.bucket, ref)
		if err != nil {
			panic(err)
		}
		for _, prefix := range prefixes {
			arr := strings.Split(prefix, "/")
			label := arr[len(arr)-2] + "/"
			c := tview.NewTreeNode(label)
			c.SetColor(tcell.ColorGreen)
			c.SetReference(ref + label)
			n.AddChild(c)
		}
		for _, object := range objects {
			if strings.HasSuffix(object, "/") {
				continue
			}
			label := object[strings.LastIndex(object, "/")+1:]
			c := tview.NewTreeNode(label)
			c.SetReference(ref + label)
			n.AddChild(c)
		}
	}
}

func (s S3Objects) Render() {
	s.expandDir(s.GetRoot())
}
