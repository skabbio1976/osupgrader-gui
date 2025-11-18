package gui

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/vmware/govmomi/object"
	"github.com/yourusername/osupgrader-gui/internal/debug"
	"github.com/yourusername/osupgrader-gui/internal/vcenter"
)

// SnapshotInfo innehåller information om en snapshot som kan tas bort
type SnapshotInfo struct {
	VMName       string
	SnapshotName string
	Ref          string // För identifiering i UI
	SnapRef      vcenter.SnapshotEntry
}

func (a *App) showSnapshotManagementScreen() {
	// Titel
	title := widget.NewLabelWithStyle(
		a.tr.SnapshotsTitle,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Status label
	statusLabel := widget.NewLabel(a.tr.SnapshotsLoading)

	// Progress
	progress := widget.NewProgressBarInfinite()

	// Lista för snapshots
	var snapshotList []SnapshotInfo
	var selectedSnapshots map[string]bool

	// Tabell (skapas senare)
	var table *widget.Table

	// Sökfilter (stödjer regex)
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(a.tr.SearchPlaceholder)
	searchEntry.Disable()

	// Filtrerad lista
	var filteredSnapshots []SnapshotInfo

	// Funktion för att uppdatera filtrerad lista (med regex-stöd)
	updateFilteredList := func(filter string) {
		filteredSnapshots = []SnapshotInfo{}

		if filter == "" {
			filteredSnapshots = snapshotList
			if table != nil {
				table.Refresh()
			}
			return
		}

		// Försök kompilera som regex (case-insensitive)
		re, err := regexp.Compile("(?i)" + filter)
		useRegex := err == nil

		if useRegex {
			debug.Log("Snapshot filter using regex: %s", filter)
		} else {
			debug.Log("Snapshot filter using substring (invalid regex): %s", filter)
		}

		for _, snap := range snapshotList {
			matched := false

			if useRegex {
				// Regex-matchning mot alla fält
				matched = re.MatchString(snap.VMName) || re.MatchString(snap.SnapshotName)
			} else {
				// Fallback till case-insensitive substring
				filterLower := strings.ToLower(filter)
				matched = strings.Contains(strings.ToLower(snap.VMName), filterLower) ||
					strings.Contains(strings.ToLower(snap.SnapshotName), filterLower)
			}

			if matched {
				filteredSnapshots = append(filteredSnapshots, snap)
			}
		}

		if table != nil {
			table.Refresh()
		}
	}

	// Skapa tabell
	table = widget.NewTable(
		func() (int, int) {
			return len(filteredSnapshots) + 1, 3
		},
		func() fyne.CanvasObject {
			// Skapa cell templates med stretch
			label := widget.NewLabel("Template")
			label.TextStyle.Monospace = false
			label.Truncation = fyne.TextTruncateClip
			return container.NewStack(
				label,
				widget.NewCheck("", nil),
			)
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			container := cell.(*fyne.Container)
			label := container.Objects[0].(*widget.Label)
			check := container.Objects[1].(*widget.Check)

			// Header rad
			if id.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				check.Hide()

				switch id.Col {
				case 0:
					label.SetText(a.tr.ColumnSelect)
					label.Show()
				case 1:
					label.SetText("VM Name")
					label.Show()
				case 2:
					label.SetText("Snapshot Name")
					label.Show()
				}
				return
			}

			// Data rader
			snapIndex := id.Row - 1
			if snapIndex >= len(filteredSnapshots) {
				label.SetText("")
				check.Hide()
				return
			}

			snap := filteredSnapshots[snapIndex]

			switch id.Col {
			case 0:
				// Checkbox kolumn
				label.Hide()
				check.Show()
				// Capture ref in closure properly
				snapRef := snap.Ref
				// Sätt OnChanged till nil innan SetChecked för att undvika trigger
				check.OnChanged = nil
				check.SetChecked(selectedSnapshots[snapRef])
				// Sätt callback efter SetChecked
				check.OnChanged = func(checked bool) {
					selectedSnapshots[snapRef] = checked
				}
			case 1:
				// VM Name
				label.Show()
				check.Hide()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(snap.VMName)
			case 2:
				// Snapshot Name
				label.Show()
				check.Hide()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(snap.SnapshotName)
			}
		},
	)

	// Sätt kolumnbredder
	table.SetColumnWidth(0, 60)  // Checkbox
	table.SetColumnWidth(1, 250) // VM Name
	table.SetColumnWidth(2, 450) // Snapshot Name

	searchEntry.OnChanged = func(text string) {
		updateFilteredList(text)
	}

	// Deklarera knappar (callbacks sätts senare för att undvika scope-problem)
	var selectAllBtn, deselectAllBtn, removeBtn, refreshBtn *widget.Button

	// Välj alla / Avmarkera alla
	selectAllBtn = widget.NewButton(a.tr.SelectAll, func() {
		for _, snap := range filteredSnapshots {
			selectedSnapshots[snap.Ref] = true
		}
		table.Refresh()
	})
	selectAllBtn.Importance = widget.HighImportance
	selectAllBtn.Disable()

	deselectAllBtn = widget.NewButton(a.tr.DeselectAll, func() {
		for _, snap := range filteredSnapshots {
			selectedSnapshots[snap.Ref] = false
		}
		table.Refresh()
	})
	deselectAllBtn.Importance = widget.HighImportance
	deselectAllBtn.Disable()

	// Ta bort valda-knapp
	removeBtn = widget.NewButton(a.tr.RemoveSelected, func() {
		// Räkna valda
		count := 0
		for _, checked := range selectedSnapshots {
			if checked {
				count++
			}
		}

		if count == 0 {
			dialog.ShowInformation(a.tr.NoSnapshotsSelected, a.tr.SelectSnapshotsFirst, a.window)
			return
		}

		// Bekräftelse
		dialog.ShowConfirm(a.tr.RemoveConfirmTitle,
			fmt.Sprintf(a.tr.RemoveConfirmMessage, count),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				// Ta bort snapshots
				go a.removeSelectedSnapshots(filteredSnapshots, selectedSnapshots, statusLabel, selectAllBtn, deselectAllBtn, removeBtn)
			}, a.window)
	})
	removeBtn.Disable()

	// Tillbaka-knapp
	backBtn := widget.NewButton("Tillbaka", func() {
		a.showVMSelectionScreen()
	})

	// Uppdatera lista-knapp
	refreshBtn = widget.NewButton(a.tr.RefreshList, func() {
		go a.loadSnapshots(statusLabel, progress, searchEntry, selectAllBtn, deselectAllBtn, removeBtn, &snapshotList, &selectedSnapshots, updateFilteredList)
	})
	refreshBtn.Importance = widget.HighImportance
	refreshBtn.Disable()

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			title,
			statusLabel,
			progress,
			searchEntry,
			container.NewHBox(selectAllBtn, deselectAllBtn, refreshBtn),
		),
		container.NewHBox(backBtn, removeBtn),
		nil,
		nil,
		table,
	)

	a.window.SetContent(content)

	// Starta laddning av snapshots
	selectedSnapshots = make(map[string]bool)
	go a.loadSnapshots(statusLabel, progress, searchEntry, selectAllBtn, deselectAllBtn, removeBtn, &snapshotList, &selectedSnapshots, updateFilteredList)
}

func (a *App) loadSnapshots(statusLabel *widget.Label, progress *widget.ProgressBarInfinite, searchEntry *widget.Entry,
	selectAllBtn, deselectAllBtn, removeBtn *widget.Button, snapshotList *[]SnapshotInfo, selectedSnapshots *map[string]bool, updateFilteredList func(string)) {

	debug.Log("Loading snapshots from all VMs...")
	statusLabel.SetText(a.tr.SnapshotsLoading)
	progress.Show()
	progress.Start()

	ctx := context.Background()
	vms := a.GetVMs()
	client := a.GetClient()

	prefix := a.config.Defaults.SnapshotNamePrefix
	if prefix == "" {
		prefix = "pre-upgrade"
	}

	*snapshotList = []SnapshotInfo{}
	found := 0

	for _, vmInfo := range vms {
		vm := object.NewVirtualMachine(client.GetVim(), vmInfo.Ref)
		snapshots, err := vcenter.ListSnapshots(ctx, vm, vmInfo.Name)
		if err != nil {
			debug.Log("WARNING: Failed to list snapshots for %s: %v", vmInfo.Name, err)
			continue
		}

		// Filtrera på pre-upgrade snapshots
		for _, snap := range snapshots {
			if strings.Contains(strings.ToLower(snap.SnapshotName), strings.ToLower(prefix)) {
				*snapshotList = append(*snapshotList, SnapshotInfo{
					VMName:       snap.VMName,
					SnapshotName: snap.SnapshotName,
					Ref:          snap.Ref.Value,
					SnapRef:      snap,
				})
				found++
			}
		}
	}

	progress.Stop()
	progress.Hide()

	if found == 0 {
		statusLabel.SetText(fmt.Sprintf("Inga pre-upgrade snapshots hittades (prefix: '%s')", prefix))
		debug.Log("No pre-upgrade snapshots found")
	} else {
		statusLabel.SetText(fmt.Sprintf("%d pre-upgrade snapshot(s) hittade", found))
		debug.Log("Found %d pre-upgrade snapshots", found)
		searchEntry.Enable()
		selectAllBtn.Enable()
		deselectAllBtn.Enable()
		removeBtn.Enable()
	}

	*selectedSnapshots = make(map[string]bool)
	updateFilteredList("")
}

func (a *App) removeSelectedSnapshots(snapshots []SnapshotInfo, selected map[string]bool,
	statusLabel *widget.Label, selectAllBtn, deselectAllBtn, removeBtn *widget.Button) {

	debug.Log("Starting snapshot removal process...")
	ctx := context.Background()

	// Räkna valda
	toRemove := []SnapshotInfo{}
	for _, snap := range snapshots {
		if selected[snap.Ref] {
			toRemove = append(toRemove, snap)
		}
	}

	removeBtn.Disable()
	selectAllBtn.Disable()
	deselectAllBtn.Disable()

	var removed, failed atomic.Int32
	var wg sync.WaitGroup

	statusLabel.SetText(fmt.Sprintf(a.tr.RemovingSnapshots, len(toRemove)))
	debug.Log("Starting parallel removal of %d snapshots", len(toRemove))

	// Starta parallella goroutines för varje snapshot
	for _, snap := range toRemove {
		wg.Add(1)
		go func(s SnapshotInfo) {
			defer wg.Done()

			debug.Log("Removing snapshot: %s on VM %s", s.SnapshotName, s.VMName)

			if err := vcenter.RemoveSnapshot(ctx, s.SnapRef.Ref); err != nil {
				debug.LogError("RemoveSnapshot", err, "VM", s.VMName, "Snapshot", s.SnapshotName)
				failed.Add(1)
			} else {
				debug.LogSuccess("RemoveSnapshot", "VM", s.VMName, "Snapshot", s.SnapshotName)
				removed.Add(1)
			}
		}(snap)
	}

	// Vänta på att alla goroutines är klara
	wg.Wait()

	removedCount := removed.Load()
	failedCount := failed.Load()

	if failedCount == 0 {
		statusLabel.SetText(fmt.Sprintf(a.tr.RemoveSuccessCount, removedCount))
		debug.Log("Snapshot removal completed successfully: %d removed", removedCount)
	} else {
		statusLabel.SetText(fmt.Sprintf(a.tr.RemoveErrorCount, removedCount, failedCount))
		debug.Log("Snapshot removal completed with errors: %d removed, %d failed", removedCount, failedCount)
	}

	removeBtn.Enable()
	selectAllBtn.Enable()
	deselectAllBtn.Enable()

	// Reload snapshot list
	go a.showSnapshotManagementScreen()
}
