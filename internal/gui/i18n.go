package gui

// Translations contains all translatable strings
type Translations struct {
	// Login screen
	AppTitle                string
	LoginTitle              string
	VCenterHost             string
	Username                string
	Password                string
	AuthMethod              string
	AuthPassword            string
	AuthSSPI                string
	AllowUnsignedCerts      string
	LoginButton             string
	SettingsButton          string
	ConnectingStatus        string
	ConnectingSSPI          string
	LoadingVMs              string
	LoginSuccess            string
	LoginError              string
	ErrorHostRequired       string
	ErrorCredsRequired      string
	ErrorSSPIFailed         string
	ErrorLoginFailed        string
	ErrorLoadVMsFailed      string

	// VM Selection screen
	VMSelectionTitle        string
	VMSelectionTitleCount   string // With count: "Select VMs to Upgrade (%d VMs available)"
	SearchPlaceholder       string
	SelectAll               string
	DeselectAll             string
	RefreshList             string
	LogOut                  string
	ManageSnapshots         string
	ContinueToUpgrade       string
	NoVMsSelected           string
	SelectVMsFirst          string
	ColumnSelect            string
	ColumnName              string
	ColumnFolder            string
	ColumnDomain            string
	ColumnOS                string
	Refreshing              string
	RefreshingMessage       string
	ErrorRefreshVMs         string

	// Upgrade screen
	UpgradeTitle            string
	UpgradeVMs              string // "Upgrade %d VMs to Windows Server 2022"
	GuestAdminUser          string
	GuestAdminPassword      string
	ISODatastorePath        string
	CreateSnapshot          string
	SnapshotNamePrefix      string
	StartUpgrade            string
	Back                    string
	FillAllFields           string
	UpgradeInProgress       string
	UpgradeComplete         string
	UpgradeErrors           string
	ShowPassword            string
	TipSaveCredentials      string
	ReadyToStart            string // "Ready to start - %d servers selected"
	FillInDetails           string
	ValidatingISO           string
	ISOValidationFailed     string
	ISOOK                   string
	ISOValidated            string
	StartingUpgrade         string // "Starting upgrade of %d servers..."
	UpgradeCompleted        string // "✓ DONE (%s) - Upgrade completed!"
	AllSuccessful           string
	SomeFailed              string

	// Snapshot management screen
	SnapshotsTitle          string
	RefreshButton           string
	RemoveSelected          string
	CloseButton             string
	SnapshotColumnVM        string
	SnapshotColumnName      string
	SnapshotColumnCreated   string
	NoSnapshots             string
	SnapshotsLoading        string
	RemoveConfirmTitle      string
	RemoveConfirmMessage    string // "Are you sure you want to remove %d snapshot(s)?\n\nThis CANNOT be undone!"
	NoSnapshotsSelected     string
	SelectSnapshotsFirst    string
	RemovingSnapshots       string // "Removing %d snapshot(s) in parallel..."
	RemoveSuccessCount      string // "✓ Done! %d snapshot(s) removed"
	RemoveErrorCount        string // "⚠ Done with errors: %d removed, %d failed"
	RemoveSuccess           string
	RemoveError             string

	// Settings dialog
	SettingsTitle           string
	TabGuestISO             string
	TabUpgrade              string
	TabTimeouts             string
	TabUI                   string
	SaveButton              string
	GuestUsername           string
	GuestPassword           string
	SnapshotPrefix          string
	ISOPath                 string
	ParallelUpgrades        string
	ParallelUpgradesInfo    string
	TimeoutMinutes          string
	DiskPrecheckGB          string
	SkipMemoryInSnapshot    string
	RebootAfterUpgrade      string
	SignalScriptSeconds     string
	SignalFilesMinutes      string
	OSVersionPollingMinutes string
	PowerOffTimeoutMinutes  string
	TimeoutsDescription     string
	DarkMode                string
	Language                string
	LanguageEnglish         string
	LanguageSwedish         string
	SettingsSaved           string
	SettingsSavedMessage    string
	SettingsError           string
}

var englishTranslations = Translations{
	// Login screen
	AppTitle:                "OSUpgrader - Windows Server Upgrade Tool",
	LoginTitle:              "OSUpgrader - vCenter Login",
	VCenterHost:             "vCenter Host",
	Username:                "Username",
	Password:                "Password",
	AuthMethod:              "Authentication Method",
	AuthPassword:            "Password",
	AuthSSPI:                "Windows SSPI/Kerberos",
	AllowUnsignedCerts:      "Allow unsigned certificates",
	LoginButton:             "Log in",
	SettingsButton:          "Settings",
	ConnectingStatus:        "Logging in...",
	ConnectingSSPI:          "Logging in with SSPI/Kerberos...",
	LoadingVMs:              "Logged in! Loading VMs...",
	LoginSuccess:            "Successfully connected to vCenter",
	LoginError:              "Login Error",
	ErrorHostRequired:       "vCenter host must be specified",
	ErrorCredsRequired:      "username and password must be filled in",
	ErrorSSPIFailed:         "SSPI login failed: %v - NOTE: SSPI/Kerberos only works on Windows and requires you to be logged in with a domain account",
	ErrorLoginFailed:        "login failed: %v",
	ErrorLoadVMsFailed:      "could not load VMs: %v",

	// VM Selection screen
	VMSelectionTitle:        "Select VMs to Upgrade",
	VMSelectionTitleCount:   "Select VMs to Upgrade (%d VMs available)",
	SearchPlaceholder:       "Search or regex (e.g. '003|009|web')...",
	SelectAll:               "Select all",
	DeselectAll:             "Deselect all",
	RefreshList:             "Refresh list",
	LogOut:                  "Log out",
	ManageSnapshots:         "Manage snapshots",
	ContinueToUpgrade:       "Continue to upgrade",
	NoVMsSelected:           "No VMs selected",
	SelectVMsFirst:          "You must select at least one VM to upgrade.",
	ColumnSelect:            "Select",
	ColumnName:              "Name",
	ColumnFolder:            "Folder",
	ColumnDomain:            "Domain",
	ColumnOS:                "OS",
	Refreshing:              "Refreshing...",
	RefreshingMessage:       "Fetching VM list from vCenter...",
	ErrorRefreshVMs:         "could not fetch VMs: %v",

	// Upgrade screen
	UpgradeTitle:            "Configure Upgrade",
	UpgradeVMs:              "Upgrade %d VMs to Windows Server 2022",
	GuestAdminUser:          "Guest admin user",
	GuestAdminPassword:      "Guest admin password",
	ISODatastorePath:        "ISO datastore path",
	CreateSnapshot:          "Create snapshot before upgrade",
	SnapshotNamePrefix:      "Snapshot name prefix",
	StartUpgrade:            "Start upgrade",
	Back:                    "Back",
	FillAllFields:           "Please fill in all required fields.",
	UpgradeInProgress:       "Upgrade in progress...",
	UpgradeComplete:         "Upgrade Complete",
	UpgradeErrors:           "Upgrade completed with errors",
	ShowPassword:            "Show password",
	TipSaveCredentials:      "💡 Tip: Save guest credentials in Settings to avoid entering them every time!",
	ReadyToStart:            "Ready to start - %d servers selected",
	FillInDetails:           "Fill in the details below and click 'Start upgrade' to begin.\n\n",
	ValidatingISO:           "Validating ISO path...",
	ISOValidationFailed:     "ISO validation failed",
	ISOOK:                   "ISO OK. Starting upgrades...",
	ISOValidated:            "ISO validated",
	StartingUpgrade:         "Starting upgrade of %d servers...\n\n",
	UpgradeCompleted:        "✓ DONE (%s) - Upgrade completed!",
	AllSuccessful:           "Status: All upgrades completed successfully!",
	SomeFailed:              "Status: Some upgrades failed, see log above for details",

	// Snapshot management screen
	SnapshotsTitle:          "Manage Snapshots",
	RefreshButton:           "Refresh",
	RemoveSelected:          "Remove selected",
	CloseButton:             "Close",
	SnapshotColumnVM:        "VM",
	SnapshotColumnName:      "Snapshot Name",
	SnapshotColumnCreated:   "Created",
	NoSnapshots:             "No pre-upgrade snapshots found",
	SnapshotsLoading:        "Loading snapshots...",
	RemoveConfirmTitle:      "Confirm Removal",
	RemoveConfirmMessage:    "Are you sure you want to remove %d snapshot(s)?\n\nThis CANNOT be undone!",
	NoSnapshotsSelected:     "No snapshots selected",
	SelectSnapshotsFirst:    "You must select at least one snapshot to remove.",
	RemovingSnapshots:       "Removing %d snapshot(s) in parallel...",
	RemoveSuccessCount:      "✓ Done! %d snapshot(s) removed",
	RemoveErrorCount:        "⚠ Done with errors: %d removed, %d failed",
	RemoveSuccess:           "Snapshots Removed",
	RemoveError:             "Error removing snapshots",

	// Settings dialog
	SettingsTitle:           "Settings",
	TabGuestISO:             "Guest & ISO",
	TabUpgrade:              "Upgrade",
	TabTimeouts:             "Timeouts",
	TabUI:                   "UI",
	SaveButton:              "Save",
	GuestUsername:           "Guest admin username",
	GuestPassword:           "Guest admin password",
	SnapshotPrefix:          "Snapshot prefix",
	ISOPath:                 "ISO datastore path",
	ParallelUpgrades:        "Parallel upgrades",
	ParallelUpgradesInfo:    "Number of VMs upgraded simultaneously (higher value = faster for many VMs)",
	TimeoutMinutes:          "Timeout (minutes)",
	DiskPrecheckGB:          "Disk precheck (GB)",
	SkipMemoryInSnapshot:    "Skip memory in snapshot",
	RebootAfterUpgrade:      "Reboot after upgrade",
	SignalScriptSeconds:     "Signal script (seconds)",
	SignalFilesMinutes:      "Signal files (minutes)",
	OSVersionPollingMinutes: "OS version polling (minutes)",
	PowerOffTimeoutMinutes:  "Power-off timeout (minutes)",
	TimeoutsDescription:     "All timeout values can be adjusted below",
	DarkMode:                "Dark mode",
	Language:                "Language",
	LanguageEnglish:         "English",
	LanguageSwedish:         "Svenska (Swedish)",
	SettingsSaved:           "Settings Saved",
	SettingsSavedMessage:    "Configuration has been saved (guest password is only stored in memory)",
	SettingsError:           "Error saving settings",
}

var swedishTranslations = Translations{
	// Login screen
	AppTitle:                "OSUpgrader - Windows Server Upgrade Tool",
	LoginTitle:              "OSUpgrader - vCenter Inloggning",
	VCenterHost:             "vCenter Host",
	Username:                "Användarnamn",
	Password:                "Lösenord",
	AuthMethod:              "Autentiseringsmetod",
	AuthPassword:            "Lösenord",
	AuthSSPI:                "Windows SSPI/Kerberos",
	AllowUnsignedCerts:      "Tillåt osignerade certifikat",
	LoginButton:             "Logga in",
	SettingsButton:          "Inställningar",
	ConnectingStatus:        "Loggar in...",
	ConnectingSSPI:          "Loggar in med SSPI/Kerberos...",
	LoadingVMs:              "Inloggad! Hämtar VMs...",
	LoginSuccess:            "Ansluten till vCenter",
	LoginError:              "Inloggningsfel",
	ErrorHostRequired:       "vCenter host måste anges",
	ErrorCredsRequired:      "användarnamn och lösenord måste fyllas i",
	ErrorSSPIFailed:         "SSPI-inloggning misslyckades: %v - OBS: SSPI/Kerberos fungerar endast på Windows och kräver att du är inloggad med ett domänkonto",
	ErrorLoginFailed:        "inloggning misslyckades: %v",
	ErrorLoadVMsFailed:      "kunde inte hämta VMs: %v",

	// VM Selection screen
	VMSelectionTitle:        "Välj VMs att uppgradera",
	VMSelectionTitleCount:   "Välj VMs att uppgradera (%d VMs tillgängliga)",
	SearchPlaceholder:       "Sök eller regex (t.ex. '003|009|web')...",
	SelectAll:               "Välj alla",
	DeselectAll:             "Avmarkera alla",
	RefreshList:             "Uppdatera lista",
	LogOut:                  "Logga ut",
	ManageSnapshots:         "Hantera snapshots",
	ContinueToUpgrade:       "Fortsätt till uppgradering",
	NoVMsSelected:           "Inga VMs valda",
	SelectVMsFirst:          "Du måste välja minst en VM att uppgradera.",
	ColumnSelect:            "Välj",
	ColumnName:              "Name",
	ColumnFolder:            "Folder",
	ColumnDomain:            "Domain",
	ColumnOS:                "OS",
	Refreshing:              "Uppdaterar...",
	RefreshingMessage:       "Hämtar VM-lista från vCenter...",
	ErrorRefreshVMs:         "kunde inte hämta VMs: %v",

	// Upgrade screen
	UpgradeTitle:            "Konfigurera uppgradering",
	UpgradeVMs:              "Uppgradera %d VMs till Windows Server 2022",
	GuestAdminUser:          "Guest admin user",
	GuestAdminPassword:      "Guest admin lösenord",
	ISODatastorePath:        "ISO datastore path",
	CreateSnapshot:          "Skapa snapshot före uppgradering",
	SnapshotNamePrefix:      "Snapshot-prefix",
	StartUpgrade:            "Starta uppgradering",
	Back:                    "Tillbaka",
	FillAllFields:           "Fyll i alla obligatoriska fält.",
	UpgradeInProgress:       "Uppgradering pågår...",
	UpgradeComplete:         "Uppgradering klar",
	UpgradeErrors:           "Uppgradering slutfördes med fel",
	ShowPassword:            "Visa lösenord",
	TipSaveCredentials:      "💡 Tips: Spara guest credentials i Inställningar för att slippa ange dem varje gång!",
	ReadyToStart:            "Redo att starta - %d servrar valda",
	FillInDetails:           "Fyll i uppgifterna nedan och klicka 'Starta uppgradering' för att börja.\n\n",
	ValidatingISO:           "Validerar ISO-path...",
	ISOValidationFailed:     "ISO-validering misslyckades",
	ISOOK:                   "ISO OK. Startar uppgraderingar...",
	ISOValidated:            "ISO validerad",
	StartingUpgrade:         "Startar uppgradering av %d servrar...\n\n",
	UpgradeCompleted:        "✓ KLAR (%s) - Uppgradering slutförd!",
	AllSuccessful:           "Status: Alla uppgraderingar slutförda utan fel!",
	SomeFailed:              "Status: Vissa uppgraderingar misslyckades, se logg ovan för detaljer",

	// Snapshot management screen
	SnapshotsTitle:          "Hantera snapshots",
	RefreshButton:           "Uppdatera",
	RemoveSelected:          "Ta bort valda",
	CloseButton:             "Stäng",
	SnapshotColumnVM:        "VM",
	SnapshotColumnName:      "Snapshot-namn",
	SnapshotColumnCreated:   "Skapad",
	NoSnapshots:             "Inga pre-upgrade snapshots hittades",
	SnapshotsLoading:        "Laddar snapshots...",
	RemoveConfirmTitle:      "Bekräfta borttagning",
	RemoveConfirmMessage:    "Är du säker på att du vill ta bort %d snapshot(s)?\n\nDetta kan INTE ångras!",
	NoSnapshotsSelected:     "Inga snapshots valda",
	SelectSnapshotsFirst:    "Du måste välja minst en snapshot att ta bort.",
	RemovingSnapshots:       "Tar bort %d snapshot(s) parallellt...",
	RemoveSuccessCount:      "✓ Klart! %d snapshot(s) borttagna",
	RemoveErrorCount:        "⚠ Klart med fel: %d borttagna, %d misslyckades",
	RemoveSuccess:           "Snapshots borttagna",
	RemoveError:             "Fel vid borttagning av snapshots",

	// Settings dialog
	SettingsTitle:           "Inställningar",
	TabGuestISO:             "Guest & ISO",
	TabUpgrade:              "Uppgradering",
	TabTimeouts:             "Timeouts",
	TabUI:                   "Användargränssnitt",
	SaveButton:              "Spara",
	GuestUsername:           "Guest admin användare",
	GuestPassword:           "Guest admin lösenord",
	SnapshotPrefix:          "Snapshot-prefix",
	ISOPath:                 "ISO datastore path",
	ParallelUpgrades:        "Parallella uppgraderingar",
	ParallelUpgradesInfo:    "Antal VMs som uppgraderas samtidigt (högt värde = snabbare för många VMs)",
	TimeoutMinutes:          "Timeout (minuter)",
	DiskPrecheckGB:          "Disk precheck (GB)",
	SkipMemoryInSnapshot:    "Hoppa över minne i snapshot",
	RebootAfterUpgrade:      "Starta om efter uppgradering",
	SignalScriptSeconds:     "Signal script (sekunder)",
	SignalFilesMinutes:      "Signal filer (minuter)",
	OSVersionPollingMinutes: "OS-version polling (minuter)",
	PowerOffTimeoutMinutes:  "Power-off timeout (minuter)",
	TimeoutsDescription:     "Alla timeout-värden kan justeras nedan",
	DarkMode:                "Dark mode",
	Language:                "Språk",
	LanguageEnglish:         "English (Engelska)",
	LanguageSwedish:         "Svenska",
	SettingsSaved:           "Inställningar sparade",
	SettingsSavedMessage:    "Konfigurationen har sparats (guest password sparas endast i minnet)",
	SettingsError:           "Fel vid sparande av inställningar",
}

// GetTranslations returns the appropriate translations based on language code
func GetTranslations(lang string) Translations {
	switch lang {
	case "sv", "swedish":
		return swedishTranslations
	case "en", "english":
		return englishTranslations
	default:
		return englishTranslations // Default to English
	}
}
