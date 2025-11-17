# ProcessMonitor.ps1
# Script för att övervaka kritiska Windows-processer med detaljerad loggning
# Loggar exakt när varje process startar upp

param(
    [string]$LogPath = "C:\Temp\ProcessMonitor_$(Get-Date -Format 'yyyyMMdd_HHmmss').log"
)

# Definiera de kritiska processerna
$criticalProcesses = @(
    "winlogon",      # Windows Logon Application
    "lsass",         # Local Security Authority Subsystem Service
    "services",      # Service Control Manager
    "csrss",         # Client/Server Runtime Subsystem
    "smss",          # Session Manager Subsystem
    "LogonUI"        # Logon User Interface (temporär process)
)

# Hashtabell för att spåra processstatus
$processStatus = @{}
foreach ($proc in $criticalProcesses) {
    $processStatus[$proc] = @{
        IsRunning = $false
        StartTime = $null
        FirstSeen = $false
    }
}

# Loggningsfunktion
function Write-Log {
    param(
        [string]$Message,
        [string]$Level = "INFO"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff"
    $logMessage = "$timestamp [$Level] $Message"
    
    # Skriv till konsol med färgkodning
    switch ($Level) {
        "ERROR" { Write-Host $logMessage -ForegroundColor Red }
        "WARNING" { Write-Host $logMessage -ForegroundColor Yellow }
        "SUCCESS" { Write-Host $logMessage -ForegroundColor Green }
        "PROCESS_UP" { Write-Host $logMessage -ForegroundColor Cyan }
        default { Write-Host $logMessage }
    }
    
    # Skriv till loggfil
    $logMessage | Out-File -FilePath $LogPath -Append -Encoding UTF8
}

# Funktion för att kontrollera om en process körs
function Test-ProcessRunning {
    param([string]$ProcessName)
    
    try {
        $process = Get-Process -Name $ProcessName -ErrorAction SilentlyContinue
        return ($null -ne $process)
    }
    catch {
        return $false
    }
}

# Funktion för att hämta process-information
function Get-ProcessInfo {
    param([string]$ProcessName)
    
    try {
        $process = Get-Process -Name $ProcessName -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($process) {
            return @{
                PID = $process.Id
                StartTime = $process.StartTime
                WorkingSet = [math]::Round($process.WorkingSet64 / 1MB, 2)
            }
        }
    }
    catch {
        # StartTime kanske inte är tillgänglig för vissa systemprocesser
        if ($process) {
            return @{
                PID = $process.Id
                StartTime = "N/A"
                WorkingSet = [math]::Round($process.WorkingSet64 / 1MB, 2)
            }
        }
    }
    return $null
}

# Skriv header till loggfilen
Write-Log "================================================================" "INFO"
Write-Log "WINDOWS CRITICAL PROCESS MONITOR - STARTED" "INFO"
Write-Log "================================================================" "INFO"
Write-Log "Monitoring started at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss.fff')" "INFO"
Write-Log "Log file location: $LogPath" "INFO"
Write-Log "Computer name: $env:COMPUTERNAME" "INFO"
Write-Log "User: $env:USERNAME" "INFO"
Write-Log "OS Version: $([System.Environment]::OSVersion.VersionString)" "INFO"
Write-Log "================================================================" "INFO"
Write-Log "Processes to monitor: $($criticalProcesses -join ', ')" "INFO"
Write-Log "================================================================" "INFO"

$scriptStartTime = Get-Date
$checkInterval = 1  # Kontrollera varje sekund för snabbare detektering
$allProcessesRunning = $false
$iteration = 0

Write-Log "Starting monitoring loop..." "INFO"

while (-not $allProcessesRunning) {
    $iteration++
    $currentTime = Get-Date
    $runningCount = 0
    $notRunning = @()
    $newlyStarted = @()
    
    foreach ($processName in $criticalProcesses) {
        $isRunning = Test-ProcessRunning -ProcessName $processName
        
        if ($isRunning) {
            $runningCount++
            
            # Om processen precis har startat
            if (-not $processStatus[$processName].IsRunning) {
                $processStatus[$processName].IsRunning = $true
                $processStatus[$processName].StartTime = $currentTime
                
                # Hämta process-information
                $procInfo = Get-ProcessInfo -ProcessName $processName
                
                if (-not $processStatus[$processName].FirstSeen) {
                    $processStatus[$processName].FirstSeen = $true
                    
                    # Logga detaljerad information när processen upptäcks
                    Write-Log "================================================================" "PROCESS_UP"
                    Write-Log "PROCESS STARTED: $processName" "PROCESS_UP"
                    Write-Log "Detection time: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss.fff')" "PROCESS_UP"
                    Write-Log "Time since monitoring start: $([math]::Round(($currentTime - $scriptStartTime).TotalSeconds, 3)) seconds" "PROCESS_UP"
                    
                    if ($procInfo) {
                        Write-Log "Process ID (PID): $($procInfo.PID)" "PROCESS_UP"
                        if ($procInfo.StartTime -ne "N/A") {
                            Write-Log "Process start time: $($procInfo.StartTime)" "PROCESS_UP"
                        }
                        Write-Log "Memory usage: $($procInfo.WorkingSet) MB" "PROCESS_UP"
                    }
                    Write-Log "================================================================" "PROCESS_UP"
                    
                    $newlyStarted += $processName
                }
            }
        }
        else {
            $notRunning += $processName
            $processStatus[$processName].IsRunning = $false
        }
    }
    
    # Logga periodisk status (var 10:e iteration)
    if ($iteration % 10 -eq 0) {
        Write-Log "Status check #$iteration - Running: $runningCount/$($criticalProcesses.Count)" "INFO"
        if ($notRunning.Count -gt 0) {
            Write-Log "Still waiting for: $($notRunning -join ', ')" "WARNING"
        }
    }
    
    # Kontrollera om alla processer är igång
    if ($runningCount -eq $criticalProcesses.Count) {
        $allProcessesRunning = $true
        $totalElapsed = (Get-Date) - $scriptStartTime
        
        Write-Log "================================================================" "SUCCESS"
        Write-Log "ALL CRITICAL PROCESSES ARE RUNNING!" "SUCCESS"
        Write-Log "================================================================" "SUCCESS"
        Write-Log "Total monitoring time: $([math]::Round($totalElapsed.TotalSeconds, 3)) seconds" "SUCCESS"
        Write-Log "Total iterations: $iteration" "SUCCESS"
        
        # Sammanfattning av processstarttider
        Write-Log "================================================================" "INFO"
        Write-Log "PROCESS STARTUP SUMMARY:" "INFO"
        Write-Log "================================================================" "INFO"
        
        $sortedProcesses = $processStatus.GetEnumerator() | 
            Where-Object { $_.Value.StartTime -ne $null } |
            Sort-Object { $_.Value.StartTime }
        
        foreach ($proc in $sortedProcesses) {
            $timeSinceStart = [math]::Round(($proc.Value.StartTime - $scriptStartTime).TotalSeconds, 3)
            Write-Log "$($proc.Key): Started after $timeSinceStart seconds" "INFO"
        }
        
        Write-Log "================================================================" "INFO"
    }
    else {
        Start-Sleep -Milliseconds ($checkInterval * 1000)
    }
}

# Skapa slutrapport
Write-Log "================================================================" "SUCCESS"
Write-Log "MONITORING COMPLETED SUCCESSFULLY" "SUCCESS"
Write-Log "Script exit time: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss.fff')" "SUCCESS"
Write-Log "Exit code: 0" "SUCCESS"
Write-Log "Log file saved to: $LogPath" "SUCCESS"
Write-Log "================================================================" "SUCCESS"



exit 0


#
