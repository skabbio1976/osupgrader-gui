package gui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/vmware/govmomi/object"
	"github.com/yourusername/osupgrader-gui/internal/debug"
	"github.com/yourusername/osupgrader-gui/internal/upgrade"
	"github.com/yourusername/osupgrader-gui/internal/vcenter"
)

func (a *App) showUpgradeScreen(selectedVMs map[string]bool) {
	// Räkna valda VMs
	var selectedNames []string
	for name, checked := range selectedVMs {
		if checked {
			selectedNames = append(selectedNames, name)
		}
	}

	// Titel
	title := widget.NewLabelWithStyle(
		fmt.Sprintf(a.tr.UpgradeVMs, len(selectedNames)),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Formulär för guest credentials - förifylla med sparade värden
	guestUserEntry := widget.NewEntry()
	if a.config.Defaults.GuestUsername != "" {
		guestUserEntry.SetText(a.config.Defaults.GuestUsername)
	} else {
		guestUserEntry.SetText("Administrator")
	}

	guestPassEntry := widget.NewPasswordEntry()
	if a.guestPassword != "" {
		guestPassEntry.SetText(a.guestPassword)
	}

	// ISO-path
	isoPathEntry := widget.NewEntry()
	if a.config.Defaults.IsoDatastorePath != "" {
		isoPathEntry.SetText(a.config.Defaults.IsoDatastorePath)
	}
	isoPathEntry.SetPlaceHolder("[datastore1] iso/windows-server-2022.iso")

	// Snapshot-alternativ
	createSnapshotCheck := widget.NewCheck(a.tr.CreateSnapshot, nil)
	createSnapshotCheck.SetChecked(true)

	// Progress-widget
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = float64(len(selectedNames))

	statusLabel := widget.NewLabel(fmt.Sprintf(a.tr.ReadyToStart, len(selectedNames)))
	logText := widget.NewMultiLineEntry()
	logText.Wrapping = fyne.TextWrapWord
	// Disable() borttaget så texten är läsbar och användaren kan markera/kopiera

	// Skriv ut valda servrar direkt när skärmen laddas
	logText.SetText(fmt.Sprintf("=== VALDA SERVRAR (%d st) ===\n", len(selectedNames)))
	for i, vmName := range selectedNames {
		logText.SetText(logText.Text + fmt.Sprintf("%d. %s\n", i+1, vmName))
	}
	logText.SetText(logText.Text + "===========================\n\n")
	logText.SetText(logText.Text + a.tr.FillInDetails)

	// Logga också valda servrar
	debug.Log("=== UPGRADE SCREEN - VALDA SERVRAR (%d st) ===", len(selectedNames))
	for i, vmName := range selectedNames {
		debug.Log("  %d. %s", i+1, vmName)
	}

	// Deklarera knappar först för att undvika scope-problem
	var startBtn *widget.Button
	var backBtn *widget.Button
	startBtn = widget.NewButton(a.tr.StartUpgrade, func() {
		guestUser := guestUserEntry.Text
		guestPass := guestPassEntry.Text
		isoPath := isoPathEntry.Text

		// Debug-logga input från GUI
		debug.Log("=== GUI INPUT DEBUG ===")
		debug.Log("GUI guestUser: %s (len=%d)", guestUser, len(guestUser))
		// debug.Log("GUI guestPass: %s (len=%d)", guestPass, len(guestPass)) // Kommenterad av säkerhetsskäl
		debug.Log("GUI guestPass length: %d", len(guestPass))
		debug.Log("GUI isoPath: %s", isoPath)
		debug.Log("=======================")

		if guestUser == "" || guestPass == "" || isoPath == "" {
			dialog.ShowError(fmt.Errorf("%s", a.tr.FillAllFields), a.window)
			return
		}

		// Validera ISO först
		statusLabel.SetText(a.tr.ValidatingISO)
		logText.SetText(logText.Text + fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("15:04:05"), a.tr.ValidatingISO, isoPath))

		go func() {
			ctx := context.Background()
			if err := upgrade.ValidateISOPath(ctx, isoPath); err != nil {
				statusLabel.SetText(a.tr.ISOValidationFailed)
				logText.SetText(logText.Text + fmt.Sprintf("[%s] ERROR: %v\n", time.Now().Format("15:04:05"), err))
				dialog.ShowError(fmt.Errorf("%s: %v", a.tr.ISOValidationFailed, err), a.window)
				return
			}

			statusLabel.SetText(a.tr.ISOOK)
			logText.SetText(logText.Text + fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), a.tr.ISOValidated))
			logText.SetText(logText.Text + fmt.Sprintf("[%s] "+a.tr.StartingUpgrade, time.Now().Format("15:04:05"), len(selectedNames)))

			// Logga start till debug-logg
			debug.Log("=== STARTAR UPPGRADERING AV %d SERVRAR ===", len(selectedNames))

			startBtn.Disable()
			backBtn.Disable()

			// Parallell uppgradering med worker pool
			maxWorkers := a.config.Upgrade.Parallel
			if maxWorkers <= 0 {
				maxWorkers = 10 // Fallback om config är felaktig
			}
			if maxWorkers > len(selectedNames) {
				maxWorkers = len(selectedNames)
			}

			debug.Log("Starting parallel upgrade with %d workers for %d VMs (config.Upgrade.Parallel=%d)", maxWorkers, len(selectedNames), a.config.Upgrade.Parallel)

			// Channels och counters
			type upgradeJob struct {
				vmName string
				vmInfo vcenter.VMInfo
			}

			type upgradeResult struct {
				vmName string
				err    error
			}

			jobs := make(chan upgradeJob, len(selectedNames))
			results := make(chan upgradeResult, len(selectedNames))
			var wg sync.WaitGroup
			var mu sync.Mutex // För thread-safe GUI updates

			completed := 0
			failures := 0

			// Starta workers
			for w := 1; w <= maxWorkers; w++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()
					debug.Log("Worker %d started", workerID)

					for job := range jobs {
						debug.Log("Worker %d processing VM: %s", workerID, job.vmName)

						// Skapa VM-objekt
						client := a.GetClient()
						vm := object.NewVirtualMachine(client.GetVim(), job.vmInfo.Ref)

						// Skapa snapshot-namn med timestamp och VM-namn
						snapshotName := fmt.Sprintf("%s-pre-%s-%s", a.config.Defaults.SnapshotNamePrefix, job.vmName, time.Now().Format("20060102-150405"))

						// Uppgraderingsalternativ
						opts := upgrade.UpgradeOptions{
							VMInfo:         job.vmInfo,
							GuestUsername:  guestUser,
							GuestPassword:  guestPass,
							ISOPath:        isoPath,
							CreateSnapshot: createSnapshotCheck.Checked,
							SnapshotName:   snapshotName,
							Config:         a.config,
						}

						// Thread-safe log update - startar
						mu.Lock()
						logText.SetText(logText.Text + fmt.Sprintf("\n[%s] === %s === (Worker %d)\n", time.Now().Format("15:04:05"), job.vmName, workerID))
						mu.Unlock()

						// Kör uppgradering
						err := upgrade.UpgradeSingleVM(vm, opts)

						// Skicka resultat
						results <- upgradeResult{
							vmName: job.vmName,
							err:    err,
						}
					}

					debug.Log("Worker %d finished", workerID)
				}(w)
			}

			// Skicka alla jobb till workers
			for _, vmName := range selectedNames {
				// Hitta VMInfo
				var vmInfo vcenter.VMInfo
				for _, vm := range a.GetVMs() {
					if vm.Name == vmName {
						vmInfo = vm
						break
					}
				}

				jobs <- upgradeJob{
					vmName: vmName,
					vmInfo: vmInfo,
				}
			}
			close(jobs)

			// Samla resultat i en goroutine
			go func() {
				wg.Wait()
				close(results)
			}()

			// Hantera resultat och uppdatera UI
			for result := range results {
				mu.Lock()

				completed++
				if result.err != nil {
					failures++
					logText.SetText(logText.Text + fmt.Sprintf("[%s] "+a.tr.UpgradeFailed+"\n", time.Now().Format("15:04:05"), result.vmName, result.err))
					statusLabel.SetText(fmt.Sprintf(a.tr.VMCompleteStatus,
						completed, len(selectedNames), result.vmName, completed-failures, failures))
				} else {
					logText.SetText(logText.Text + fmt.Sprintf("[%s] "+a.tr.UpgradeCompleted+"\n", time.Now().Format("15:04:05"), result.vmName))
					statusLabel.SetText(fmt.Sprintf(a.tr.VMSuccessStatus,
						completed, len(selectedNames), result.vmName, completed-failures, failures))
				}

				progressBar.SetValue(float64(completed))

				mu.Unlock()
			}

			// Klart - ingen popup, bara status och logg
			statusLabel.SetText(fmt.Sprintf(a.tr.AllCompleteStatus, completed-failures, len(selectedNames), failures))
			logText.SetText(logText.Text + fmt.Sprintf("\n[%s] %s\n", time.Now().Format("15:04:05"), a.tr.SummaryHeader))
			logText.SetText(logText.Text + fmt.Sprintf("[%s] "+a.tr.SummaryTotal+"\n", time.Now().Format("15:04:05"), len(selectedNames)))
			logText.SetText(logText.Text + fmt.Sprintf("[%s] "+a.tr.SummarySucceeded+"\n", time.Now().Format("15:04:05"), completed-failures))
			logText.SetText(logText.Text + fmt.Sprintf("[%s] "+a.tr.SummaryFailed+"\n", time.Now().Format("15:04:05"), failures))
			if failures == 0 {
				logText.SetText(logText.Text + fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), a.tr.AllSuccessful))
			} else {
				logText.SetText(logText.Text + fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), a.tr.SomeFailed))
			}
			logText.SetText(logText.Text + fmt.Sprintf("[%s] =======================\n\n", time.Now().Format("15:04:05")))

			startBtn.Enable()
			backBtn.Enable()
		}()
	})

	// Tillbaka-knapp
	backBtn = widget.NewButton(a.tr.Back, func() {
		a.showVMSelectionScreen()
	})

	// Inställningar-knapp
	settingsBtn := widget.NewButton(a.tr.SettingsButton, func() {
		a.showSettingsDialog()
	})

	// Info-text om att spara credentials
	infoText := widget.NewLabel(a.tr.TipSaveCredentials)
	infoText.Wrapping = fyne.TextWrapWord

	// Log scroll
	logScroll := container.NewVScroll(logText)
	logScroll.SetMinSize(fyne.NewSize(900, 300))

	// Layout
	form := container.NewVBox(
		infoText,
		widget.NewForm(
			widget.NewFormItem(a.tr.GuestAdminUser+":", guestUserEntry),
			widget.NewFormItem(a.tr.GuestAdminPassword+":", guestPassEntry),
			widget.NewFormItem(a.tr.ISODatastorePath+":", isoPathEntry),
		),
		createSnapshotCheck,
	)

	content := container.NewBorder(
		container.NewVBox(
			title,
			form,
		),
		container.NewVBox(
			progressBar,
			statusLabel,
			container.NewHBox(backBtn, settingsBtn, startBtn),
		),
		nil,
		nil,
		logScroll,
	)

	a.window.SetContent(content)
}
