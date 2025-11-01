package internal

import (
	"strings"

	r53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bporter816/aws-tui/internal/repo"
	"github.com/bporter816/aws-tui/internal/view"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
)

type Route53RecordForm struct {
	*tview.Form
	view.Route53
	repo         *repo.Route53
	hostedZoneId string
	hostedZoneName string
	app          *Application
	mode         string // "create", "update", or "delete"
	existingRecord *r53Types.ResourceRecordSet
	onComplete   func()
}

func NewRoute53RecordForm(repo *repo.Route53, hostedZoneId, hostedZoneName, mode string, existingRecord *r53Types.ResourceRecordSet, app *Application, onComplete func()) *Route53RecordForm {
	form := tview.NewForm()

	r := &Route53RecordForm{
		Form:           form,
		repo:           repo,
		hostedZoneId:   hostedZoneId,
		hostedZoneName: hostedZoneName,
		app:            app,
		mode:           mode,
		existingRecord: existingRecord,
		onComplete:     onComplete,
	}

	r.buildForm()
	return r
}

func (r *Route53RecordForm) GetLabels() []string {
	switch r.mode {
	case "create":
		return []string{r.hostedZoneId, "Records", "Create Record"}
	case "update":
		return []string{r.hostedZoneId, "Records", "Update Record"}
	case "delete":
		return []string{r.hostedZoneId, "Records", "Delete Record"}
	default:
		return []string{r.hostedZoneId, "Records", "Form"}
	}
}

func (r *Route53RecordForm) GetKeyActions() []KeyAction {
	return []KeyAction{}
}

func (r *Route53RecordForm) Render() {
	// Form is already built
}

func (r *Route53RecordForm) buildForm() {
	var recordName, recordType, recordValue, recordTTL string

	// Normalize zone name (ensure it ends with a dot)
	zoneName := strings.TrimSuffix(r.hostedZoneName, ".")

	// Pre-fill form for update/delete modes
	if r.existingRecord != nil {
		if r.existingRecord.Name != nil {
			fullName := strings.TrimSuffix(*r.existingRecord.Name, ".")
			// Remove zone name suffix to show only the record prefix
			if strings.HasSuffix(fullName, zoneName) {
				recordName = strings.TrimSuffix(fullName, "."+zoneName)
				if recordName == zoneName {
					recordName = "@" // Apex record
				}
			} else {
				recordName = fullName
			}
		}
		recordType = string(r.existingRecord.Type)
		if r.existingRecord.TTL != nil {
			recordTTL = strconv.FormatInt(*r.existingRecord.TTL, 10)
		} else {
			recordTTL = "300"
		}

		// Get value from ResourceRecords or AliasTarget
		if r.existingRecord.AliasTarget != nil && r.existingRecord.AliasTarget.DNSName != nil {
			recordValue = *r.existingRecord.AliasTarget.DNSName
		} else if len(r.existingRecord.ResourceRecords) > 0 {
			var values []string
			for _, rr := range r.existingRecord.ResourceRecords {
				if rr.Value != nil {
					values = append(values, *rr.Value)
				}
			}
			recordValue = strings.Join(values, "\n")
		}
	} else {
		recordTTL = "300"
	}

	r.Form.Clear(true)

	if r.mode == "delete" {
		// Delete mode: show read-only fields and confirm button
		displayName := recordName + "." + zoneName
		r.Form.AddTextView("Record Name:", displayName, 0, 1, false, false)
		r.Form.AddTextView("Type:", recordType, 0, 1, false, false)
		r.Form.AddTextView("TTL:", recordTTL, 0, 1, false, false)
		r.Form.AddTextView("Value:", recordValue, 0, 5, false, false)
		r.Form.AddButton("Delete", r.deleteHandler)
		r.Form.AddButton("Cancel", r.cancelHandler)
	} else {
		// Create/Update mode: editable fields with zone suffix displayed
		recordNameLabel := "Record Name (." + zoneName + "):"
		r.Form.AddInputField(recordNameLabel, recordName, 50, nil, nil)

		// Record type dropdown
		recordTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "SRV", "PTR", "CAA"}
		currentIndex := 0
		for i, rt := range recordTypes {
			if rt == recordType {
				currentIndex = i
				break
			}
		}
		r.Form.AddDropDown("Type:", recordTypes, currentIndex, nil)

		r.Form.AddInputField("TTL:", recordTTL, 10, nil, nil)
		r.Form.AddTextArea("Value:", recordValue, 50, 5, 0, nil)

		if r.mode == "create" {
			r.Form.AddButton("Create", r.createHandler)
		} else {
			r.Form.AddButton("Update", r.updateHandler)
		}
		r.Form.AddButton("Cancel", r.cancelHandler)
	}

	// Set focus-related colors for better visibility
	r.Form.SetFieldBackgroundColor(tcell.ColorDarkBlue)
	r.Form.SetFieldTextColor(tcell.ColorWhite)
	r.Form.SetLabelColor(tcell.ColorYellow)
	r.Form.SetButtonBackgroundColor(tcell.ColorDarkCyan)
	r.Form.SetButtonTextColor(tcell.ColorWhite)

	r.Form.SetBorder(true)
	r.Form.SetBorderColor(tcell.ColorGreen)
	if r.mode == "delete" {
		r.Form.SetTitle(" Delete Record - Confirm ")
		r.Form.SetBorderColor(tcell.ColorRed)
	} else if r.mode == "update" {
		r.Form.SetTitle(" Update Record ")
		r.Form.SetBorderColor(tcell.ColorYellow)
	} else {
		r.Form.SetTitle(" Create Record ")
		r.Form.SetBorderColor(tcell.ColorGreen)
	}

	r.Form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			r.cancelHandler()
			return nil
		}
		return event
	})
}

func (r *Route53RecordForm) createHandler() {
	recordName := r.Form.GetFormItem(0).(*tview.InputField).GetText()
	recordTypeIndex, _ := r.Form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
	recordTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "SRV", "PTR", "CAA"}
	recordType := recordTypes[recordTypeIndex]
	ttlStr := r.Form.GetFormItem(2).(*tview.InputField).GetText()
	recordValue := r.Form.GetFormItem(3).(*tview.TextArea).GetText()

	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		r.showError("Invalid TTL value")
		return
	}

	if recordValue == "" {
		r.showError("Value is required")
		return
	}

	// Build full record name by appending zone name
	zoneName := strings.TrimSuffix(r.hostedZoneName, ".")
	var fullRecordName string

	recordName = strings.TrimSpace(recordName)
	if recordName == "" || recordName == "@" {
		// Apex record
		fullRecordName = zoneName + "."
	} else {
		// Subdomain or full name
		if strings.Contains(recordName, ".") && strings.HasSuffix(recordName, zoneName) {
			// Already contains zone name
			fullRecordName = recordName
		} else {
			// Prefix only, append zone name
			fullRecordName = recordName + "." + zoneName
		}
		// Ensure it ends with a dot
		if !strings.HasSuffix(fullRecordName, ".") {
			fullRecordName = fullRecordName + "."
		}
	}

	// Parse values (one per line)
	valueLines := strings.Split(recordValue, "\n")
	var resourceRecords []r53Types.ResourceRecord
	for _, line := range valueLines {
		line = strings.TrimSpace(line)
		if line != "" {
			resourceRecords = append(resourceRecords, r53Types.ResourceRecord{
				Value: aws.String(line),
			})
		}
	}

	if len(resourceRecords) == 0 {
		r.showError("At least one value is required")
		return
	}

	record := r53Types.ResourceRecordSet{
		Name:            aws.String(fullRecordName),
		Type:            r53Types.RRType(recordType),
		TTL:             aws.Int64(ttl),
		ResourceRecords: resourceRecords,
	}

	err = r.repo.CreateRecord(r.hostedZoneId, record)
	if err != nil {
		r.showError("Failed to create record: " + err.Error())
		return
	}

	r.onComplete()
	r.app.Close()
}

func (r *Route53RecordForm) updateHandler() {
	if r.existingRecord == nil {
		r.showError("No existing record to update")
		return
	}

	recordName := r.Form.GetFormItem(0).(*tview.InputField).GetText()
	recordTypeIndex, _ := r.Form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
	recordTypes := []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "SRV", "PTR", "CAA"}
	recordType := recordTypes[recordTypeIndex]
	ttlStr := r.Form.GetFormItem(2).(*tview.InputField).GetText()
	recordValue := r.Form.GetFormItem(3).(*tview.TextArea).GetText()

	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		r.showError("Invalid TTL value")
		return
	}

	if recordValue == "" {
		r.showError("Value is required")
		return
	}

	// Build full record name by appending zone name
	zoneName := strings.TrimSuffix(r.hostedZoneName, ".")
	var fullRecordName string

	recordName = strings.TrimSpace(recordName)
	if recordName == "" || recordName == "@" {
		// Apex record
		fullRecordName = zoneName + "."
	} else {
		// Subdomain or full name
		if strings.Contains(recordName, ".") && strings.HasSuffix(recordName, zoneName) {
			// Already contains zone name
			fullRecordName = recordName
		} else {
			// Prefix only, append zone name
			fullRecordName = recordName + "." + zoneName
		}
		// Ensure it ends with a dot
		if !strings.HasSuffix(fullRecordName, ".") {
			fullRecordName = fullRecordName + "."
		}
	}

	// Parse values (one per line)
	valueLines := strings.Split(recordValue, "\n")
	var resourceRecords []r53Types.ResourceRecord
	for _, line := range valueLines {
		line = strings.TrimSpace(line)
		if line != "" {
			resourceRecords = append(resourceRecords, r53Types.ResourceRecord{
				Value: aws.String(line),
			})
		}
	}

	if len(resourceRecords) == 0 {
		r.showError("At least one value is required")
		return
	}

	newRecord := r53Types.ResourceRecordSet{
		Name:            aws.String(fullRecordName),
		Type:            r53Types.RRType(recordType),
		TTL:             aws.Int64(ttl),
		ResourceRecords: resourceRecords,
	}

	err = r.repo.UpdateRecord(r.hostedZoneId, *r.existingRecord, newRecord)
	if err != nil {
		r.showError("Failed to update record: " + err.Error())
		return
	}

	r.onComplete()
	r.app.Close()
}

func (r *Route53RecordForm) deleteHandler() {
	if r.existingRecord == nil {
		r.showError("No record to delete")
		return
	}

	err := r.repo.DeleteRecord(r.hostedZoneId, *r.existingRecord)
	if err != nil {
		r.showError("Failed to delete record: " + err.Error())
		return
	}

	r.onComplete()
	r.app.Close()
}

func (r *Route53RecordForm) cancelHandler() {
	r.app.Close()
}

func (r *Route53RecordForm) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			r.app.pages.RemovePage("error")
		})
	r.app.pages.AddPage("error", modal, true, true)
}
