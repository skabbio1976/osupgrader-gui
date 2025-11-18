package gui

import (
	"fmt"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/skabbio1976/osupgrader-gui/internal/debug"
	"github.com/skabbio1976/osupgrader-gui/internal/vcenter"
)

func (a *App) showVMSelectionScreen() {
	vms := a.GetVMs()

	// Titel
	title := widget.NewLabelWithStyle(
		fmt.Sprintf(a.tr.VMSelectionTitleCount, len(vms)),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Sökfilter (stödjer regex)
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(a.tr.SearchPlaceholder)

	// Selection state
	selectedVMs := make(map[string]bool)

	// Filtrerad lista av VMs
	var filteredVMs []vcenter.VMInfo

	// Beräkna dynamiska kolumnbredder baserat på innehåll
	calculateColumnWidths := func() (float32, float32, float32, float32, float32) {
		const charWidth = 8.0 // Ungefärlig bredd per tecken
		const minWidth = 80.0
		const padding = 20.0

		// Header text som minimum
		maxNameLen := len("Name")
		maxFolderLen := len("Folder")
		maxDomainLen := len("Domain")
		maxOSLen := len("OS")

		// Hitta maxlängd för varje kolumn
		for _, vm := range vms {
			if len(vm.Name) > maxNameLen {
				maxNameLen = len(vm.Name)
			}
			if len(vm.Folder) > maxFolderLen {
				maxFolderLen = len(vm.Folder)
			}
			if len(vm.Domain) > maxDomainLen {
				maxDomainLen = len(vm.Domain)
			}
			if len(vm.OS) > maxOSLen {
				maxOSLen = len(vm.OS)
			}
		}

		// Beräkna bredder med padding
		checkboxWidth := float32(60)
		nameWidth := float32(maxNameLen)*charWidth + padding
		folderWidth := float32(maxFolderLen)*charWidth + padding
		domainWidth := float32(maxDomainLen)*charWidth + padding
		osWidth := float32(maxOSLen)*charWidth + padding

		// Sätt minimum bredder
		if nameWidth < minWidth {
			nameWidth = minWidth
		}
		if folderWidth < minWidth {
			folderWidth = minWidth
		}
		if domainWidth < minWidth {
			domainWidth = minWidth
		}
		if osWidth < minWidth {
			osWidth = minWidth
		}

		return checkboxWidth, nameWidth, folderWidth, domainWidth, osWidth
	}

	// Skapa tabell
	table := widget.NewTable(
		func() (int, int) {
			// Antal rader (VMs + 1 header rad) och antal kolumner
			return len(filteredVMs) + 1, 5
		},
		func() fyne.CanvasObject {
			// Skapa cell templates med mindre font
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
					label.SetText(a.tr.ColumnName)
					label.Show()
				case 2:
					label.SetText(a.tr.ColumnFolder)
					label.Show()
				case 3:
					label.SetText(a.tr.ColumnDomain)
					label.Show()
				case 4:
					label.SetText(a.tr.ColumnOS)
					label.Show()
				}
				return
			}

			// Data rader
			vmIndex := id.Row - 1
			if vmIndex >= len(filteredVMs) {
				label.SetText("")
				check.Hide()
				return
			}

			vm := filteredVMs[vmIndex]

			switch id.Col {
			case 0:
				// Checkbox kolumn
				label.Hide()
				check.Show()
				// Capture VM name in closure properly
				vmName := vm.Name
				// Sätt OnChanged till nil innan SetChecked för att undvika trigger
				check.OnChanged = nil
				check.SetChecked(selectedVMs[vmName])
				// Sätt callback efter SetChecked
				check.OnChanged = func(checked bool) {
					selectedVMs[vmName] = checked
				}
			case 1:
				// Name kolumn
				label.Show()
				check.Hide()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(vm.Name)
			case 2:
				// Folder kolumn
				label.Show()
				check.Hide()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(vm.Folder)
			case 3:
				// Domain kolumn
				label.Show()
				check.Hide()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(vm.Domain)
			case 4:
				// OS kolumn
				label.Show()
				check.Hide()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(vm.OS)
			}
		},
	)

	// Sätt dynamiska kolumnbredder
	checkboxW, nameW, folderW, domainW, osW := calculateColumnWidths()
	table.SetColumnWidth(0, checkboxW) // Checkbox
	table.SetColumnWidth(1, nameW)     // Name
	table.SetColumnWidth(2, folderW)   // Folder
	table.SetColumnWidth(3, domainW)   // Domain
	table.SetColumnWidth(4, osW)       // OS

	// Logga kolumnbredder för debugging
	totalWidth := checkboxW + nameW + folderW + domainW + osW
	debug.Log("Dynamic column widths: Checkbox=%.0f, Name=%.0f, Folder=%.0f, Domain=%.0f, OS=%.0f, Total=%.0f",
		checkboxW, nameW, folderW, domainW, osW, totalWidth)

	// Justera fönsterstorlek om nödvändigt (lägg till padding för UI-element)
	minWindowWidth := totalWidth + 100 // Extra för scrollbar och padding
	currentSize := a.window.Canvas().Size()
	if minWindowWidth > float32(currentSize.Width) {
		a.window.Resize(fyne.NewSize(minWindowWidth, float32(currentSize.Height)))
		debug.Log("Window resized to width: %.0f", minWindowWidth)
	}

	// Funktion för att uppdatera filtrerad lista (med regex-stöd)
	updateFilteredList := func(filter string) {
		filteredVMs = []vcenter.VMInfo{}

		if filter == "" {
			filteredVMs = vms
			table.Refresh()
			return
		}

		// Försök kompilera som regex (case-insensitive)
		re, err := regexp.Compile("(?i)" + filter)
		useRegex := err == nil

		if useRegex {
			debug.Log("Filter using regex: %s", filter)
		} else {
			debug.Log("Filter using substring (invalid regex): %s", filter)
		}

		for _, vm := range vms {
			matched := false

			if useRegex {
				// Regex-matchning mot alla fält
				matched = re.MatchString(vm.Name) ||
					re.MatchString(vm.Folder) ||
					re.MatchString(vm.Domain) ||
					re.MatchString(vm.OS)
			} else {
				// Fallback till case-insensitive substring
				filterLower := strings.ToLower(filter)
				matched = strings.Contains(strings.ToLower(vm.Name), filterLower) ||
					strings.Contains(strings.ToLower(vm.Folder), filterLower) ||
					strings.Contains(strings.ToLower(vm.Domain), filterLower) ||
					strings.Contains(strings.ToLower(vm.OS), filterLower)
			}

			if matched {
				filteredVMs = append(filteredVMs, vm)
			}
		}
		table.Refresh()
	}

	// Initial filtrering
	updateFilteredList("")

	// Uppdatera när användaren söker
	searchEntry.OnChanged = func(text string) {
		updateFilteredList(text)
	}

	// Välj alla / Avmarkera alla
	selectAllBtn := widget.NewButton(a.tr.SelectAll, func() {
		for _, vm := range filteredVMs {
			selectedVMs[vm.Name] = true
		}
		table.Refresh()
	})
	selectAllBtn.Importance = widget.HighImportance

	deselectAllBtn := widget.NewButton(a.tr.DeselectAll, func() {
		for _, vm := range filteredVMs {
			selectedVMs[vm.Name] = false
		}
		table.Refresh()
	})
	deselectAllBtn.Importance = widget.HighImportance

	// Fortsätt-knapp
	continueBtn := widget.NewButton(a.tr.ContinueToUpgrade, func() {
		// Räkna valda VMs
		count := 0
		for _, checked := range selectedVMs {
			if checked {
				count++
			}
		}

		if count == 0 {
			dialog.ShowInformation(a.tr.NoVMsSelected, a.tr.SelectVMsFirst, a.window)
			return
		}

		// Gå till upgrade-skärm
		a.showUpgradeScreen(selectedVMs)
	})

	// Tillbaka-knapp
	backBtn := widget.NewButton(a.tr.LogOut, func() {
		a.SetClient(nil)
		a.showLoginScreen()
	})

	// Refresh-knapp
	refreshBtn := widget.NewButton(a.tr.RefreshList, func() {
		dialog.ShowInformation(a.tr.Refreshing, a.tr.RefreshingMessage, a.window)
		go func() {
			vms, err := vcenter.GetVMInfos()
			if err != nil {
				dialog.ShowError(fmt.Errorf(a.tr.ErrorRefreshVMs, err), a.window)
				return
			}
			a.SetVMs(vms)
			a.showVMSelectionScreen()
		}()
	})
	refreshBtn.Importance = widget.HighImportance

	// Snapshot-hantering knapp
	snapshotBtn := widget.NewButton(a.tr.ManageSnapshots, func() {
		a.showSnapshotManagementScreen()
	})
	snapshotBtn.Importance = widget.HighImportance

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			title,
			searchEntry,
			container.NewHBox(selectAllBtn, deselectAllBtn, refreshBtn, snapshotBtn),
		),
		container.NewHBox(backBtn, continueBtn),
		nil,
		nil,
		table,
	)

	a.window.SetContent(content)
}
