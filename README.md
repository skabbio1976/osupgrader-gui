# OSUpgrader GUI

En grafisk version av OSUpgrader - ett automatiserat verktyg för att uppgradera Windows Server 2016/2019 till Windows Server 2022 Datacenter via VMware vCenter/vSphere.

## Översikt

OSUpgrader GUI är en Fyne-baserad grafisk applikation som gör det enkelt att uppgradera flera Windows Server-VMs samtidigt via en intuitiv användargränssnitt.

## Funktioner

- **Grafiskt gränssnitt** med Fyne-ramverket
  - Automatisk DPI-skalning för optimal visning på alla skärmar
- **vCenter-inloggning** med stöd för:
  - Lösenordsautentisering (alla plattformar)
  - Windows SSPI/Kerberos single sign-on (Windows endast) ✓ Testad och verifierad
  - Osignerade certifikat
- **VM-selection** med tabell-vy (Name, Folder, Domain, OS kolumner), sökfiltrering och multi-select
- **Multi-domän support**:
  - Automatisk domän-append till användarnamn (t.ex. `upgrade` → `upgrade@domain.local`)
  - Möjliggör samma servicekonto i flera domäner
  - Visar VM:ens domän i tabellvy
- **Snapshot-hantering**:
  - Automatisk snapshot före uppgradering
  - Separat skärm för att hantera och ta bort pre-upgrade snapshots
  - Batch-borttagning av flera snapshots samtidigt
  - Filtrering på snapshot-prefix
- **Parallella uppgraderingar** med konfigurerbar samtidighet
- **Progress tracking** med real-time loggning och readable text
- **ISO-validering** innan uppgradering startar
- **Konfigurationshantering** via GUI-dialog med sparade guest-credentials
- **Debug-loggning** (valfritt med `-d/--debug` flagga):
  - Detaljerad loggning till `debuglogg.txt`
  - Säker loggning (lösenord aldrig i klartext)
  - Perfekt för troubleshooting i airgapped miljöer
- **Säker autentisering**:
  - Credential-validering innan uppgradering
  - Förhindrar account lockout från misslyckade försök

## Systemkrav

- Go 1.24 eller senare
- Linux/Windows/macOS
- Tillgång till VMware vCenter
- Windows Server 2022 Datacenter ISO på datastore

## Installation

### Bygg från källkod

```bash
# Klona projektet
cd /home/jok/gitrepos/goprojects/active/osupgrader-gui

# Hämta dependencies
go mod tidy

# Bygg applikationen
go build -o osupgrader-gui ./cmd/osupgrader-gui

# Kör applikationen
./osupgrader-gui
```

### Bygg för både Linux och Windows

```bash
# Använd build-scriptet (kräver Docker för Windows-bygge)
./build.sh
```

### Manuell Windows-kompilering

```bash
# Installera fyne-cross först
go install github.com/fyne-io/fyne-cross@latest

# Bygg för Windows från Linux (kräver Docker)
~/go/bin/fyne-cross windows -arch=amd64 -app-id com.example.osupgrader ./cmd/osupgrader-gui

# Extrahera
cd fyne-cross/dist/windows-amd64
unzip osupgrader-gui.exe.zip
```

## Användning

1. **Starta applikationen**
   ```bash
   # Normal användning (ingen debug-loggning)
   ./osupgrader-gui

   # Med debug-loggning (för troubleshooting)
   ./osupgrader-gui -d
   # eller
   ./osupgrader-gui --debug
   ```

   När debug-loggning är aktiverad skapas `debuglogg.txt` i samma mapp som programmet med detaljerad information om alla operationer.

2. **Logga in på vCenter**
   - Ange vCenter-host (t.ex. `vcenter.example.local`)
   - Välj autentiseringsmetod:
     - **Lösenord**: Ange användarnamn och lösenord
     - **Windows SSPI/Kerberos**: Single sign-on med ditt Windows-domänkonto (endast Windows)
       - Ingen lösenordsinmatning krävs
       - Använder dina Windows-credentials automatiskt
       - Perfekt för domän-miljöer med integrerad autentisering
   - Markera "Tillåt osignerade certifikat" om nödvändigt
   - Klicka på "Logga in"

3. **Välj VMs att uppgradera**
   - Tabell-vy visar alla VMs med kolumner: Välj, Name, Folder, Domain, OS
   - Sök efter VMs med sökfältet (söker i alla kolumner inklusive domän)
   - Välj VMs genom att markera checkboxarna i första kolumnen
   - Använd "Välj alla" / "Avmarkera alla" för bulkoperationer
   - Klicka på "Hantera snapshots" för att ta bort gamla pre-upgrade snapshots
   - Klicka på "Fortsätt till uppgradering"

4. **Konfigurera uppgradering**
   - Ange guest admin-användare (t.ex. `upgrade`)
     - **Multi-domän support**: Ange bara användarnamn utan domän (t.ex. `upgrade`)
     - Systemet lägger automatiskt till VM:ens domän: `upgrade@domain1.local`
     - Fungerar perfekt med samma servicekonto i flera domäner
     - Om du vill ange specifik domän, använd `DOMAIN\user` eller `user@domain.com`
   - Ange guest-lösenord
   - **💡 Tips**: Spara guest credentials i Inställningar för att slippa ange dem varje gång!
   - Ange ISO datastore path (t.ex. `[datastore1] iso/windows-server-2022.iso`)
   - Välj om snapshot ska skapas före uppgradering
   - Klicka på "Starta uppgradering"

5. **Övervaka progress**
   - Progress bar visar framsteg
   - Real-time logg visar detaljerad information (texten är läsbar och kan markeras/kopieras)
   - Status-meddelanden uppdateras kontinuerligt
   - Uppgraderingen pågår i bakgrunden på guest OS

6. **Hantera snapshots efter uppgradering** (rekommenderat workflow)
   - Efter uppgradering: Låt appägare verifiera att systemet fungerar
   - Gå till "Hantera snapshots"
   - Välj pre-upgrade snapshots att ta bort
   - Bekräfta borttagning (kan inte ångras!)
   - Frigör diskutrymme på datastore

## Konfiguration

Konfigurationen sparas i `~/conf.json` och kan redigeras via GUI:s inställningsdialog:

```json
{
  "vcenter": {
    "vcenter_url": "vcenter.example.local",
    "username": "administrator@vsphere.local",
    "mode": "password",
    "insecure": true
  },
  "defaults": {
    "snapshot_name_prefix": "pre-upgrade",
    "iso_datastore_path": "[datastore1] iso/windows-server-2022.iso",
    "skip_memory_in_snapshot": true,
    "glvk": "WX4NM-KYWYW-QJJR4-XV3QB-6VM33",
    "guest_username": "upgrade"
  },
  "upgrade": {
    "parallel": 2,
    "reboot": true,
    "timeout_minutes": 90,
    "precheck_disk_gb": 10
  },
  "logging": {
    "level": "info",
    "file": "osupgrader.log"
  },
  "ui": {
    "language": "sv"
  }
}
```

### Konfigurationsalternativ

#### vCenter-inställningar
- **vcenter_url**: vCenter server hostname
- **username**: vCenter användarnamn
- **insecure**: Tillåt osignerade SSL-certifikat

#### Guest OS-credentials
- **guest_username**: Windows admin-användare på VMs (t.ex. `upgrade`)
  - Sparas i konfigurationsfilen för bekvämlighet
  - Om användarnamn saknar domän (`\` eller `@`) läggs VM:ens domän till automatiskt
  - Exempel: `upgrade` → `upgrade@domain1.local` (baserat på VM:ens domän)
- **guest_password**: Windows admin-lösenord (masked i GUI med visa/dölj-knapp)
  - ⚠️ **Säkerhetsanmärkning**: Lösenord sparas endast i minnet under applikationens körning och skrivs ALDRIG till konfigurationsfilen

#### Upgrade-inställningar
- **snapshot_name_prefix**: Prefix för snapshot-namn
- **iso_datastore_path**: Sökväg till Windows Server 2022 ISO
- **skip_memory_in_snapshot**: Hoppa över minne i snapshot (snabbare)
- **glvk**: Windows Server 2022 Datacenter GVLK-nyckel
- **parallel**: Antal parallella uppgraderingar (1-10)
- **reboot**: Starta om automatiskt efter uppgradering
- **timeout_minutes**: Timeout för uppgradering per VM
- **precheck_disk_gb**: Minimum ledigt diskutrymme (GB)
- **poweroff_minutes**: Max tid att vänta på att Windows stänger av sig själv innan vCenter forcerar power off

#### Timeout-inställningar
- **signal_script_seconds**: Väntetid på att signaltask-scriptet slutförs
- **signal_files_minutes**: Väntetid på att scheduled-taskens signalfiler dyker upp
- **target_os_minutes**: Max tid att vänta på målsatt OS-version
- **poweroff_minutes**: Max tid att vänta på gäst-shutdown innan hård power off

## Uppgraderingsprocess

1. **Validering**
   - Validera guest credentials (förhindrar account lockout)
   - Kontrollera att ISO-filen finns på datastoren
   - Kontrollera diskutrymme på guest OS (minst 10 GB ledigt)
   - Kontrollera att VM är påslagen och VMware Tools körs

2. **Snapshot**
   - Skapa snapshot för återställning (valfritt)
   - Verifiera att snapshot skapades korrekt
   - Namnformat: `pre-upgrade-pre-YYYYMMDD-HHMM`

3. **ISO-montering**
   - Montera Windows Server 2022 ISO till CD-ROM
   - Verifiera att ISO är monterad

4. **Uppgradering**
   - Kör PowerShell upgrade-script via VMware Tools
   - Scriptet detekterar Core/Desktop automatiskt
   - Väljer rätt WIM-image index (3=Core, 4=Desktop)
   - Startar Windows Setup med `/auto upgrade /noreboot`
   - Väntar på att setup.exe slutförs (med `-Wait`)
   - Schemalägger en mjuk shutdown i Windows (60 sekunder för att städa upp tjänster)

5. **Övervakning**
   - Pollning av PowerShell script-exit och kontroll av exit code
   - Väntar på att VM går till `poweredOff`, och forcerar `PowerOff` via vCenter om det inte sker inom `poweroff_minutes`
   - Sover 60 sekunder och `PowerOn`:ar VM:en via vCenter innan nästa fas
   - Pollning av VMware Tools/OS-version varje 45 sekunder tills Windows Server 2022/2025 rapporteras
   - Timeout efter konfigurerad tid (standard: 90 minuter + konfigurerbar power-off timeout)

6. **Avslutning**
   - Väntar på scheduled-taskens signalfiler (task-baserad indikator) för att se att inloggningsmiljön är klar
   - Demontera ISO när uppgraderingen är klar
   - Verifierar att OS-version är 2022 eller 2025

## Projektstruktur

```
osupgrader-gui/
├── cmd/
│   └── osupgrader-gui/
│       └── main.go              # Huvudprogrammet (med -d/--debug flagga)
├── internal/
│   ├── config/
│   │   └── config.go            # Konfigurationshantering
│   ├── debug/
│   │   └── logger.go            # Debug-loggning till fil
│   ├── vcenter/
│   │   ├── client.go            # vCenter-klient och inloggning
│   │   ├── inventory.go         # VM-inventory-hantering (med domän)
│   │   ├── snapshot.go          # Snapshot-operationer
│   │   └── types.go             # Datatyper (VMInfo med Domain)
│   ├── upgrade/
│   │   ├── upgrade.go           # Uppgraderingslogik (auto-domain append)
│   │   ├── validators.go        # Validerings-funktioner
│   │   └── iso.go               # ISO-hantering
│   └── gui/
│       ├── app.go               # Huvudapplikation (DPI-skalning)
│       ├── login.go             # Login-skärm
│       ├── vmselection.go       # VM-selection-skärm (med Domain-kolumn)
│       ├── upgrade.go           # Upgrade-workflow-skärm
│       ├── snapshots.go         # Snapshot-hanteringsskärm (NY!)
│       └── settings.go          # Inställningsdialog
├── go.mod
├── go.sum
└── README.md
```

## Säkerhetsfunktioner

- **Lösenord lagras aldrig** i konfigurationsfilen
- **Säker debug-loggning**: Lösenord loggas aldrig i klartext (endast längd)
- **Credential-validering**: Kontrollerar credentials före uppgradering för att förhindra account lockout
- **Windows SSPI/Kerberos-stöd** för säker single sign-on utan lösenordsinmatning
- **Snapshot-verifiering** förhindrar dataförlust
- **Snapshot-hantering med bekräftelse**: Bekräftelsedialog för borttagning av snapshots
- **ISO-validering** före snapshot sparar tid
- **Thread-safe** operationer med mutex-skydd
- **VMware Tools crash recovery** hanterar omstarter under uppgradering
- **Timeout-hantering** förhindrar hängande uppgraderingar
- **Multi-domän support**: Automatisk domän-append minskar risk för fel användarnamn

## Windows SSPI/Kerberos-autentisering

SSPI (Security Support Provider Interface) är Microsofts API för autentisering och säkerhet i Windows. När du använder SSPI-inloggning:

1. **Transparent autentisering**: Applikationen använder dina Windows-credentials automatiskt
2. **Ingen lösenordsinmatning**: Du behöver inte ange lösenord - perfekt för smartcard/token-användare
3. **Domän-integration**: Fungerar sömlöst i Active Directory-miljöer
4. **Kerberos-protokoll**: Säker ticket-baserad autentisering mot vCenter
5. **SPN-baserad**: Använder Service Principal Name `host/vcenter.domain.local` för autentisering

**Tekniska detaljer:**
- Implementerad via `github.com/alexbrainman/sspi/negotiate`
- Stöder multi-round SSPI-handshake med `SSPIChallenge`
- Kompatibel med både PowerCLI och standard govmomi-sessions
- Endast tillgänglig på Windows-plattformen (stub på Linux/macOS)

## Felsökning

### Debug-loggning
För detaljerad troubleshooting, starta applikationen med debug-flaggan:
```bash
./osupgrader-gui -d
```

Detta skapar `debuglogg.txt` i samma mapp som programmet med:
- Alla API-anrop till vCenter
- Guest operations-detaljer
- Autentiseringsförsök (username och lösenordslängd, men INTE lösenordet)
- PowerShell script-exekvering
- Snapshot-operationer
- ISO-montering/demontering
- Alla fel med stack traces

**Viktig information i debug-loggen:**
- Timestamps för alla operationer
- VM-namn och domän-information
- Exit codes från PowerShell-script
- OS-version före och efter uppgradering

### Inloggning misslyckades
- **Lösenordsautentisering**:
  - Kontrollera vCenter-URL och användaruppgifter
  - Aktivera "Tillåt osignerade certifikat" om self-signed cert används
  - Kontrollera nätverksåtkomst till vCenter
- **SSPI/Kerberos-autentisering**:
  - Fungerar endast på Windows
  - Kräver att du är inloggad med ett domänkonto
  - vCenter-servern måste vara Windows-integrerad (Active Directory)
  - Kontrollera att Kerberos SPN är korrekt konfigurerad (`host/vcenter.domain.local`)
  - På Linux används endast lösenordsautentisering

### ISO-validering misslyckades
- Kontrollera att ISO-sökvägen är korrekt: `[datastore1] iso/file.iso`
- Verifiera att datastoren finns och är tillgänglig
- Kontrollera att ISO-filen existerar på datastoren
- Använd debug-loggning för att se exakt vilken datastore som söks

### Autentisering mot guest OS misslyckades
- **Account lockout-problem**:
  - Applikationen validerar credentials INNAN uppgradering för att förhindra lockout
  - Om credentials är felaktiga, får du ett fel omedelbart utan upprepade försök
- **Multi-domän användning**:
  - Ange bara användarnamn utan domän (t.ex. `upgrade`)
  - Systemet lägger automatiskt till VM:ens domän
  - Kontrollera att VM:ens domän är korrekt i tabellvyn
  - Om auto-append inte fungerar, använd `DOMAIN\user` eller `user@domain.com`
- **Debug-tips**:
  - Kör med `-d` flagga
  - Kolla `debuglogg.txt` för att se vilket username som faktiskt används
  - Exempel: `Auto-appended domain to username: upgrade@domain1.local`

### Uppgradering misslyckas
- Kontrollera att VMware Tools är installerade och körs
- Verifiera att guest-credentials är korrekta
- Kontrollera diskutrymme på guest OS (minst 10 GB)
- Se loggfilen `C:\Windows\Temp\upgrade.log` på guest OS
- **PowerShell script-problem**:
  - Kolla `C:\Windows\Temp\setup_stdout.log` och `setup_stderr.log`
  - Verifiera att setup.exe kördes (kolla PID i debug-loggen)
  - Kontrollera exit code från PowerShell-script (ska vara 0)
- **Timeout-problem**:
  - Standard timeout är 90 minuter
  - Öka timeout i inställningar om uppgraderingen tar längre tid
  - Långsamma VMs kan behöva 120-180 minuter

### Snapshot-hantering
- **Kan inte hitta snapshots**:
  - Kontrollera att `snapshot_name_prefix` i config matchar snapshot-namn
  - Default prefix är `pre-upgrade`
  - Snapshot-namn format: `pre-upgrade-pre-YYYYMMDD-HHMM`
- **Borttagning misslyckades**:
  - Kontrollera att inga andra operationer pågår på VM:en
  - Verifiera vCenter-permissions för snapshot-borttagning
  - Vissa snapshots kan vara låsta av backup-jobb

## Relaterade projekt

- **osupgrader** - TUI-version med batch-läge för CLI-automatisering

## Licens

Internt projekt - kontakta projektägaren för licensinformation.

## Support

För buggrapporter och funktionsförfrågningar, kontakta utvecklingsteamet.
