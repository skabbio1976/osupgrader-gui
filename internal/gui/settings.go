package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/skabbio1976/osupgrader-gui/internal/config"
)

func (a *App) showSettingsDialog() {
	// Guest credentials (username sparas, password endast i minnet)
	guestUserEntry := widget.NewEntry()
	guestUserEntry.SetText(a.config.Defaults.GuestUsername)
	if a.config.Defaults.GuestUsername == "" {
		guestUserEntry.SetText("Administrator")
	}

	guestPassEntry := widget.NewPasswordEntry()
	guestPassEntry.SetText(a.guestPassword)
	guestPassPlain := widget.NewEntry()
	guestPassPlain.SetText(a.guestPassword)
	guestPassPlain.Hide()

	// Toggle för att visa/dölja lösenord
	showPasswordCheck := widget.NewCheck(a.tr.ShowPassword, func(checked bool) {
		if checked {
			guestPassPlain.SetText(guestPassEntry.Text)
			guestPassEntry.Hide()
			guestPassPlain.Show()
		} else {
			guestPassEntry.SetText(guestPassPlain.Text)
			guestPassPlain.Hide()
			guestPassEntry.Show()
		}
	})

	// Upgrade settings
	snapshotPrefixEntry := widget.NewEntry()
	snapshotPrefixEntry.SetText(a.config.Defaults.SnapshotNamePrefix)

	isoPathEntry := widget.NewEntry()
	isoPathEntry.SetText(a.config.Defaults.IsoDatastorePath)

	parallelEntry := widget.NewEntry()
	parallelEntry.SetText(strconv.Itoa(a.config.Upgrade.Parallel))
	parallelEntry.SetPlaceHolder("10")

	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(strconv.Itoa(a.config.Upgrade.TimeoutMinutes))
	timeoutEntry.SetPlaceHolder("90")

	diskCheckEntry := widget.NewEntry()
	diskCheckEntry.SetText(strconv.Itoa(a.config.Upgrade.PrecheckDiskGB))

	skipMemoryCheck := widget.NewCheck(a.tr.SkipMemoryInSnapshot, nil)
	skipMemoryCheck.SetChecked(a.config.Defaults.SkipMemoryInSnapshot)

	rebootCheck := widget.NewCheck(a.tr.RebootAfterUpgrade, nil)
	rebootCheck.SetChecked(a.config.Upgrade.Reboot)

	// Dark mode toggle
	darkModeCheck := widget.NewCheck(a.tr.DarkMode, func(checked bool) {
		if checked {
			a.fyneApp.Settings().SetTheme(&darkTheme{})
		} else {
			a.fyneApp.Settings().SetTheme(&lightTheme{})
		}
		a.config.UI.DarkMode = checked
	})
	darkModeCheck.SetChecked(a.config.UI.DarkMode)

	// Language selection
	languageSelect := widget.NewRadioGroup([]string{a.tr.LanguageEnglish, a.tr.LanguageSwedish}, func(selected string) {
		// No action here, will be handled on save
	})
	// Set current language
	if a.config.UI.Language == "sv" {
		languageSelect.SetSelected(a.tr.LanguageSwedish)
	} else {
		languageSelect.SetSelected(a.tr.LanguageEnglish)
	}

	// Container för password-fält (stacked)
	passContainer := container.NewStack(guestPassEntry, guestPassPlain)

	// Info text för parallella uppgraderingar
	parallelInfo := widget.NewLabel(a.tr.ParallelUpgradesInfo)
	parallelInfo.Wrapping = fyne.TextWrapWord

	// Timeout-inställningar
	signalScriptTimeoutEntry := widget.NewEntry()
	signalScriptTimeoutEntry.SetText(strconv.Itoa(a.config.Timeouts.SignalScriptSeconds))

	signalFilesTimeoutEntry := widget.NewEntry()
	signalFilesTimeoutEntry.SetText(strconv.Itoa(a.config.Timeouts.SignalFilesMinutes))

	targetOsEntry := widget.NewEntry()
	targetOsEntry.SetText(strconv.Itoa(a.config.Timeouts.TargetOSMinutes))

	powerOffEntry := widget.NewEntry()
	powerOffEntry.SetText(strconv.Itoa(a.config.Timeouts.PowerOffMinutes))

	labeled := func(label string, obj fyne.CanvasObject) fyne.CanvasObject {
		if label == "" {
			return obj
		}
		return container.NewVBox(widget.NewLabel(label), obj)
	}

	guestTab := container.NewVScroll(container.NewVBox(
		labeled(a.tr.GuestUsername, guestUserEntry),
		labeled(a.tr.GuestPassword, passContainer),
		showPasswordCheck,
		widget.NewSeparator(),
		labeled(a.tr.SnapshotPrefix, snapshotPrefixEntry),
		labeled(a.tr.ISOPath, isoPathEntry),
	))

	upgradeTab := container.NewVScroll(container.NewVBox(
		labeled(a.tr.ParallelUpgrades, parallelEntry),
		parallelInfo,
		labeled(a.tr.TimeoutMinutes, timeoutEntry),
		labeled(a.tr.DiskPrecheckGB, diskCheckEntry),
		skipMemoryCheck,
		rebootCheck,
	))

	timeoutGrid := container.NewGridWithColumns(2,
		labeled(a.tr.SignalScriptSeconds, signalScriptTimeoutEntry),
		labeled(a.tr.SignalFilesMinutes, signalFilesTimeoutEntry),
		labeled(a.tr.OSVersionPollingMinutes, targetOsEntry),
		labeled(a.tr.PowerOffTimeoutMinutes, powerOffEntry),
	)

	timeoutsTab := container.NewVScroll(container.NewVBox(
		widget.NewLabel(a.tr.TimeoutsDescription),
		timeoutGrid,
	))

	uiTab := container.NewVScroll(container.NewVBox(
		widget.NewLabel(a.tr.Language+":"),
		languageSelect,
		widget.NewSeparator(),
		darkModeCheck,
	))

	tabs := container.NewAppTabs(
		container.NewTabItem(a.tr.TabGuestISO, guestTab),
		container.NewTabItem(a.tr.TabUpgrade, upgradeTab),
		container.NewTabItem(a.tr.TabTimeouts, timeoutsTab),
		container.NewTabItem(a.tr.TabUI, uiTab),
	)

	saveAction := func() {
		a.config.Defaults.GuestUsername = guestUserEntry.Text
		if guestPassEntry.Hidden {
			a.guestPassword = guestPassPlain.Text
		} else {
			a.guestPassword = guestPassEntry.Text
		}

		a.config.Defaults.SnapshotNamePrefix = snapshotPrefixEntry.Text
		a.config.Defaults.IsoDatastorePath = isoPathEntry.Text
		a.config.Defaults.SkipMemoryInSnapshot = skipMemoryCheck.Checked
		a.config.Upgrade.Reboot = rebootCheck.Checked

		if parallel, err := strconv.Atoi(parallelEntry.Text); err == nil {
			a.config.Upgrade.Parallel = parallel
		}
		if timeout, err := strconv.Atoi(timeoutEntry.Text); err == nil {
			a.config.Upgrade.TimeoutMinutes = timeout
		}
		if diskCheck, err := strconv.Atoi(diskCheckEntry.Text); err == nil {
			a.config.Upgrade.PrecheckDiskGB = diskCheck
		}

		if value, err := strconv.Atoi(signalScriptTimeoutEntry.Text); err == nil {
			a.config.Timeouts.SignalScriptSeconds = value
		}
		if value, err := strconv.Atoi(signalFilesTimeoutEntry.Text); err == nil {
			a.config.Timeouts.SignalFilesMinutes = value
		}
		if value, err := strconv.Atoi(targetOsEntry.Text); err == nil {
			a.config.Timeouts.TargetOSMinutes = value
		}
		if value, err := strconv.Atoi(powerOffEntry.Text); err == nil {
			a.config.Timeouts.PowerOffMinutes = value
		}

		// Handle language change
		newLang := "en"
		if languageSelect.Selected == a.tr.LanguageSwedish {
			newLang = "sv"
		}
		if newLang != a.config.UI.Language {
			a.SetLanguage(newLang)
			// Show message that restart is needed for full effect
			dialog.ShowInformation(a.tr.SettingsSaved, a.tr.SettingsSavedMessage+"\n\nNote: Please restart the application for all language changes to take effect.", a.window)
		}

		if err := config.Save(a.config); err != nil {
			dialog.ShowError(err, a.window)
		} else {
			dialog.ShowInformation(a.tr.SettingsSaved, a.tr.SettingsSavedMessage, a.window)
		}
	}

	saveButton := widget.NewButton(a.tr.SaveButton, saveAction)

	var settingsDialog dialog.Dialog
	closeButton := widget.NewButton(a.tr.CloseButton, func() {
		if settingsDialog != nil {
			settingsDialog.Hide()
		}
	})

	buttonRow := container.NewHBox(layout.NewSpacer(), saveButton, closeButton)

	header := container.NewVBox(
		widget.NewLabelWithStyle(a.tr.SettingsTitle, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)

	content := container.NewBorder(header, buttonRow, nil, nil, tabs)

	settingsDialog = dialog.NewCustomWithoutButtons(a.tr.SettingsTitle, content, a.window)
	settingsDialog.Resize(fyne.NewSize(900, 700))
	settingsDialog.Show()
}
