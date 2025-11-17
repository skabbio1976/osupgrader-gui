# Script för att skapa scheduled task som signalerar när systemet är redo
# Använder schtasks.exe istället för PowerShell cmdlets för att undvika UAC/elevation-problem

$ErrorActionPreference = 'Stop'

if(!(test-path 'C:\Temp')){
    New-Item -ItemType Directory -Path "C:\Temp" -Force
}

$TaskSignalFile = 'C:\Temp\osupgrader_ready.txt'

# Ta bort gammal signal-fil om den finns
if((test-path $TaskSignalFile)){
    Remove-Item -Path $TaskSignalFile -Force
}

try {
    # === SCHEDULED TASK via schtasks.exe (fungerar utan UAC elevation) ===
    Write-Output "Creating scheduled task XML..."

    # Skapa task XML
    $taskXml = @"
<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Description>OS Upgrader Signal Task</Description>
  </RegistrationInfo>
  <Triggers>
    <BootTrigger>
      <Delay>PT6M</Delay>
      <Enabled>true</Enabled>
    </BootTrigger>
  </Triggers>
  <Principals>
    <Principal id="Author">
      <UserId>S-1-5-18</UserId>
      <RunLevel>HighestAvailable</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <AllowHardTerminate>true</AllowHardTerminate>
    <StartWhenAvailable>true</StartWhenAvailable>
    <RunOnlyIfNetworkAvailable>false</RunOnlyIfNetworkAvailable>
    <AllowStartOnDemand>true</AllowStartOnDemand>
    <Enabled>true</Enabled>
    <Hidden>false</Hidden>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>powershell.exe</Command>
      <Arguments>-NoProfile -WindowStyle Hidden -Command "New-Item -Path '$TaskSignalFile' -ItemType File -Force; schtasks.exe /Delete /TN 'OSUpgraderSignal' /F"</Arguments>
    </Exec>
  </Actions>
</Task>
"@

    # Spara XML till temp-fil
    $xmlPath = "C:\Temp\osupgrader_task.xml"
    $taskXml | Out-File -FilePath $xmlPath -Encoding unicode -Force

    Write-Output "Registering scheduled task with schtasks.exe..."

    # Använd schtasks.exe för att skapa tasken (fungerar utan UAC)
    $result = schtasks.exe /Create /TN "OSUpgraderSignal" /XML $xmlPath /F 2>&1
    $exitCode = $LASTEXITCODE

    Write-Output "schtasks.exe exit code: $exitCode"
    Write-Output "schtasks.exe output: $result"

    if ($exitCode -ne 0) {
        Write-Output "ERROR: Scheduled task creation failed with exit code $exitCode"
        if (Test-Path $xmlPath) {
            Remove-Item $xmlPath -Force
        }
        exit $exitCode
    } else {
        Write-Output "Scheduled task created successfully"
    }

    $verify = schtasks.exe /Query /TN "OSUpgraderSignal" /FO LIST /V 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Output "ERROR: Scheduled task verification failed: $verify"
        if (Test-Path $xmlPath) {
            Remove-Item $xmlPath -Force
        }
        exit $LASTEXITCODE
    }

    # Ta bort temp XML-fil
    if (Test-Path $xmlPath) {
        Remove-Item $xmlPath -Force
    }

    exit 0
} catch {
    Write-Output "ERROR: $($_.Exception.Message)"
    Write-Output "Stack trace: $($_.ScriptStackTrace)"
    exit 1
}
