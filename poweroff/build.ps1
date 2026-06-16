param(
    [string]$MSVCPath
)

#Requires -Version 5.1
$ErrorActionPreference = "Stop"

$FixedPath = "C:\Program Files (x86)\Microsoft Visual Studio\18\BuildTools\VC\Auxiliary\Build\vcvarsall.bat"

Write-Host "Building poweroff.dll..."

$vcvars = if ($MSVCPath) { $MSVCPath } else { $FixedPath }

if (-not (Test-Path $vcvars)) {
    Write-Error "MSVC not found: $vcvars"
    exit 1
}

Write-Host "Building with MSVC: $vcvars"
cmd /c "`"$vcvars`" x64 && cd /d $PSScriptRoot && cl /LD poweroff.c /link /OUT:poweroff.dll advapi32.lib user32.lib"

Remove-Item $PSScriptRoot\*.obj, $PSScriptRoot\*.lib, $PSScriptRoot\*.exp -ErrorAction SilentlyContinue

Write-Host "Build complete: poweroff.dll"
