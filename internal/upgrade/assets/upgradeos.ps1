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
    $os = Get-WmiObject -Class Win32_OperatingSystem
    $sku = $os.OperatingSystemSKU
    
    Write-Log 'Detecting current installation type...'
    switch ($sku) {
        8  { 
            $imageIndex = 4
            $GLVK = 'WX4NM-KYWYW-QJJR4-XV3QB-6VM33'
            $edition = 'Datacenter Desktop'
        } # Datacenter Desktop
        7  { 
            $imageIndex = 2 
            $GLVK = 'VDYBN-27WPP-V4HQT-9VMD4-VMK7H'
            $edition = 'Standard Desktop'
        } # Standard Desktop
        48 { 
            $imageindex = 3
            $GLVK = 'WX4NM-KYWYW-QJJR4-XV3QB-6VM33'
            $edition = 'Datacenter Core'
        }   # Datacenter Core
        49 { 
            $imageindex = 1
            $GLVK = 'VDYBN-27WPP-V4HQT-9VMD4-VMK7H' 
            $edition = 'Standard Core'
        } # Standard Core
        default { exit 1 }
    }

    Write-Log '=== Upgrade Start ==='
    Write-Log "Edition: $edition"
    Write-Log "GLVK: $GLVK"
    Write-Log "Image Index: $imageIndex"

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
    Write-Log 'Product key: $GLVK'
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