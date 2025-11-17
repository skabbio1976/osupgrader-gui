package gui

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/yourusername/osupgrader-gui/internal/config"
	"github.com/yourusername/osupgrader-gui/internal/debug"
	"github.com/yourusername/osupgrader-gui/internal/vcenter"
)

// Custom themes för att undvika deprecated warnings
type darkTheme struct{}

func (d darkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}

func (d darkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (d darkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (d darkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type lightTheme struct{}

func (l lightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

func (l lightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (l lightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (l lightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// App representerar huvudapplikationen
type App struct {
	fyneApp       fyne.App
	window        fyne.Window
	config        *config.AppConfig
	client        *vcenter.Client
	vms           []vcenter.VMInfo
	guestPassword string // Hålls i minnet, sparas ej
	mockMode      bool   // Mock mode för testing
}

// NewApp skapar en ny GUI-applikation
func NewApp(debugMode bool, mockMode bool) *App {
	// Initialisera debug-loggning om debug mode är aktiverad
	if debugMode {
		if err := debug.Init(); err != nil {
			log.Printf("VARNING: Kunde inte initialisera debug-loggning: %v", err)
		} else {
			debug.Log("OSUpgrader GUI Started")
			debug.Log("Debug log path: %s", debug.GetLogPath())
			if mockMode {
				debug.Log("Mock mode ENABLED - using fake VMs")
			}
		}
	}

	// Applicera automatisk DPI-skalning INNAN app.New()
	// Använder registry på Windows för att undvika PowerShell-fönster
	applyDPIScale()

	a := &App{
		fyneApp:  app.NewWithID("com.example.osupgrader"),
		mockMode: mockMode,
	}

	a.window = a.fyneApp.NewWindow("OSUpgrader - Windows Server Upgrade Tool")
	a.window.Resize(fyne.NewSize(1000, 800))

	// Ladda konfiguration
	cfg, err := config.Load()
	if err != nil {
		debug.LogError("ConfigLoad", err)
		// Använd defaultkonfiguration
		cfg = &config.AppConfig{}
	} else {
		debug.Log("Configuration loaded successfully")
	}
	a.config = cfg

	// Applicera tema från config
	if a.config.UI.DarkMode {
		a.fyneApp.Settings().SetTheme(&darkTheme{})
		debug.Log("Dark mode enabled")
	} else {
		a.fyneApp.Settings().SetTheme(&lightTheme{})
		debug.Log("Light mode enabled")
	}

	return a
}

// Run startar applikationen
func (a *App) Run() {
	// Stäng debug-loggning när appen avslutas
	defer debug.Close()

	// Logga när fönstret stängs
	a.window.SetOnClosed(func() {
		debug.Log("Application window closed")
	})

	// Om mock mode, generera fake VMs och gå direkt till VM selection
	if a.mockMode {
		a.generateMockVMs()
		a.showVMSelectionScreen()
	} else {
		a.showLoginScreen()
	}

	a.window.ShowAndRun()
}

// GetConfig returnerar applikationens konfiguration
func (a *App) GetConfig() *config.AppConfig {
	return a.config
}

// SetClient sätter vCenter-klienten
func (a *App) SetClient(client *vcenter.Client) {
	a.client = client
}

// GetClient returnerar vCenter-klienten
func (a *App) GetClient() *vcenter.Client {
	return a.client
}

// SetVMs sätter VM-listan
func (a *App) SetVMs(vms []vcenter.VMInfo) {
	a.vms = vms
}

// GetVMs returnerar VM-listan
func (a *App) GetVMs() []vcenter.VMInfo {
	return a.vms
}

// GetWindow returnerar huvudfönstret
func (a *App) GetWindow() fyne.Window {
	return a.window
}

// generateMockVMs genererar fake VMs för testing
func (a *App) generateMockVMs() {
	debug.Log("Generating mock VMs (srv001-srv100)...")

	folders := []string{"Production/WebServers", "Production/DatabaseServers", "Development/TestServers", "Staging/AppServers"}
	domains := []string{
		"verylong.production.subdomain.corporate.infrastructure.example.com",
		"extremelylong.development.testing.integration.environment.example.com",
		"superlongname.staging.application.deployment.infrastructure.example.com",
		"incrediblylong.corporate.business.application.management.example.com",
	}
	oses := []string{
		"Microsoft Windows Server 2016 (64-bit)",
		"Microsoft Windows Server 2019 (64-bit)",
		"Microsoft Windows Server 2012 R2 (64-bit)",
	}

	mockVMs := make([]vcenter.VMInfo, 0, 100)

	for i := 1; i <= 100; i++ {
		vmName := fmt.Sprintf("srv%03d", i)
		folder := folders[i%len(folders)]
		domain := domains[i%len(domains)]
		os := oses[i%len(oses)]

		// Skapa FQDN från vmName och domain
		fqdn := fmt.Sprintf("%s.%s", vmName, domain)

		mockVM := vcenter.VMInfo{
			Name:   vmName,
			Folder: folder,
			Domain: fqdn,
			OS:     os,
			Ref:    types.ManagedObjectReference{Type: "VirtualMachine", Value: fmt.Sprintf("vm-%d", i)},
		}

		mockVMs = append(mockVMs, mockVM)

		// Logga varje mock VM
		debug.Log("Mock VM created: Name=%s, Folder=%s, Domain=%s, OS=%s", vmName, folder, fqdn, os)
	}

	a.vms = mockVMs
	debug.Log("Mock VM generation complete: %d VMs created", len(mockVMs))
}
