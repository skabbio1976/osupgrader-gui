# OSUpgrader GUI - Code Review & TODO

## 📋 CODE REVIEW SAMMANFATTNING

### ✅ **Styrkor / Det som fungerar bra:**

**Arkitektur & Struktur:**
- Välorganiserad projektstruktur med tydlig separation (`cmd/`, `internal/`)
- God användning av Go modules och embedded FS för PowerShell-scripts
- Thread-safe implementering med mutex för parallella uppgraderingar
- Stark separation of concerns (GUI, vCenter, upgrade, config)

**Säkerhet:**
- Bra: Lösenord lagras aldrig i config-filen
- Utmärkt: Credential-validering INNAN uppgradering för att undvika account lockout
- SSPI/Kerberos-stöd för Windows-miljöer
- Säker debug-loggning utan lösenord i klartext

**Funktionalitet:**
- Multi-domän support med automatisk domän-append
- Snapshot-hantering med verifiering
- Robust felhantering med retry-logik
- Worker pool-pattern för parallellisering
- ISO-validering innan snapshot (smart!)

**Användarupplevelse:**
- Fyne GUI med DPI-skalning
- Realtidsloggning med läsbar output
- Progress tracking
- Bra felmeddelanden på svenska

---

## ⚠️ **Förbättringsområden:**

### 1. **Hardkodade värden (KRITISKT för Professional-stöd)**

**Plats:** `internal/upgrade/assets/uppgradeos.ps1:15` och flera andra ställen

**Problem:** Hardkodar "Datacenter" edition

```powershell
Write-Log '=== Windows Server 2022 Datacenter Upgrade Start ==='
```

**Plats:** `internal/upgrade/assets/uppgradeos.ps1:49`

```powershell
$installOption = if ($isCore) { 'ServerDatacenterCore' } else { 'ServerDatacenter' }
```

**Lösning:** Se sektion om Professional-stöd nedan.

---

### 2. **Hardkodad målversion**

**Plats:** `internal/upgrade/upgrade.go:295`

```go
targetOS := []string{"windows server 2022", "windows server 2025"}
```

**Förbättring:** Detta borde vara konfigurerbart via `AppConfig` för att stödja olika målversioner.

---

### 3. **GVLK är endast Datacenter**

**Plats:** `internal/config/config.go:140`

```go
Glvk: "WX4NM-KYWYW-QJJR4-XV3QB-6VM33", // Datacenter GVLK
```

**Problem:** Saknar Standard GVLK-alternativ.

---

### 4. **Timeout-hantering kunde vara bättre**

**Plats:** `internal/upgrade/upgrade.go:205-211`

```go
shutdownTimeout := time.Duration(opts.Config.Timeouts.PowerOffMinutes) * time.Minute
if shutdownTimeout <= 0 {
    shutdownTimeout = 5 * time.Minute
}
```

**Förbättring:** Lägg till validering av timeout-värden vid config load.

---

### 5. **Felhantering i PowerShell-script**

**Plats:** `internal/upgrade/assets/uppgradeos.ps1:30-31`

```powershell
} else {
    exit 1
    # Hantera edge case
}
```

**Förbättring:** Logga vilken InstallationType som hittades innan exit.

---

### 6. **GUI Responsiveness**

**Plats:** `internal/gui/upgrade.go:108-114`

ISO-validering blockerar UI i en goroutine men saknar timeout:

```go
go func() {
    ctx := context.Background()
    if err := upgrade.ValidateISOPath(ctx, isoPath); err != nil {
```

**Förbättring:** Kontexten borde ha timeout:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

---

### 7. **Databasstruktur saknas**

**Problem:** Det finns ingen persistens av uppgraderingshistorik. Om applikationen kraschar förlorar man all historik.

**Förbättring:** Överväg att logga till JSON-fil eller SQLite:
- VM-namn
- Start/slut-tid
- Status (success/failure)
- Felmeddelanden

---

### 8. **PowerShell-script image index logik (KRITISKT)**

**Plats:** `internal/upgrade/assets/uppgradeos.ps1:22-32`

**Problem:** Image index är hardkodad:
```powershell
if ($installationType -eq "Server Core") {
    $installType = 'Core'
    $imageIndex = 3
```

**Risk:** Image index varierar mellan Windows Server-versioner och editioner:
- Windows Server 2022 Datacenter: Index 4 (Desktop), 3 (Core)
- Windows Server 2022 Standard: Index 2 (Desktop), 1 (Core)

Att anta att index 3/4 alltid är rätt kan leda till fel edition being installed!

**Kritisk förbättring:** Använd `DISM /Get-WimInfo` för att dynamiskt hitta rätt index baserat på edition name.

---

### 9. **Ingen rollback-funktion**

**Problem:** Om uppgradering misslyckas finns ingen automatisk rollback till snapshot.

**Förbättring:** Lägg till optional "auto-rollback on failure" i GUI.

---

### 10. **Saknar pre-flight checks**

**Förbättringar att lägga till:**
- Kontrollera Windows Update-status (pending updates kan störa)
- Verifiera att inga reboot pending
- Kolla om Bitlocker är aktiverat (kan orsaka problem)
- Verifiera domain membership innan upgrade

---

### 11. **Språkstöd - Engelska saknas (NY)**

**Problem:** Applikationen är helt på svenska, vilket begränsar användningen i internationella miljöer.

**Påverkade filer:**
- `internal/gui/*.go` - Alla UI-strängar
- `internal/config/config.go` - UI config har `Language` men används inte
- `internal/upgrade/assets/uppgradeos.ps1` - Loggar på svenska
- README.md - Endast svenska

**Förbättring:** Implementera i18n (internationalisering):

**Steg 1:** Använd ett i18n-library som `github.com/nicksnyder/go-i18n/v2`

**Steg 2:** Skapa översättningsfiler:
```
locales/
  ├── sv.json  # Svenska
  └── en.json  # Engelska
```

**Steg 3:** Uppdatera config för att faktiskt använda språkinställningen:
```go
// internal/gui/app.go
func (a *App) tr(key string) string {
    // Lookup translation based on a.config.UI.Language
}
```

**Steg 4:** Ersätt alla hardkodade strängar:
```go
// Före:
statusLabel.SetText("ISO-validering misslyckades")

// Efter:
statusLabel.SetText(a.tr("upgrade.iso_validation_failed"))
```

**Steg 5:** PowerShell-scripts:
- Antingen håll dem på engelska (internationell standard)
- Eller skicka locale som parameter och använd motsvarande meddelanden

**Prioritet:** Medel-Hög (beroende på målgrupp)

---

## 🎯 **Implementera Professional Edition-stöd**

### **Steg 1: Lägg till Edition-konfiguration**

**Fil:** `internal/config/config.go`

Lägg till `TargetEdition` i `DefaultsConfig`:

```go
type DefaultsConfig struct {
    SnapshotNamePrefix   string `json:"snapshot_name_prefix"`
    IsoDatastorePath     string `json:"iso_datastore_path"`
    SkipMemoryInSnapshot bool   `json:"skip_memory_in_snapshot"`
    Glvk                 string `json:"glvk"`
    GuestUsername        string `json:"guest_username,omitempty"`
    TargetEdition        string `json:"target_edition,omitempty"` // NY: "datacenter" eller "standard"
}
```

Uppdatera default i `createDefaultConfig()`:

```go
Defaults: DefaultsConfig{
    SnapshotNamePrefix:   "pre-upgrade",
    IsoDatastorePath:     "[datastore1] iso/windows-server-2022.iso",
    SkipMemoryInSnapshot: true,
    Glvk:                 "WX4NM-KYWYW-QJJR4-XV3QB-6VM33", // Datacenter
    GuestUsername:        "Administrator",
    TargetEdition:        "datacenter", // NY
},
```

---

### **Steg 2: GVLK-mappning**

**Fil:** `internal/config/config.go`

Skapa en helper-funktion för att hämta rätt GVLK:

```go
// GetGVLKForEdition returnerar rätt GVLK baserat på edition
func GetGVLKForEdition(edition string) string {
    gvlks := map[string]string{
        "datacenter": "WX4NM-KYWYW-QJJR4-XV3QB-6VM33",
        "standard":   "VDYBN-27WPP-V4HQT-9VMD4-VMK7H",
    }

    key, exists := gvlks[strings.ToLower(edition)]
    if !exists {
        return gvlks["datacenter"] // Fallback
    }
    return key
}
```

---

### **Steg 3: Uppdatera PowerShell-script**

**Fil:** `internal/upgrade/assets/uppgradeos.ps1`

**Komplett nyskriven version med dynamisk image detection:**

```powershell
param(
    [string]$GLVK,
    [string]$Edition = "Datacenter"  # NY: "Datacenter" eller "Standard"
)

function Write-Log {
    param([string]$Message)
    $ts = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
    "$ts - $Message" | Out-File -Append -FilePath $LogFile
    Write-Output $Message
}

if(!(test-path 'C:\Temp')){
    New-Item -ItemType Directory -Path "C:\Temp" -Force
}
$ErrorActionPreference = 'Stop'
$LogFile = 'C:\Temp\upgrade.log'

try {
    Write-Log "=== Windows Server 2022 $Edition Upgrade Start ==="

    # --- Detect installation type (Core/Desktop) ---
    Write-Log 'Detecting current installation type...'
    $installationType = (Get-ItemProperty -Path "HKLM:\Software\Microsoft\Windows NT\CurrentVersion" -Name "InstallationType").InstallationType

    if (-not $installationType) {
        throw "Could not detect InstallationType from registry"
    }

    $isCore = ($installationType -eq "Server Core")
    $installType = if ($isCore) { 'Core' } else { 'Desktop' }
    Write-Log "Detected installation type: $installType (Registry: $installationType)"

    # --- Locate installation media ---
    Write-Log 'Locating installation media (CD/DVD with install.wim)...'
    $cdDrive = Get-CimInstance Win32_LogicalDisk | Where-Object { $_.DriveType -eq 5 } | Select-Object -First 1
    if (-not $cdDrive) { throw 'Ingen CD-ROM-enhet hittades' }
    $media = $cdDrive.DeviceID
    $wimPath = "$media\sources\install.wim"
    $setupPath = "$media\setup.exe"
    if (-not (Test-Path $wimPath)) { throw "install.wim hittades inte på $media" }
    if (-not (Test-Path $setupPath)) { throw "setup.exe saknas på $media" }
    Write-Log "Media OK: $media"

    # --- Dynamisk image index-detektion ---
    Write-Log "Detecting image index for edition: $Edition ($installType)"

    # Bygg målnamn
    $targetImageName = if ($Edition -eq "Standard") {
        if ($isCore) { "Windows Server 2022 SERVERSTANDARDCORE" } else { "Windows Server 2022 SERVERSTANDARD" }
    } else {
        if ($isCore) { "Windows Server 2022 SERVERDATACENTERCORE" } else { "Windows Server 2022 SERVERDATACENTER" }
    }

    Write-Log "Target image name: $targetImageName"

    # Använd DISM för att hitta rätt index
    $dismInfo = & dism /Get-WimInfo /WimFile:$wimPath | Out-String
    Write-Log "DISM WIM Info retrieved (length: $($dismInfo.Length) chars)"

    # Parsa DISM-output för att hitta index
    $imageIndex = $null
    $lines = $dismInfo -split "`n"
    $currentIndex = $null

    foreach ($line in $lines) {
        if ($line -match '^Index\s*:\s*(\d+)') {
            $currentIndex = $matches[1]
        }
        if ($line -match '^Name\s*:\s*(.+)$') {
            $imageName = $matches[1].Trim()
            Write-Log "Found image at index $currentIndex : $imageName"
            if ($imageName -eq $targetImageName) {
                $imageIndex = $currentIndex
                Write-Log "MATCH! Using image index $imageIndex: $imageName"
                break
            }
        }
    }

    if (-not $imageIndex) {
        Write-Log "ERROR: Could not find image index for $targetImageName"
        Write-Log "DISM output:`n$dismInfo"
        throw "Kunde inte hitta rätt image index för edition $Edition"
    }

    Write-Log "Final image index: $imageIndex"

    # --- Build arguments ---
    Write-Log "Product key: $($GLVK.Substring(0,5))-xxxxx-xxxxx-xxxxx-xxxxx"
    $setupArgs = @('/auto','upgrade','/noreboot','/dynamicupdate','disable','/showoobe','none','/telemetry','disable','/Compat','IgnoreWarning','/eula','accept')
    $setupArgs += @('/imageindex',$imageIndex)
    $setupArgs += @('/pkey',$GLVK)
    Write-Log ("Command: $setupPath " + ($setupArgs -join ' '))

    # --- Start setup and wait for completion ---
    $proc = Start-Process -FilePath $setupPath -ArgumentList $setupArgs -PassThru -WindowStyle Hidden -RedirectStandardOutput 'C:\Windows\Temp\setup_stdout.log' -RedirectStandardError 'C:\Windows\Temp\setup_stderr.log' -Wait
    Write-Log "Setup process completed with exit code: $($proc.ExitCode)"

    if ($proc.ExitCode -ne 0) {
        Write-Log "ERROR: Setup failed with exit code $($proc.ExitCode)"
        if (Test-Path 'C:\Windows\Temp\setup_stderr.log') {
            $stderr = Get-Content 'C:\Windows\Temp\setup_stderr.log' -Raw
            Write-Log "STDERR: $stderr"
        }
        throw "Setup misslyckades med exit code: $($proc.ExitCode)"
    }

    Write-Log 'Setup completed successfully'

    Write-Log 'Scheduling shutdown in 60 seconds...'
    shutdown -s -f -t 60

} catch {
    Write-Log "FATAL ERROR: $($_.Exception.Message)"
    Write-Log "Stack trace: $($_.ScriptStackTrace)"
    throw
}
```

---

### **Steg 4: Uppdatera Go-kod för att skicka edition**

**Fil:** `internal/upgrade/upgrade.go`

**Uppdatera UpgradeOptions struct (rad 30-38):**

```go
type UpgradeOptions struct {
    VMInfo         vcenter.VMInfo
    GuestUsername  string
    GuestPassword  string
    ISOPath        string
    CreateSnapshot bool
    SnapshotName   string
    TargetEdition  string  // NY: "datacenter" eller "standard"
    Config         *config.AppConfig
}
```

**Uppdatera startGuestUpgrade() (rad 408-420):**

```go
// Extrahera och läs in PowerShell-scriptet från embedded FS
scriptTemplate, cleanup, err := extractAndReadPowerShellScript()
if err != nil {
    debug.LogError("ExtractPowerShellScript", err)
    return 0, fmt.Errorf("kunde inte extrahera PowerShell-script: %w", err)
}
defer cleanup()

// Hämta edition från opts eller använd default
edition := opts.TargetEdition
if edition == "" {
    edition = "Datacenter"
}
// Kapitalisera första bokstaven
edition = strings.Title(strings.ToLower(edition))

// Hämta rätt GVLK för edition om ingen specifik är satt
gvlkToUse := gvlk
if gvlk == "" || gvlk == config.GetGVLKForEdition("datacenter") {
    gvlkToUse = config.GetGVLKForEdition(opts.TargetEdition)
}

// Ersätt param-raden med faktiska värden
script := strings.Replace(scriptTemplate, "param([string]$GLVK)",
    fmt.Sprintf("$GLVK = '%s'\n$Edition = '%s'", gvlkToUse, edition), 1)

// Ta även bort den andra param-raden om den finns
script = strings.Replace(script, "param(\n    [string]$GLVK,\n    [string]$Edition = \"Datacenter\"",
    fmt.Sprintf("# Parameters injected\n$GLVK = '%s'\n$Edition = '%s'", gvlkToUse, edition), 1)
```

---

### **Steg 5: Uppdatera GUI för edition-val**

**Fil:** `internal/gui/upgrade.go`

**Lägg till edition dropdown efter rad 54:**

```go
// Edition-val
editionSelect := widget.NewSelect(
    []string{"Datacenter", "Standard"},
    func(selected string) {
        // Uppdatera GVLK automatiskt när edition ändras
        // (valfritt - kan visa användaren vilken GVLK som kommer användas)
    },
)
// Sätt default från config
if a.config.Defaults.TargetEdition != "" {
    editionSelect.SetSelected(strings.Title(a.config.Defaults.TargetEdition))
} else {
    editionSelect.SetSelected("Datacenter")
}
```

**Uppdatera formuläret (rad 286-294):**

```go
form := container.NewVBox(
    infoText,
    widget.NewForm(
        widget.NewFormItem("Target Edition:", editionSelect), // NY
        widget.NewFormItem("Guest admin user:", guestUserEntry),
        widget.NewFormItem("Guest password:", guestPassEntry),
        widget.NewFormItem("ISO datastore path:", isoPathEntry),
    ),
    createSnapshotCheck,
)
```

**Uppdatera UpgradeOptions i worker (rad 174-183):**

```go
// Uppgraderingsalternativ
opts := upgrade.UpgradeOptions{
    VMInfo:         job.vmInfo,
    GuestUsername:  guestUser,
    GuestPassword:  guestPass,
    ISOPath:        isoPath,
    CreateSnapshot: createSnapshotCheck.Checked,
    SnapshotName:   snapshotName,
    TargetEdition:  strings.ToLower(editionSelect.Selected), // NY
    Config:         a.config,
}
```

---

## 📝 **TODO-lista - Prioriterad**

### 🔴 **Hög Prioritet**

- [ ] **Implementera Professional Edition-stöd**
  - [ ] Lägg till `TargetEdition` i config (`internal/config/config.go`)
  - [ ] Skapa `GetGVLKForEdition()` helper-funktion
  - [ ] Uppdatera `uppgradeos.ps1` med dynamisk DISM image detection
  - [ ] Lägg till `TargetEdition` i `UpgradeOptions` struct
  - [ ] Uppdatera `startGuestUpgrade()` för att skicka edition till script
  - [ ] Lägg till edition dropdown i GUI (`internal/gui/upgrade.go`)
  - [ ] Testa med både Standard och Datacenter ISOs

- [ ] **Fixa kritisk image index-logik**
  - [ ] Implementera dynamisk DISM-baserad image detection istället för hardkodade index
  - [ ] Testa med olika Windows Server 2022 ISO-versioner
  - [ ] Lägg till fallback-logik om DISM parsing misslyckas

### 🟡 **Medel Prioritet**

- [ ] **Implementera engelska språkstöd**
  - [ ] Integrera `github.com/nicksnyder/go-i18n/v2` library
  - [ ] Skapa `locales/sv.json` och `locales/en.json`
  - [ ] Extrahera alla UI-strängar till translation keys
  - [ ] Implementera `tr()` helper-funktion i GUI
  - [ ] Uppdatera alla GUI-filer att använda översättningar
  - [ ] Lägg till språkväljare i Settings dialog
  - [ ] Uppdatera PowerShell-scripts att använda engelska loggar (eller parametriserad locale)
  - [ ] Översätt README.md till engelska (eller skapa README.en.md)

- [ ] **Lägg till uppgraderingshistorik**
  - [ ] Implementera SQLite eller JSON-baserad historikloggning
  - [ ] Skapa datamodell för upgrade history
  - [ ] Visa historik i GUI (ny tab/skärm)
  - [ ] Möjlighet att exportera historik till CSV/JSON

- [ ] **Förbättra pre-flight checks**
  - [ ] Kontrollera Windows Update-status
  - [ ] Verifiera pending reboot
  - [ ] Kolla Bitlocker-status
  - [ ] Verifiera domain membership

- [ ] **Timeout-förbättringar**
  - [ ] Lägg till validering av timeout-värden vid config load
  - [ ] Lägg till timeout för ISO-validering i GUI
  - [ ] Förbättra timeout-felmeddelanden

### 🟢 **Låg Prioritet**

- [ ] **Rollback-funktion**
  - [ ] Lägg till "auto-rollback on failure" option i GUI
  - [ ] Implementera logik för att detektera misslyckad upgrade
  - [ ] Automatisk revert till snapshot vid fel

- [ ] **Performance-optimeringar**
  - [ ] Connection pooling för vCenter-klienter
  - [ ] Batch snapshot operations
  - [ ] Cacha VM inventory

- [ ] **Förbättra felhantering**
  - [ ] Logga InstallationType vid PowerShell script exit
  - [ ] Bättre felmeddelanden i alla Go-filer
  - [ ] Retry-logik för transienta fel

- [ ] **Säkerhetsförbättringar**
  - [ ] Validera GVLK-format (regex: `^[A-Z0-9]{5}(-[A-Z0-9]{5}){4}$`)
  - [ ] Rate limiting för credential validation
  - [ ] Audit logging till centraliserad plats
  - [ ] Role-based access control

---

## 🔒 **Säkerhetsrekommendationer**

1. **Validera GVLK-format:** Kontrollera att GVLK matchar `XXXXX-XXXXX-XXXXX-XXXXX-XXXXX` pattern
2. **Rate limiting:** Lägg till delay mellan credential validation-försök
3. **Audit logging:** Logga alla uppgraderingar till centraliserad plats
4. **Role-based access:** Överväg att kräva speciell roll för bulk-uppgraderingar

---

## ⚡ **Performance-optimeringar**

1. **Connection pooling:** Återanvänd vCenter-klient för flera VMs
2. **Batch snapshot operations:** Om möjligt, gruppera snapshot-skapande
3. **Caching:** Cacha VM inventory mellan omgångar

---

## 📚 **Referensinformation**

### Windows Server 2022 GVLK Keys

| Edition | GVLK |
|---------|------|
| Datacenter | `WX4NM-KYWYW-QJJR4-XV3QB-6VM33` |
| Standard | `VDYBN-27WPP-V4HQT-9VMD4-VMK7H` |

### Typiska WIM Image Index

| Edition | Desktop | Core |
|---------|---------|------|
| Standard | 2 | 1 |
| Datacenter | 4 | 3 |

**OBS:** Dessa kan variera mellan ISO-versioner - använd alltid DISM för att verifiera!

---

## 📞 **Support & Dokumentation**

- VMware govmomi docs: https://github.com/vmware/govmomi
- Fyne GUI docs: https://developer.fyne.io/
- Windows Server upgrade docs: https://docs.microsoft.com/en-us/windows-server/upgrade/

---

**Skapad:** 2025-11-18
**Version:** 0.2.0
**Status:** Under utveckling
