param(
    [string]$MSVCPath
)

#Requires -Version 5.1
$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force -Path dist | Out-Null

Write-Host "Building win-powerctl..."
go build -o dist/win-powerctl.exe ./cmd/win-powerctl

Write-Host "Building poweroff.dll..."
& "$PSScriptRoot\poweroff\build.ps1" -MSVCPath $MSVCPath

Move-Item poweroff/poweroff.dll dist/ -Force
Copy-Item config.ini dist/

Write-Host "Build complete: dist\"
Get-ChildItem dist/ -Name
