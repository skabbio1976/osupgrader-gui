## Johan Worklog

### 2025-11-15
- **Timeout-konfiguration**: utökade `internal/config/config.go` med en dedikerad `TimeoutConfig` samt GUI-bindningar i `internal/gui/settings.go` så att alla tidsgränser (signal-script, VMware Tools, ProcessMonitor, TiWorker, mål-OS och eventlogg) kan justeras utan rebuild.  
- **Signaleringshårdning**: uppdaterade `internal/upgrade/assets/createsignaltasks.ps1` så att skriptet avbryter vid fel på `schtasks.exe` eller `reg.exe` och verifierar att uppgiften registrerades innan uppgraderingen får fortsätta.  
- **Efter-reboot-validering**:  
  - `internal/upgrade/upgrade.go` kör nu alltid `waitForTargetOS`, `verifySetupCompletionEvent` (event ID 19/55 inom konfigurerbar lookback) och `waitForTiWorkerCompletion` efter ProcessMonitor-steget.  
  - `waitForPostRebootSignals` och `executeSignalTaskScript` använder de nya timeoutvärdena och loggar exakta tidsgränser.  
  - Lade till PowerShell-hjälpare som hämtar Windows Setup-event via guest operations och säkrar TiWorker-övervakningen.  
  - **Referenser för readiness-detektion**:  
    - Windows Setup event ID 19/55 (slutförd installation) via *Microsoft-Windows-Setup/Operational* loggen — se <https://blog.matrixpost.net/mastering-windows-updates-microsoft-updates-part-1/>.  
    - TiWorker.exe som indikator på pågående efterservicing — diskussion och rekommendationer i <https://stackoverflow.com/questions/55155443/what-can-i-query-to-see-if-windows-is-booted-and-done-with-updates>.
- **GUI-förbättringar**:  
  - Byggde om `internal/gui/settings.go` till flikbaserad layout (`Guest & ISO`, `Upgrade`, `Timeouts`, `UI`) med scroll och tvåkolumners grid för timeoutfält, vilket löste problemet med för lång modal dialog.  
  - Lade till egna `Spara`/`Stäng`-knappar utanför `AppTabs` och ökade dialogstorleken (nu `900x700`) för bättre läsbarhet på mindre skärmar.


