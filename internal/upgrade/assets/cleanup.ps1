if((test-path "C:\Temp\osupgrader_ready.txt")){
   # Remove-Item "C:\Temp\osupgrader_ready.txt" -Force -ErrorAction SilentlyContinue
}

# Remove scheduled task
schtasks.exe /Delete /TN "OSUpgraderSignal" /F

# Remove uploaded files
if((test-path "C:\Temp\createsignaltasks.ps1")){
   Remove-Item "C:\Temp\createsignaltasks.ps1" -Force -ErrorAction SilentlyContinue
}
if((test-path "C:\Temp\processmonitor.ps1")){
   Remove-Item "C:\Temp\processmonitor.ps1" -Force -ErrorAction SilentlyContinue
}

