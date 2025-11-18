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
	passwordEntry.SetPlaceHolder(a.tr.Password)

	insecureCheck := widget.NewCheck(a.tr.AllowUnsignedCerts, nil)
	insecureCheck.SetChecked(a.config.VCenter.Insecure)

	// Autentiseringsmetod
	authMethodSelect := widget.NewRadioGroup([]string{a.tr.AuthPassword, a.tr.AuthSSPI}, func(selected string) {
		if selected == a.tr.AuthSSPI {
			usernameEntry.Disable()
			passwordEntry.Disable()
		} else {
			usernameEntry.Enable()
			passwordEntry.Enable()
		}
	})
	authMethodSelect.SetSelected(a.tr.AuthPassword)
	if a.config.VCenter.Mode == "sspi" {
		authMethodSelect.SetSelected(a.tr.AuthSSPI)
	}

	statusLabel := widget.NewLabel("")

	// Login-knapp (deklarera först)
	var loginBtn *widget.Button
	loginBtn = widget.NewButton(a.tr.LoginButton, func() {
		host := hostEntry.Text
		authMethod := authMethodSelect.Selected

		if host == "" {
			dialog.ShowError(fmt.Errorf("%s", a.tr.ErrorHostRequired), a.window)
			return
		}

		// Uppdatera config
		a.config.VCenter.Host = host
		a.config.VCenter.Insecure = insecureCheck.Checked

		var client *vcenter.Client
		var err error

		if authMethod == a.tr.AuthSSPI {
			// SSPI-inloggning
			if usernameEntry.Text != "" {
				a.config.VCenter.Username = usernameEntry.Text
			}
			a.config.VCenter.Mode = "sspi"

			statusLabel.SetText(a.tr.ConnectingSSPI)
			loginBtn.Disable()

			// Logga in i bakgrunden
			go func() {
				client, err = vcenter.LoginSSPI(&a.config.VCenter)
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf(a.tr.ErrorSSPIFailed, err), a.window)
					return
				}

				a.SetClient(client)
				statusLabel.SetText(a.tr.LoadingVMs)

				// Hämta VMs
				vms, err := vcenter.GetVMInfos()
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf(a.tr.ErrorLoadVMsFailed, err), a.window)
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
				dialog.ShowError(fmt.Errorf("%s", a.tr.ErrorCredsRequired), a.window)
				return
			}

			a.config.VCenter.Username = username
			a.config.VCenter.Mode = "password"

			statusLabel.SetText(a.tr.ConnectingStatus)
			loginBtn.Disable()

			// Logga in i bakgrunden
			go func() {
				client, err = vcenter.Login(&a.config.VCenter, password)
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf(a.tr.ErrorLoginFailed, err), a.window)
					return
				}

				a.SetClient(client)
				statusLabel.SetText(a.tr.LoadingVMs)

				// Hämta VMs
				vms, err := vcenter.GetVMInfos()
				if err != nil {
					statusLabel.SetText("")
					loginBtn.Enable()
					dialog.ShowError(fmt.Errorf(a.tr.ErrorLoadVMsFailed, err), a.window)
					return
				}

				a.SetVMs(vms)

				// Byt till VM-selection-skärm
				a.showVMSelectionScreen()
			}()
		}
	})

	// Inställningar-knapp
	settingsBtn := widget.NewButton(a.tr.SettingsButton, func() {
		a.showSettingsDialog()
	})

	// Layout
	form := container.NewVBox(
		widget.NewLabelWithStyle(a.tr.AppTitle, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
		widget.NewForm(
			widget.NewFormItem(a.tr.VCenterHost+":", hostEntry),
		),
		widget.NewLabel(a.tr.AuthMethod+":"),
		authMethodSelect,
		widget.NewLabel(""),
		widget.NewForm(
			widget.NewFormItem(a.tr.Username+":", usernameEntry),
			widget.NewFormItem(a.tr.Password+":", passwordEntry),
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
