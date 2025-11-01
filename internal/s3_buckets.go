package internal

import (
	"strings"

	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/ui"
	"github.com/bporter816/aws-tui/internal/utils"
	"github.com/bporter816/aws-tui/internal/view"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type S3Buckets struct {
	*tview.Flex
	view.S3
	repo          *repo.S3
	app           *Application
	table         *ui.Table
	filterLabel   *tview.TextView
	searchField   *tview.InputField
	searchVisible bool
	allData       [][]string
	filteredData  [][]string
	lastFilter    string
	isFiltered    bool
}

func NewS3Buckets(repo *repo.S3, app *Application) *S3Buckets {
	table := ui.NewTable([]string{
		"NAME",
		"CREATED",
	}, 1, 0)

	filterLabel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	searchField := tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(50)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)

	s := &S3Buckets{
		Flex:          flex,
		repo:          repo,
		app:           app,
		table:         table,
		filterLabel:   filterLabel,
		searchField:   searchField,
		searchVisible: false,
		allData:       make([][]string, 0),
		filteredData:  make([][]string, 0),
		lastFilter:    "",
		isFiltered:    false,
	}

	s.table.SetSelectedFunc(s.selectHandler)
	s.setupSearch()
	return s
}

func (s S3Buckets) GetLabels() []string {
	return []string{"Buckets"}
}

func (s *S3Buckets) selectHandler(row, col int) {
	bucket, err := s.table.GetColSelection("NAME")
	if err != nil {
		return
	}
	objectsView := NewS3Objects(s.repo, bucket, s.app)
	s.app.AddAndSwitch(objectsView)
}

func (s *S3Buckets) bucketPolicyHandler() {
	bucket, err := s.table.GetColSelection("NAME")
	if err != nil {
		return
	}
	policyView := NewS3BucketPolicy(s.repo, bucket, s.app)
	s.app.AddAndSwitch(policyView)
}

func (s *S3Buckets) corsRulesHandler() {
	bucket, err := s.table.GetColSelection("NAME")
	if err != nil {
		return
	}
	corsRulesView := NewS3CORSRules(s.repo, bucket, s.app)
	s.app.AddAndSwitch(corsRulesView)
}

func (s *S3Buckets) tagsHandler() {
	bucket, err := s.table.GetColSelection("NAME")
	if err != nil {
		return
	}
	tagsView := NewTags(s.repo, s.GetService(), "bucket:"+bucket, s.app)
	s.app.AddAndSwitch(tagsView)
}

func (s *S3Buckets) setupSearch() {
	s.searchField.SetChangedFunc(func(text string) {
		s.filterDataLive(text)
	})

	s.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			s.commitFilter()
			return nil
		}
		return event
	})

	s.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			s.showSearch()
			return nil
		}
		if event.Key() == tcell.KeyHome || (event.Key() == tcell.KeyRune && event.Rune() == 'g') {
			s.table.Select(1, 0)
			return nil
		}
		return event
	})
}

func (s *S3Buckets) showSearch() {
	if !s.searchVisible {
		s.searchVisible = true
		s.searchField.SetText(s.lastFilter)
		s.Flex.Clear()
		s.Flex.AddItem(s.searchField, 1, 0, true)
		s.Flex.AddItem(s.table, 0, 1, false)
		s.app.app.SetFocus(s.searchField)
	}
}

func (s *S3Buckets) updateFilterDisplay() {
	s.Flex.Clear()
	if s.isFiltered && s.lastFilter != "" {
		s.filterLabel.SetText("Filter: " + s.lastFilter)
		s.Flex.AddItem(s.filterLabel, 1, 0, false)
	}
	s.Flex.AddItem(s.table, 0, 1, true)
}

func (s *S3Buckets) commitFilter() {
	s.searchVisible = false
	text := s.searchField.GetText()
	s.lastFilter = text
	s.isFiltered = (text != "")

	s.updateFilterDisplay()
	s.app.app.SetFocus(s.table)

	if s.isFiltered {
		s.setTableData(s.filteredData)
	} else {
		s.setTableData(s.allData)
	}
}

func (s *S3Buckets) filterDataLive(searchText string) {
	if searchText == "" {
		s.filteredData = s.allData
		s.setTableData(s.allData)
		return
	}

	searchLower := strings.ToLower(searchText)
	var filtered [][]string

	for _, row := range s.allData {
		for _, col := range row {
			if strings.Contains(strings.ToLower(col), searchLower) {
				filtered = append(filtered, row)
				break
			}
		}
	}

	s.filteredData = filtered
	s.setTableData(filtered)
}

func (s *S3Buckets) setTableData(data [][]string) {
	rowCount := s.table.GetRowCount()
	for i := rowCount - 1; i > 0; i-- {
		s.table.RemoveRow(i)
	}

	s.table.SetData(data)

	if len(data) > 0 {
		s.table.Select(1, 0)
	}
}

func (s *S3Buckets) GetKeyActions() []KeyAction {
	return []KeyAction{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			Description: "Filter",
			Action:      s.showSearch,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
			Description: "Bucket Policy",
			Action:      s.bucketPolicyHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
			Description: "CORS Rules",
			Action:      s.corsRulesHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'T', tcell.ModNone),
			Description: "Tags",
			Action:      s.tagsHandler,
		},
	}
}

func (s *S3Buckets) Render() {
	model, err := s.repo.ListBuckets()
	if err != nil {
		panic(err)
	}

	var data [][]string
	for _, v := range model {
		var created string
		if v.CreationDate != nil {
			created = v.CreationDate.Format(utils.DefaultTimeFormat)
		}
		data = append(data, []string{
			utils.DerefString(v.Name, ""),
			created,
		})
	}
	s.allData = data

	if s.isFiltered && s.lastFilter != "" {
		s.filterDataLive(s.lastFilter)
	} else {
		s.setTableData(data)
	}
}
