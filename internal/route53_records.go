package internal

import (
	"strconv"
	"strings"

	r53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/bporter816/aws-tui/internal/model"
	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/ui"
	"github.com/bporter816/aws-tui/internal/utils"
	"github.com/bporter816/aws-tui/internal/view"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Route53Records struct {
	*tview.Flex
	view.Route53
	repo           *repo.Route53
	hostedZoneId   string
	hostedZoneName string
	app            *Application
	cachedRecords  []model.Route53Record
	table          *ui.Table
	filterLabel    *tview.TextView
	searchField    *tview.InputField
	searchVisible  bool
	allData        [][]string
	filteredData   [][]string
	lastFilter     string
	isFiltered     bool
}

func NewRoute53Records(repo *repo.Route53, zoneId, zoneName string, app *Application) *Route53Records {
	table := ui.NewTable([]string{
		"RECORD NAME",
		"TYPE",
		"ROUTING",
		"DIFF",
		"LABEL",
		"TTL",
		"ALIAS",
		"VALUE",
	}, 1, 1)

	filterLabel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	searchField := tview.NewInputField().
		SetLabel("Filter: ").
		SetFieldWidth(50)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow)
	flex.AddItem(table, 0, 1, true)

	r := &Route53Records{
		Flex:           flex,
		repo:           repo,
		hostedZoneId:   zoneId,
		hostedZoneName: zoneName,
		app:            app,
		table:          table,
		filterLabel:    filterLabel,
		searchField:    searchField,
		searchVisible:  false,
		allData:        make([][]string, 0),
		filteredData:   make([][]string, 0),
		lastFilter:     "",
		isFiltered:     false,
	}

	r.setupSearch()
	return r
}

func (r *Route53Records) GetLabels() []string {
	return []string{r.hostedZoneId, "Records"}
}

func (r *Route53Records) createRecordHandler() {
	form := NewRoute53RecordForm(r.repo, r.hostedZoneId, r.hostedZoneName, "create", nil, r.app, func() {
		r.Render()
	})
	r.app.AddAndSwitch(form)
}

func (r *Route53Records) updateRecordHandler() {
	row, err := r.table.GetRowSelection()
	if err != nil {
		return
	}

	if row-1 >= len(r.cachedRecords) {
		return
	}

	record := r.cachedRecords[row-1]
	recordSet := r53Types.ResourceRecordSet(record)
	form := NewRoute53RecordForm(r.repo, r.hostedZoneId, r.hostedZoneName, "update", &recordSet, r.app, func() {
		r.Render()
	})
	r.app.AddAndSwitch(form)
}

func (r *Route53Records) deleteRecordHandler() {
	row, err := r.table.GetRowSelection()
	if err != nil {
		return
	}

	if row-1 >= len(r.cachedRecords) {
		return
	}

	record := r.cachedRecords[row-1]
	recordSet := r53Types.ResourceRecordSet(record)
	form := NewRoute53RecordForm(r.repo, r.hostedZoneId, r.hostedZoneName, "delete", &recordSet, r.app, func() {
		r.Render()
	})
	r.app.AddAndSwitch(form)
}

func (r *Route53Records) setupSearch() {
	r.searchField.SetChangedFunc(func(text string) {
		r.filterDataLive(text)
	})

	r.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Tab: Switch to table
			r.commitFilter()
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			// Enter: Commit filter and keep filtered state
			r.commitFilter()
			return nil
		}
		return event
	})

	r.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			r.showSearch()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			// Tab: Switch to filter field
			r.showSearch()
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			r.updateRecordHandler()
			return nil
		}
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 || event.Key() == tcell.KeyDelete {
			r.deleteRecordHandler()
			return nil
		}
		if event.Key() == tcell.KeyHome || (event.Key() == tcell.KeyRune && event.Rune() == 'g') {
			r.table.Select(1, 0)
			return nil
		}
		return event
	})
}

func (r *Route53Records) showSearch() {
	if !r.searchVisible {
		r.searchVisible = true
		// Set previous filter text if any
		r.searchField.SetText(r.lastFilter)
		r.Flex.Clear()
		r.Flex.AddItem(r.searchField, 1, 0, true)
		r.Flex.AddItem(r.table, 0, 1, false)
		r.app.app.SetFocus(r.searchField)
	}
}

func (r *Route53Records) updateFilterDisplay() {
	r.Flex.Clear()
	if r.isFiltered && r.lastFilter != "" {
		r.filterLabel.SetText("Filter: " + r.lastFilter)
		r.Flex.AddItem(r.filterLabel, 1, 0, false)
	}
	r.Flex.AddItem(r.table, 0, 1, true)
}

func (r *Route53Records) commitFilter() {
	// Enter: Keep filter, hide search field, return focus to table
	r.searchVisible = false
	text := r.searchField.GetText()
	r.lastFilter = text
	r.isFiltered = (text != "")

	r.updateFilterDisplay()
	r.app.app.SetFocus(r.table)

	// Keep the filtered data
	if r.isFiltered {
		r.setTableData(r.filteredData)
	} else {
		r.setTableData(r.allData)
	}
}

func (r *Route53Records) filterDataLive(searchText string) {
	// Real-time filtering while typing
	if searchText == "" {
		r.filteredData = r.allData
		r.setTableData(r.allData)
		return
	}

	searchLower := strings.ToLower(searchText)
	var filtered [][]string

	for _, row := range r.allData {
		for _, col := range row {
			if strings.Contains(strings.ToLower(col), searchLower) {
				filtered = append(filtered, row)
				break
			}
		}
	}

	r.filteredData = filtered
	r.setTableData(filtered)
}

func (r *Route53Records) setTableData(data [][]string) {
	// Clear existing rows (except header)
	rowCount := r.table.GetRowCount()
	for i := rowCount - 1; i > 0; i-- {
		r.table.RemoveRow(i)
	}

	// Add new data
	r.table.SetData(data)

	// Select first row if exists
	if len(data) > 0 {
		r.table.Select(1, 0)
	}
}

func (r *Route53Records) GetKeyActions() []KeyAction {
	return []KeyAction{
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone),
			Description: "Filter",
			Action:      r.showSearch,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
			Description: "Create",
			Action:      r.createRecordHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone),
			Description: "Update",
			Action:      r.updateRecordHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone),
			Description: "Delete",
			Action:      r.deleteRecordHandler,
		},
		{
			Key:         tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone),
			Description: "Delete",
			Action:      r.deleteRecordHandler,
		},
	}
}

func (r *Route53Records) Render() {
	model, err := r.repo.ListRecords(r.hostedZoneId)
	if err != nil {
		panic(err)
	}

	// Cache the records for use in handlers
	r.cachedRecords = model

	var data [][]string
	for _, v := range model {
		routingPolicy := "Simple"
		differentiator := "-"
		label := "-"

		// TODO IP routing policy and health checks

		if string(v.Failover) != "" {
			routingPolicy = "Failover"
			differentiator = string(v.Failover)
		}

		if string(v.Region) != "" {
			routingPolicy = "Latency"
			differentiator = string(v.Region)
		}

		if v.GeoLocation != nil {
			routingPolicy = "Geolocation"
			// TODO verify logic here
			if v.GeoLocation.ContinentCode != nil {
				differentiator = *v.GeoLocation.ContinentCode
			} else if v.GeoLocation.CountryCode != nil {
				differentiator = *v.GeoLocation.CountryCode
			} else if v.GeoLocation.SubdivisionCode != nil {
				// TODO include country code
				differentiator = *v.GeoLocation.SubdivisionCode
			}
		}

		if v.MultiValueAnswer != nil && *v.MultiValueAnswer {
			routingPolicy = "MultiValue"
			// TODO is there anything for the differentiator?
		}

		if v.Weight != nil {
			routingPolicy = "Weighted"
			differentiator = strconv.FormatInt(*v.Weight, 10)
		}

		if routingPolicy != "Simple" && v.SetIdentifier != nil {
			label = *v.SetIdentifier
		}

		if v.AliasTarget == nil {
			// not an alias
			data = append(data, []string{
				strings.TrimSuffix(*v.Name, "."),
				string(v.Type),
				routingPolicy,
				differentiator,
				label,
				strconv.FormatInt(*v.TTL, 10),
				"No",
				// TODO consider removing, also see utils/route53.go
				// utils.JoinRoute53ResourceRecords(v.ResourceRecords, ","),
				utils.FormatRoute53ResourceRecords(v.ResourceRecords),
			})
		} else {
			// is an alias
			data = append(data, []string{
				strings.TrimSuffix(*v.Name, "."),
				string(v.Type),
				routingPolicy,
				differentiator,
				label,
				"-",
				"Yes",
				*v.AliasTarget.DNSName,
			})
		}
	}
	r.allData = data

	// Re-apply filter if active
	if r.isFiltered && r.lastFilter != "" {
		r.filterDataLive(r.lastFilter)
	} else {
		r.setTableData(data)
	}
}
