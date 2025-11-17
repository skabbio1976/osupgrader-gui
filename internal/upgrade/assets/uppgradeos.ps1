param([string]$GLVK)
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
    Write-Log '=== Windows Server 2022 Datacenter Upgrade Start ==='

    # --- Detect installation type (Core/Desktop) ---
    Write-Log 'Detecting current installation type...'
    $installType = 'Desktop'
    $installationType = (Get-ItemProperty -Path "HKLM:\Software\Microsoft\Windows NT\CurrentVersion" -Name "InstallationType").InstallationType
    $imageIndex = $null
    if ($installationType -eq "Server Core") {
        $installType = 'Core'
        $imageIndex = 3
        # T.ex. använd sconfig eller dism
    } elseif ($installationType -eq "Server") {
        $imageIndex = 4# Anropa GUI-baserad uppgraderingslogik
        # T.ex. kan använda Windows Update UI eller andra GUI-verktyg
    } else {
        exit 1
        # Hantera edge case
    }
    $isCore = ($installType -eq 'Core')
    Write-Log "Detected installation type: $installType"

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


    # --- Build arguments ---
    $installOption = if ($isCore) { 'ServerDatacenterCore' } else { 'ServerDatacenter' }
    Write-Log "Fallback InstallOption: $installOption"
    Write-Log 'Product key: %s'
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