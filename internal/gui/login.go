package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/yourusername/osupgrader-gui/internal/vcenter"
)

func (a *App) showLoginScreen() {
	// Skapa formulär
	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("vcenter.example.local")
	if a.config.VCenter.Host != "" {
		hostEntry.SetText(a.config.VCenter.Host)
	}

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("administrator@vsphere.local")
	if a.config.VCenter.Username != "" {
		usernameEntry.SetText(a.config.VCenter.Username)
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Lösenord")

	insecureCheck := widget.NewCheck("Tillåt osignerade certifikat", nil)
	insecureCheck.SetChecked(a.config.VCenter.Insecure)

	// Autentiseringsmetod
	authMethodSelect := widget.NewRadioGroup([]string{"Lösenord", "Windows SSPI/Kerberos"}, func(selected string) {
		if selected == "Windows SSPI/Kerberos" {
			usernameEntry.Disable()
			passwordEntry.Disable()
		} else {
			usernameEntry.Enable()
			passwordEntry.Enable()
		}
	})
	authMethodSelect.SetSelected("Lösenord")
	if a.config.VCenter.Mode == "sspi" {
		authMethodSelect.SetSelected("Windows SSPI/Kerberos")
	}

	statusLabel := widget.NewLabel("")

	// Login-knapp (deklarera först)
	var loginBtn *widget.Button
	loginBtn = widget.NewButton("Logga in", func() {
		host := hostEntry.Text
		authMethod := authMethodSelect.Selected

		if host == "" {
			dialog.ShowError(fmt.Errorf("vCenter host måste anges"), a.window)
			return
		}

		// Uppdatera config
		a.config.VCenter.Host = host
		a.config.VCenter.Insecure = insecureCheck.Checked

		var client *vcenter.Client
		var err error

		if authMethod == "Windows SSPI/Kerberos" {
			// SSPI-inloggning
			if usernameEntry.Text != "" {
				a.config.VCenter.Username = usernameEntry.Text
			}
			a.config.VCenter.Mode = "sspi"

			statusLabel.SetText("Loggar in med SSPI/Kerberos...")
			loginBtn.Disable()

			// Logga in i bakgrunden
			go func() {
				client, err = vcenter.LoginSSPI(&a.config.VCenter)
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf("SSPI-inloggning misslyckades: %v - OBS: SSPI/Kerberos fungerar endast på Windows och kräver att du är inloggad med ett domänkonto", err), a.window)
					return
				}

				a.SetClient(client)
				statusLabel.SetText("Inloggad! Hämtar VMs...")

				// Hämta VMs
				vms, err := vcenter.GetVMInfos()
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf("kunde inte hämta VMs: %v", err), a.window)
					return
				}

				a.SetVMs(vms)

				// Byt till VM-selection-skärm
				a.showVMSelectionScreen()
			}()
		} else {
			// Lösenordsinloggning
			username := usernameEntry.Text
			password := passwordEntry.Text

			if username == "" || password == "" {
				dialog.ShowError(fmt.Errorf("användarnamn och lösenord måste fyllas i"), a.window)
				return
			}

			a.config.VCenter.Username = username
			a.config.VCenter.Mode = "password"

			statusLabel.SetText("Loggar in...")
			loginBtn.Disable()

			// Logga in i bakgrunden
			go func() {
				client, err = vcenter.Login(&a.config.VCenter, password)
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf("inloggning misslyckades: %v", err), a.window)
					return
				}

				a.SetClient(client)
				statusLabel.SetText("Inloggad! Hämtar VMs...")

				// Hämta VMs
				vms, err := vcenter.GetVMInfos()
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf("kunde inte hämta VMs: %v", err), a.window)
					return
				}

				a.SetVMs(vms)

				// Byt till VM-selection-skärm
				a.showVMSelectionScreen()
			}()
		}
	})

	// Inställningar-knapp
	settingsBtn := widget.NewButton("Inställningar", func() {
		a.showSettingsDialog()
	})

	// Layout
	form := container.NewVBox(
		widget.NewLabelWithStyle("OSUpgrader - Windows Server Upgrade Tool", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
		widget.NewForm(
			widget.NewFormItem("vCenter Host:", hostEntry),
		),
		widget.NewLabel("Autentiseringsmetod:"),
		authMethodSelect,
		widget.NewLabel(""),
		widget.NewForm(
			widget.NewFormItem("Användarnamn:", usernameEntry),
			widget.NewFormItem("Lösenord:", passwordEntry),
		),
		insecureCheck,
		widget.NewLabel(""),
		statusLabel,
		container.NewHBox(
			loginBtn,
			settingsBtn,
		),
	)

	a.window.SetContent(container.NewCenter(form))
}
