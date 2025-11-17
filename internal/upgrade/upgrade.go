package upgrade

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	efs "github.com/skabbio1976/eFS"
	"github.com/vmware/govmomi/guest"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/yourusername/osupgrader-gui/internal/config"
	"github.com/yourusername/osupgrader-gui/internal/debug"
	"github.com/yourusername/osupgrader-gui/internal/vcenter"
)

//go:embed assets/*.ps1
var assetsFS embed.FS

// UpgradeOptions innehåller alla options för en uppgradering
type UpgradeOptions struct {
	VMInfo         vcenter.VMInfo
	GuestUsername  string
	GuestPassword  string
	ISOPath        string
	CreateSnapshot bool
	SnapshotName   string
	Config         *config.AppConfig
}

// UpgradeResult innehåller resultatet av en uppgradering
type UpgradeResult struct {
	VMName  string
	Success bool
	Error   error
	Steps   []UpgradeStep
}

// UpgradeStep representerar ett steg i uppgraderingsprocessen
type UpgradeStep struct {
	Name      string
	Status    string // "pending", "in_progress", "completed", "failed"
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

// UpgradeSingleVM uppgraderar en enskild VM
func UpgradeSingleVM(vm *object.VirtualMachine, opts UpgradeOptions) error {
	// debug.LogFunction("UpgradeSingleVM",
	// 	"VM", opts.VMInfo.Name,
	// 	"ISOPath", opts.ISOPath,
	// 	"CreateSnapshot", opts.CreateSnapshot,
	// 	"SnapshotName", opts.SnapshotName,
	// 	"TimeoutMinutes", opts.Config.Upgrade.TimeoutMinutes,
	// 	"PrecheckDiskGB", opts.Config.Upgrade.PrecheckDiskGB,
	// )

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opts.Config.Upgrade.TimeoutMinutes)*time.Minute)
	defer cancel()

	// 0. Kontrollera om upgrade redan pågår
	// debug.Log("Step 0: Checking if upgrade already in progress...")
	inProgress, err := CheckUpgradeInProgress(ctx, vm)
	if err != nil {
		debug.LogError("CheckUpgradeInProgress", err, "VM", opts.VMInfo.Name)
		return fmt.Errorf("status check: %w", err)
	}
	if inProgress {
		debug.LogError("UpgradeInProgress", fmt.Errorf("upgrade already in progress"), "VM", opts.VMInfo.Name)
		return fmt.Errorf("upgrade verkar redan pågå på denna VM")
	}
	// debug.LogSuccess("CheckUpgradeInProgress", "VM", opts.VMInfo.Name)

	// 1. Disk precheck
	if opts.Config.Upgrade.PrecheckDiskGB > 0 {
		// debug.Log("Step 1: Disk space precheck...")
		sysDrive, err := GetSystemDrive(ctx, vm)
		if err != nil {
			debug.LogError("GetSystemDrive", err, "VM", opts.VMInfo.Name)
			return fmt.Errorf("kunde inte hitta system drive: %w", err)
		}
		// debug.Log("System drive detected: %s", sysDrive)

		free, err := GetDiskFreeGB(ctx, vm, sysDrive)
		if err != nil {
			debug.LogError("GetDiskFreeGB", err, "VM", opts.VMInfo.Name, "Drive", sysDrive)
			return fmt.Errorf("diskcheck: %w", err)
		}
		// debug.Log("Free space: %d GB (required: %d GB)", free, opts.Config.Upgrade.PrecheckDiskGB)

		if int(free) < opts.Config.Upgrade.PrecheckDiskGB {
			debug.LogError("InsufficientDiskSpace", fmt.Errorf("not enough disk space"),
				"VM", opts.VMInfo.Name, "Free", free, "Required", opts.Config.Upgrade.PrecheckDiskGB)
			return fmt.Errorf("disk: %d GB ledigt < krav %d GB", free, opts.Config.Upgrade.PrecheckDiskGB)
		}
		// debug.LogSuccess("DiskPrecheck", "VM", opts.VMInfo.Name, "Free", free)
	}

	// 2. Snapshot
	if opts.CreateSnapshot {
		// debug.Log("Step 2: Creating snapshot...")
		// debug.Log("Snapshot name: %s, Include memory: %v", opts.SnapshotName, !opts.Config.Defaults.SkipMemoryInSnapshot)

		if err := vcenter.CreateSnapshot(ctx, vm, opts.SnapshotName, "Pre upgrade", !opts.Config.Defaults.SkipMemoryInSnapshot, false); err != nil {
			debug.LogError("CreateSnapshot", err, "VM", opts.VMInfo.Name, "SnapshotName", opts.SnapshotName)
			return fmt.Errorf("snapshot: %w", err)
		}
		// debug.LogSuccess("CreateSnapshot", "VM", opts.VMInfo.Name, "Name", opts.SnapshotName)
	}

	// 3. Mount ISO
	// debug.Log("Step 3: Mounting ISO...")
	// debug.Log("ISO path: %s", opts.ISOPath)
	if err := MountISO(ctx, vm, opts.ISOPath); err != nil {
		debug.LogError("MountISO", err, "VM", opts.VMInfo.Name, "ISOPath", opts.ISOPath)
		return fmt.Errorf("mount iso: %w", err)
	}
	// debug.LogSuccess("MountISO", "VM", opts.VMInfo.Name, "ISOPath", opts.ISOPath)

	// 4. Förbered guest credentials och setup post-reboot signaling FÖRE uppgraderingen
	// debug.Log("Step 4: Preparing guest credentials and post-reboot signaling...")

	// Bygg username - lägg till domän om den inte redan finns
	username := opts.GuestUsername
	if !strings.Contains(username, "\\") && !strings.Contains(username, "@") {
		// Inget domän-format hittat, lägg till @domain om vi har det
		if opts.VMInfo.Domain != "" {
			// Extrahera domän från FQDN (t.ex. srv01.testdom.se -> testdom.se)
			domain := opts.VMInfo.Domain
			if strings.Contains(domain, ".") {
				parts := strings.SplitN(domain, ".", 2)
				if len(parts) == 2 {
					domain = parts[1] // Ta bort hostname-delen
				}
			}
			username = username + "@" + domain
			// debug.Log("Auto-appended domain to username: %s (extracted from FQDN: %s)", username, opts.VMInfo.Domain)
		}
	}

	gc := vcenter.GuestCreds{
		User: username,
		Pass: opts.GuestPassword,
	}
	// debug.Log("Guest user: %s, GVLK: %s", username, truncateGVLK(opts.Config.Defaults.Glvk))

	// 4.5. Ladda upp alla PowerShell-scripts till gästen
	debug.Log("Step 4.5: Uploading all PowerShell scripts to guest (BEFORE upgrade)...")
	if err := uploadScriptsToGuest(ctx, vm, gc, opts.VMInfo.Name); err != nil {
		debug.LogError("UploadScripts", err, "VM", opts.VMInfo.Name)
		return fmt.Errorf("kunde inte ladda upp scripts: %w", err)
	}

	// 4.6. Kör createsignaltasks.ps1 för att sätta upp post-reboot signalering
	debug.Log("Step 4.6: Setting up post-reboot signal mechanisms (BEFORE upgrade)...")
	if err := executeSignalTaskScript(ctx, vm, gc, opts.VMInfo.Name, opts.Config.Timeouts); err != nil {
		debug.LogError("ExecuteSignalTaskScript", err, "VM", opts.VMInfo.Name)
		// Inte kritiskt - fortsätt ändå, men logga varning
		debug.Log("WARNING: Failed to set up signal task script, upgrade will continue but signal detection may fail")
	} else {
		debug.LogSuccess("SignalTaskSetup", "VM", opts.VMInfo.Name)
	}

	// 5. Guest upgrade script
	debug.Log("Step 5: Starting guest upgrade script...")

	pid, err := startGuestUpgrade(ctx, vm, gc, opts.Config.Defaults.Glvk)
	if err != nil {
		debug.LogError("StartGuestUpgrade", err, "VM", opts.VMInfo.Name, "GuestUser", opts.GuestUsername)
		return fmt.Errorf("guest script: %w", err)
	}
	debug.LogSuccess("StartGuestUpgrade", "VM", opts.VMInfo.Name, "PID", pid)

	// 6. Vänta på att scriptet avslutas
	debug.Log("Step 6: Waiting for upgrade script to complete (PID: %d)...", pid)
	exitCode, err := waitForProcessExit(ctx, vm, gc, pid, opts.VMInfo.Name)
	if err != nil {
		debug.LogError("WaitForProcessExit", err, "VM", opts.VMInfo.Name, "PID", pid)
		return fmt.Errorf("script wait: %w", err)
	}
	if exitCode != 0 {
		debug.LogError("ScriptExitCode", fmt.Errorf("non-zero exit code"), "VM", opts.VMInfo.Name, "ExitCode", exitCode)
		return fmt.Errorf("upgrade script failed med exit code %d", exitCode)
	}
	debug.LogSuccess("ScriptCompleted", "VM", opts.VMInfo.Name, "ExitCode", exitCode)

	// 7. Vänta på att gästen stänger av (scriptet är klart men kan ta tid innan Windows stoppar) och hantera shutdown/power cycle
	debug.Log("Step 7: Giving Windows 60 seconds before checking power state...")
	select {
	case <-time.After(60 * time.Second):
	case <-ctx.Done():
		return fmt.Errorf("avbruten innan shutdown-check hann köras: %w", ctx.Err())
	}

	shutdownTimeout := time.Duration(opts.Config.Timeouts.PowerOffMinutes) * time.Minute
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Minute
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	abortGuestCheck := false

	pollTicker := time.NewTicker(15 * time.Second)
	defer pollTicker.Stop()

waitForPowerOff:
	for {
		select {
		case <-shutdownCtx.Done():
			debug.Log("WARNING: Guest shutdown timeout reached, attempting forced power off...")
			abortGuestCheck = true
			break waitForPowerOff
		case <-pollTicker.C:
			var o mo.VirtualMachine
			if err := vm.Properties(shutdownCtx, vm.Reference(), []string{"runtime.powerState"}, &o); err != nil {
				debug.Log("[%s] WARNING: Failed to query runtime.powerState while waiting for shutdown: %v", opts.VMInfo.Name, err)
				continue
			}
			if o.Runtime.PowerState == types.VirtualMachinePowerStatePoweredOff {
				debug.Log("Step 7: VM is powered off")
				break waitForPowerOff
			}
		}
	}

	if abortGuestCheck {
		debug.Log("Step 7: Forcing power off via vCenter...")
		powerOffCtx, powerOffCancel := context.WithTimeout(ctx, 10*time.Minute)
		defer powerOffCancel()
		powerOffTask, err := vm.PowerOff(powerOffCtx)
		if err != nil {
			debug.LogError("PowerOff", err, "VM", opts.VMInfo.Name)
			return fmt.Errorf("power off: %w", err)
		}
		if err := powerOffTask.Wait(powerOffCtx); err != nil {
			debug.LogError("PowerOffWait", err, "VM", opts.VMInfo.Name)
			return fmt.Errorf("power off wait: %w", err)
		}
		debug.LogSuccess("PowerOff", "VM", opts.VMInfo.Name)
	}

	debug.Log("Step 7: Waiting 60 seconds before powering on via vCenter...")
	powerOnDelay := time.NewTimer(60 * time.Second)
	defer powerOnDelay.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("avbruten innan power on hann köras: %w", ctx.Err())
	case <-powerOnDelay.C:
	}

	powerOnCtx, powerOnCancel := context.WithTimeout(ctx, 10*time.Minute)
	defer powerOnCancel()

	powerOnTask, err := vm.PowerOn(powerOnCtx)
	if err != nil {
		debug.LogError("PowerOn", err, "VM", opts.VMInfo.Name)
		return fmt.Errorf("power on: %w", err)
	}
	if err := powerOnTask.Wait(powerOnCtx); err != nil {
		debug.LogError("PowerOnWait", err, "VM", opts.VMInfo.Name)
		return fmt.Errorf("power on wait: %w", err)
	}
	debug.LogSuccess("PowerOn", "VM", opts.VMInfo.Name)

	// Kort väntan innan vi går vidare så VMware Tools hinner initialiseras
	time.Sleep(20 * time.Second)

	// 8. ProcessMonitor är temporärt inaktiverad eftersom den blir klar långt innan OS går att logga in i
	debug.Log("Step 8: Skipping ProcessMonitor (disabled; completes before OS logon)")
	/*
		// Detta step väntar implicit på att VMware Tools är tillgängliga och OS är uppe
		debug.Log("Step 8: Running ProcessMonitor to verify critical Windows processes...")
		if err := runProcessMonitor(ctx, vm, gc, opts.VMInfo.Name, opts.Config.Timeouts); err != nil {
			debug.LogError("RunProcessMonitor", err, "VM", opts.VMInfo.Name)
			// Inte kritiskt - fortsätt ändå
			debug.Log("WARNING: ProcessMonitor failed, continuing anyway...")
		} else {
			debug.LogSuccess("ProcessMonitorCompleted", "VM", opts.VMInfo.Name)
		}
	*/

	// 8.5. Verifiera att OS-versionen matchar målversionen
	targetOS := []string{"windows server 2022", "windows server 2025"}
	debug.Log("Step 8.5: Validating guest OS version against targets: %v...", targetOS)
	if err := waitForTargetOS(ctx, vm, targetOS, opts.VMInfo.Name, time.Duration(opts.Config.Timeouts.TargetOSMinutes)*time.Minute); err != nil {
		debug.LogError("WaitForTargetOS", err, "VM", opts.VMInfo.Name)
		return fmt.Errorf("os version: %w", err)
	}
	debug.LogSuccess("TargetOSDetected", "VM", opts.VMInfo.Name)

	// 8.6. SetupEventVerification är temporärt inaktiverad då event-loggen uppdateras långt före inloggningstid
	debug.Log("Step 8.6: Skipping SetupEventVerification (disabled; events appear before OS logon)")
	/*
		eventLookback := time.Duration(opts.Config.Timeouts.EventLogMinutes) * time.Minute
		if eventLookback <= 0 {
			eventLookback = 10 * time.Minute
		}
		debug.Log("Step 8.6: Checking Windows Setup event log for completion markers (lookback: %v)...", eventLookback)
		eventCtx, eventCancel := context.WithTimeout(ctx, eventLookback)
		defer eventCancel()
		if err := verifySetupCompletionEvent(eventCtx, vm, gc, opts.VMInfo.Name, upgradeStart, eventLookback); err != nil {
			debug.LogError("SetupEventVerification", err, "VM", opts.VMInfo.Name)
			return fmt.Errorf("setup event verification: %w", err)
		}
		debug.LogSuccess("SetupEventDetected", "VM", opts.VMInfo.Name)
	*/

	// 9. Verifiera att Windows är redo genom att vänta på signalfilen från scheduled task
	debug.Log("Step 9: Waiting for post-reboot task signal file...")
	if err := waitForPostRebootSignals(ctx, vm, gc, opts.VMInfo.Name, opts.Config.Timeouts); err != nil {
		// Kontrollera om det är ett timeout-fel
		if strings.Contains(err.Error(), "LOGONUI_TIMEOUT") {
			// Logga varning men fortsätt ändå
			debug.Log("WARNING: Task signal file not created within timeout - server %s should be checked manually", opts.VMInfo.Name)
			debug.Log("WARNING: %v", err)
		} else {
			// Annat fel, avbryt uppgraderingen
			debug.LogError("WaitForSignalFiles", err, "VM", opts.VMInfo.Name)
			return fmt.Errorf("signal file check: %w", err)
		}
	} else {
		debug.LogSuccess("WaitForSignalFiles", "VM", opts.VMInfo.Name)
	}

	// 10. Unmount ISO
	debug.Log("Step 10: Unmounting ISO...")
	if err := UnmountISO(context.Background(), vm); err != nil {
		debug.Log("WARNING: unmount ISO failed: %v", err)
	} else {
		debug.LogSuccess("UnmountISO", "VM", opts.VMInfo.Name)
	}

	debug.LogSuccess("UpgradeSingleVM", "VM", opts.VMInfo.Name)
	return nil
}

func startGuestUpgrade(ctx context.Context, vm *object.VirtualMachine, gc vcenter.GuestCreds, gvlk string) (int64, error) {
	c := vm.Client()

	if gvlk == "" {
		return 0, fmt.Errorf("ingen GVLK-nyckel i config")
	}

	// debug.Log("Creating guest OperationsManager...")

	// Skapa OperationsManager för guest operations
	opsMgr := guest.NewOperationsManager(c, vm.Reference())

	// Hämta ProcessManager
	pm, err := opsMgr.ProcessManager(ctx)
	if err != nil {
		debug.LogError("GetProcessManager", err)
		return 0, fmt.Errorf("kunde inte få ProcessManager: %w", err)
	}

	// debug.Log("ProcessManager created successfully")

	// Validera credentials FÖRST innan vi försöker köra något
	// Detta förhindrar account lockout från upprepade misslyckade försök
	auth := &types.NamePasswordAuthentication{Username: gc.User, Password: gc.Pass}

	// debug.Log("=== AUTHENTICATION DEBUG ===")
	// debug.Log("Username (raw): %s", gc.User)
	// debug.Log("Password (raw): %s", gc.Pass) // Kommenterad av säkerhetsskäl
	// debug.Log("Username length: %d chars", len(gc.User))
	// debug.Log("Password length: %d chars", len(gc.Pass))
	// debug.Log("Username contains backslash: %v", strings.Contains(gc.User, "\\"))
	// debug.Log("Username contains @: %v", strings.Contains(gc.User, "@"))
	// debug.Log("===========================")

	// Hämta AuthManager för att validera credentials
	am, err := opsMgr.AuthManager(ctx)
	if err != nil {
		debug.LogError("GetAuthManager", err)
		return 0, fmt.Errorf("kunde inte få AuthManager: %w", err)
	}

	// debug.Log("Calling ValidateCredentials with:")
	// debug.Log("  auth.Username = %q", auth.Username)
	// debug.Log("  auth.Password = %q", auth.Password) // Kommenterad av säkerhetsskäl

	// Validera credentials innan vi försöker starta något
	err = am.ValidateCredentials(ctx, auth)
	if err != nil {
		debug.LogError("ValidateCredentials", err,
			"Username", gc.User,
			"PasswordLength", len(gc.Pass),
		)
		return 0, fmt.Errorf("autentisering misslyckades för användare '%s': %w", gc.User, err)
	}

	// debug.LogSuccess("ValidateCredentials", "Username", gc.User)
	// debug.Log("Guest credentials validated successfully!")

	// Extrahera och läs in PowerShell-scriptet från embedded FS
	scriptTemplate, cleanup, err := extractAndReadPowerShellScript()
	if err != nil {
		debug.LogError("ExtractPowerShellScript", err)
		return 0, fmt.Errorf("kunde inte extrahera PowerShell-script: %w", err)
	}
	defer cleanup()

	// Ersätt param([string]$GLVK) med att sätta variabeln direkt
	script := strings.Replace(scriptTemplate, "param([string]$GLVK)", fmt.Sprintf("$GLVK = '%s'", gvlk), 1)

	// debug.Log("PowerShell script prepared, GVLK: %s", truncateGVLK(gvlk))

	encoded := encodePowerShell(script)

	// debug.Log("PowerShell script encoded, length: %d bytes", len(encoded))

	spec := &types.GuestProgramSpec{
		ProgramPath:      "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		Arguments:        "-NoLogo -NonInteractive -ExecutionPolicy Bypass -EncodedCommand " + encoded,
		WorkingDirectory: "C:\\Windows\\Temp",
	}

	// debug.Log("Starting program in guest...")
	// debug.Log("Program: %s", spec.ProgramPath)
	// debug.Log("Working dir: %s", spec.WorkingDirectory)

	// Använd den validerade authen (credentials redan validerade ovan)
	pid, err := pm.StartProgram(ctx, auth, spec)
	if err != nil {
		debug.LogError("ProcessManager.StartProgram", err,
			"ProgramPath", spec.ProgramPath,
			"WorkingDir", spec.WorkingDirectory,
		)
		return 0, fmt.Errorf("kunde inte starta upgrade-script: %w", err)
	}

	// debug.Log("Program started successfully, PID: %d", pid)
	// debug.LogSuccess("StartProgram", "PID", pid)

	return pid, nil
}

func waitForProcessExit(ctx context.Context, vm *object.VirtualMachine, gc vcenter.GuestCreds, pid int64, serverName string) (int32, error) {
	c := vm.Client()
	opsMgr := guest.NewOperationsManager(c, vm.Reference())
	pm, err := opsMgr.ProcessManager(ctx)
	if err != nil {
		return -1, fmt.Errorf("kunde inte få ProcessManager: %w", err)
	}

	auth := &types.NamePasswordAuthentication{Username: gc.User, Password: gc.Pass}
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	debug.Log("[%s] Polling for process exit (PID: %d)...", serverName, pid)

	for {
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		case <-ticker.C:
			// ListProcesses returnerar lista med process-info
			procs, err := pm.ListProcesses(ctx, auth, []int64{pid})
			if err != nil {
				debug.Log("[%s] WARNING: ListProcesses error: %v", serverName, err)
				continue
			}

			if len(procs) == 0 {
				debug.Log("[%s] Process not found in list - might have exited", serverName)
				return 0, nil // Process finns inte = antagligen avslutad OK
			}

			proc := procs[0]
			if proc.EndTime != nil {
				// Process har avslutat
				debug.Log("[%s] Process exited at %v with exit code %d", serverName, proc.EndTime, proc.ExitCode)
				return proc.ExitCode, nil
			}

			debug.Log("[%s] Process still running (PID: %d)...", serverName, pid)
		}
	}
}

func waitForTargetOS(ctx context.Context, vm *object.VirtualMachine, targets []string, serverName string, timeout time.Duration) error {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	lowerTargets := make([]string, len(targets))
	for i, t := range targets {
		lowerTargets[i] = strings.ToLower(t)
	}

	consecutiveErrors := 0
	maxConsecutiveErrors := 5

	var timeoutCh <-chan time.Time
	if timeout > 0 {
		timeoutCh = time.After(timeout)
	}

	debug.Log("[%s] Polling for OS version change (target: %v, timeout: %v)...", serverName, targets, timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutCh:
			return fmt.Errorf("timeout while waiting for OS version to match %v (waited %v)", targets, timeout)
		case <-ticker.C:
			var o mo.VirtualMachine
			if err := vm.Properties(ctx, vm.Reference(), []string{"guest.guestFullName", "guest.toolsRunningStatus"}, &o); err != nil {
				consecutiveErrors++
				debug.Log("[%s] WARNING: Properties error (%d/%d): %v", serverName, consecutiveErrors, maxConsecutiveErrors, err)
				if consecutiveErrors >= maxConsecutiveErrors {
					return fmt.Errorf("för många fel i rad vid polling av OS-version (%d): %w", consecutiveErrors, err)
				}
				continue
			}

			consecutiveErrors = 0

			if o.Guest != nil {
				toolsStatus := o.Guest.ToolsRunningStatus

				if toolsStatus != "guestToolsRunning" && toolsStatus != "" {
					debug.Log("[%s] VMware Tools status: %s (waiting for guestToolsRunning)", serverName, toolsStatus)
					continue
				}

				if o.Guest.GuestFullName != "" {
					g := strings.ToLower(o.Guest.GuestFullName)
					debug.Log("[%s] Current OS: %s", serverName, o.Guest.GuestFullName)
					for _, t := range lowerTargets {
						if strings.Contains(g, t) {
							debug.Log("[%s] Target OS detected: %s contains %s", serverName, o.Guest.GuestFullName, t)
							return nil
						}
					}
				}
			}
		}
	}
}

// waitForPostRebootSignals verifierar att Windows är redo genom att leta efter signalfiler
// - Scheduled task signal: skapas av en scheduled task vid startup
// Detta är den mest tillförlitliga metoden för att veta att systemet är helt klart efter reboot
func waitForPostRebootSignals(ctx context.Context, vm *object.VirtualMachine, gc vcenter.GuestCreds, serverName string, timeouts config.TimeoutConfig) error {
	c := vm.Client()
	opsMgr := guest.NewOperationsManager(c, vm.Reference())

	fm, err := opsMgr.FileManager(ctx)
	if err != nil {
		return fmt.Errorf("kunde inte få FileManager: %w", err)
	}

	auth := &types.NamePasswordAuthentication{Username: gc.User, Password: gc.Pass}
	taskSignalFile := "C:\\Temp\\osupgrader_ready.txt"

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Timeout för underdimensionerade servrar
	signalTimeout := time.Duration(timeouts.SignalFilesMinutes) * time.Minute
	if signalTimeout <= 0 {
		signalTimeout = 30 * time.Minute
	}
	timeout := time.After(signalTimeout)

	debug.Log("[%s] Polling for post-reboot task signal file (every 30s, timeout %v)...", serverName, signalTimeout)
	debug.Log("[%s] Task signal file: %s", serverName, taskSignalFile)

	taskFileFound := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			// Vid timeout, returnera ett specifikt fel som kan hanteras annorlunda
			return fmt.Errorf("LOGONUI_TIMEOUT: Task signal file hittades inte inom %v - servern bör kontrolleras manuellt (Task: %v)", signalTimeout, taskFileFound)
		case <-ticker.C:
			// Kolla task signal file
			if !taskFileFound {
				_, err := fm.InitiateFileTransferFromGuest(ctx, auth, taskSignalFile)
				if err == nil {
					taskFileFound = true
					debug.Log("[%s] ✓ Task signal file detected at %s", serverName, time.Now().Format("2006-01-02 15:04:05"))
				}
			}

			// Om task-filen hittats, SUCCESS!
			if taskFileFound {
				debug.Log("[%s] SUCCESS! Task signal file detected - system is ready!", serverName)
				debug.Log("[%s] Running cleanup script to remove signal file...", serverName)

				// Kör cleanup.ps1 för att ta bort signalfilerna
				cleanupScript, cleanupCleanup, err := extractAndReadCleanupScript()
				if err != nil {
					debug.Log("WARNING: Could not extract cleanup script: %v", err)
				} else {
					defer cleanupCleanup()

					pm, pmErr := opsMgr.ProcessManager(ctx)
					if pmErr == nil {
						encoded := encodePowerShell(cleanupScript)
						spec := &types.GuestProgramSpec{
							ProgramPath: "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
							Arguments:   "-NoLogo -NonInteractive -ExecutionPolicy Bypass -EncodedCommand " + encoded,
						}
						_, err := pm.StartProgram(ctx, auth, spec)
						if err != nil {
							debug.Log("WARNING: Cleanup script failed: %v", err)
						} else {
							debug.LogSuccess("CleanupScript", "Server", serverName)
						}
					}
				}

				return nil
			}

			// Visa status
			if !taskFileFound {
				debug.Log("[%s] Still waiting... (Task: %v)", serverName, taskFileFound)
			}
		}
	}
}

// func truncateGVLK(key string) string {
// 	if len(key) <= 10 {
// 		return "***MASKED***"
// 	}
// 	return key[:5] + "-*****-*****-" + key[len(key)-5:]
// }

// encodePowerShell UTF-16LE + base64
func encodePowerShell(s string) string {
	runes := []rune(s)
	utf16Vals := utf16.Encode(runes)
	bytes := make([]byte, len(utf16Vals)*2)
	for i, v := range utf16Vals {
		bytes[i*2] = byte(v)
		bytes[i*2+1] = byte(v >> 8)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

// extractAndReadPowerShellScript extraherar PowerShell-scriptet från embedded FS
// till användarens hemkatalog och läser in det som en string
func extractAndReadPowerShellScript() (string, func(), error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("kunde inte hitta hemkatalog: %w", err)
	}

	// debug.Log("Extraherar PowerShell-script till %s...", homeDir)

	// Extrahera filen till hemkatalogen
	extractedPath, cleanup, err := efs.ExtractFile(assetsFS, "assets/uppgradeos.ps1", "osupgrader_", homeDir)
	if err != nil {
		return "", nil, fmt.Errorf("kunde inte extrahera PowerShell-script: %w", err)
	}

	// debug.Log("PowerShell-script extraherat till: %s", extractedPath)

	// Läs in filen som string
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("kunde inte läsa PowerShell-script: %w", err)
	}

	// debug.LogSuccess("ExtractPowerShellScript", "Path", extractedPath, "Size", len(content))

	return string(content), cleanup, nil
}

// extractAndReadCleanupScript extraherar cleanup PowerShell-scriptet från embedded FS
// till användarens hemkatalog och läser in det som en string
func extractAndReadCleanupScript() (string, func(), error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("kunde inte hitta hemkatalog: %w", err)
	}

	// debug.Log("Extraherar cleanup-script till %s...", homeDir)

	// Extrahera filen till hemkatalogen
	extractedPath, cleanup, err := efs.ExtractFile(assetsFS, "assets/cleanup.ps1", "cleanup_", homeDir)
	if err != nil {
		return "", nil, fmt.Errorf("kunde inte extrahera cleanup-script: %w", err)
	}

	// debug.Log("Cleanup-script extraherat till: %s", extractedPath)

	// Läs in filen som string
	content, err := os.ReadFile(extractedPath)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("kunde inte läsa cleanup-script: %w", err)
	}

	// debug.LogSuccess("ExtractCleanupScript", "Path", extractedPath, "Size", len(content))

	return string(content), cleanup, nil
}

// uploadFileToGuest laddar upp en fil från embedded FS till gästen via VMware FileManager
func uploadFileToGuest(ctx context.Context, vm *object.VirtualMachine, gc vcenter.GuestCreds, embeddedPath, guestPath, serverName string) error {
	c := vm.Client()
	opsMgr := guest.NewOperationsManager(c, vm.Reference())

	// Hämta FileManager
	fm, err := opsMgr.FileManager(ctx)
	if err != nil {
		return fmt.Errorf("kunde inte få FileManager: %w", err)
	}

	auth := &types.NamePasswordAuthentication{Username: gc.User, Password: gc.Pass}

	debug.Log("[%s] Extraherar %s lokalt...", serverName, embeddedPath)

	// Extrahera scriptet lokalt
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("kunde inte hitta hemkatalog: %w", err)
	}

	extractedPath, cleanup, err := efs.ExtractFile(assetsFS, embeddedPath, "temp_", homeDir)
	if err != nil {
		return fmt.Errorf("kunde inte extrahera %s: %w", embeddedPath, err)
	}
	defer cleanup()

	// Läs filinnehållet
	fileContent, err := os.ReadFile(extractedPath)
	if err != nil {
		return fmt.Errorf("kunde inte läsa fil: %w", err)
	}

	debug.Log("[%s] Laddar upp till %s (%d bytes)...", serverName, guestPath, len(fileContent))

	// Initiera filöverföring
	fileTransferInfo, err := fm.InitiateFileTransferToGuest(ctx, auth, guestPath, &types.GuestFileAttributes{}, int64(len(fileContent)), true)
	if err != nil {
		return fmt.Errorf("kunde inte initiera filöverföring: %w", err)
	}

	// Ladda upp via HTTP PUT
	req, err := http.NewRequestWithContext(ctx, "PUT", fileTransferInfo, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("kunde inte skapa upload-request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = int64(len(fileContent))

	var uploadErr error
	err = c.Client.Do(ctx, req, func(resp *http.Response) error {
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			body, _ := io.ReadAll(resp.Body)
			uploadErr = fmt.Errorf("filuppladdning misslyckades med status %d: %s", resp.StatusCode, string(body))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("kunde inte ladda upp fil: %w", err)
	}
	if uploadErr != nil {
		return uploadErr
	}

	debug.LogSuccess("FileUpload", "Server", serverName, "Path", guestPath)
	return nil
}

// uploadScriptsToGuest laddar upp alla PowerShell-scripts som behövs till gästen
func uploadScriptsToGuest(ctx context.Context, vm *object.VirtualMachine, gc vcenter.GuestCreds, serverName string) error {
	debug.Log("[%s] Uploading all required PowerShell scripts to guest...", serverName)

	scripts := []struct {
		embedded string
		guest    string
	}{
		{"assets/createsignaltasks.ps1", "C:\\Temp\\createsignaltasks.ps1"},
		{"assets/processmonitor.ps1", "C:\\Temp\\processmonitor.ps1"},
		// {"assets/cleanup.ps1", "C:\\Temp\\cleanup.ps1"},
	}

	for _, script := range scripts {
		if err := uploadFileToGuest(ctx, vm, gc, script.embedded, script.guest, serverName); err != nil {
			return fmt.Errorf("failed to upload %s: %w", script.embedded, err)
		}
	}

	debug.LogSuccess("UploadScripts", "Server", serverName, "Count", len(scripts))
	return nil
}

// executeSignalTaskScript kör createsignaltasks.ps1 som redan finns på gästen
func executeSignalTaskScript(ctx context.Context, vm *object.VirtualMachine, gc vcenter.GuestCreds, serverName string, timeouts config.TimeoutConfig) error {
	c := vm.Client()
	opsMgr := guest.NewOperationsManager(c, vm.Reference())

	auth := &types.NamePasswordAuthentication{Username: gc.User, Password: gc.Pass}

	debug.Log("[%s] Executing createsignaltasks.ps1...", serverName)

	pm, err := opsMgr.ProcessManager(ctx)
	if err != nil {
		return fmt.Errorf("kunde inte få ProcessManager: %w", err)
	}

	spec := &types.GuestProgramSpec{
		ProgramPath: "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		Arguments:   "-NoProfile -ExecutionPolicy Bypass -File C:\\Temp\\createsignaltasks.ps1",
	}

	pid, err := pm.StartProgram(ctx, auth, spec)
	if err != nil {
		return fmt.Errorf("kunde inte starta script: %w", err)
	}

	debug.Log("[%s] Script startat, PID: %d", serverName, pid)

	// Vänta på att scriptet avslutas
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutSeconds := timeouts.SignalScriptSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			debug.Log("[%s] WARNING: Script timeout (%ds), fortsätter ändå...", serverName, timeoutSeconds)
			return nil
		case <-ticker.C:
			procs, err := pm.ListProcesses(ctx, auth, []int64{pid})
			if err != nil {
				debug.Log("[%s] WARNING: ListProcesses error: %v", serverName, err)
				continue
			}

			if len(procs) == 0 {
				debug.LogSuccess("CreateSignalTaskScript", "Server", serverName, "PID", pid)
				return nil
			}

			proc := procs[0]
			if proc.EndTime != nil {
				if proc.ExitCode != 0 {
					return fmt.Errorf("script avslutades med exit code %d", proc.ExitCode)
				}
				debug.LogSuccess("CreateSignalTaskScript", "Server", serverName, "PID", pid, "ExitCode", proc.ExitCode)
				return nil
			}
		}
	}
}
