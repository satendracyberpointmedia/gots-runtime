# GoTS Runtime - Windows Build Script
# Safe version (no here-strings, no Unicode, no encoding issues)

param(
    [string]$Version = "0.1.0",
    [switch]$MSI,
    [switch]$NSIS,
    [switch]$Portable,
    [switch]$All
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "GoTS Runtime - Windows Build Script" -ForegroundColor Cyan
Write-Host "Version: $Version" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Directories
$BUILD_DIR = "build\windows"
$DIST_DIR  = "dist\windows"

Write-Host ""
Write-Host "Creating build directories..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $BUILD_DIR | Out-Null
New-Item -ItemType Directory -Force -Path $DIST_DIR  | Out-Null

# -----------------------------
# Build Go executable
# -----------------------------
Write-Host ""
Write-Host "Building GoTS Runtime executable..." -ForegroundColor Yellow

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

if (-not (Test-Path "cmd\gots")) {
    Write-Host "ERROR: cmd\gots directory not found" -ForegroundColor Red
    Write-Host "Expected: cmd\gots\main.go" -ForegroundColor Yellow
    exit 1
}

go build -o "$BUILD_DIR\gots.exe" -ldflags "-s -w -X main.Version=$Version" .\cmd\gots

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go build failed" -ForegroundColor Red
    exit 1
}

Write-Host "[OK] Executable built: $BUILD_DIR\gots.exe" -ForegroundColor Green

# File size
$exeSize = (Get-Item "$BUILD_DIR\gots.exe").Length / 1MB
Write-Host "Size: $([math]::Round($exeSize, 2)) MB" -ForegroundColor Gray

# -----------------------------
# Copy standard library
# -----------------------------
Write-Host ""
Write-Host "Copying standard library..." -ForegroundColor Yellow

if (Test-Path "stdlib") {
    Copy-Item -Path "stdlib" -Destination "$BUILD_DIR\stdlib" -Recurse -Force
    Write-Host "[OK] stdlib copied" -ForegroundColor Green
} else {
    New-Item -ItemType Directory -Force -Path "$BUILD_DIR\stdlib" | Out-Null
    Write-Host "[WARN] stdlib not found, empty folder created" -ForegroundColor Yellow
}

# -----------------------------
# Copy documentation
# -----------------------------
Write-Host ""
Write-Host "Copying documentation..." -ForegroundColor Yellow

if (Test-Path "README.md") {
    Copy-Item "README.md" "$BUILD_DIR\" -Force
}
if (Test-Path "LICENSE") {
    Copy-Item "LICENSE" "$BUILD_DIR\" -Force
}

Write-Host "[OK] Documentation copied" -ForegroundColor Green

# -----------------------------
# Portable ZIP builder
# -----------------------------
function Build-Portable {
    Write-Host ""
    Write-Host "Creating Portable ZIP package..." -ForegroundColor Cyan

    $zipName = "gots-runtime-$Version-windows-amd64-portable.zip"
    $zipPath = "$DIST_DIR\$zipName"

    if (Test-Path $zipPath) {
        Remove-Item $zipPath -Force
    }

    $readmeLines = @(
        "GoTS Runtime v$Version",
        "",
        "Portable Edition for Windows (x64)",
        "",
        "Installation:",
        "1. Extract this ZIP to any folder",
        "2. (Optional) Add folder to PATH",
        "3. Run gots.exe",
        "",
        "Notes:",
        "- stdlib folder must be next to gots.exe",
        "- OR set GOTS_STDLIB_PATH environment variable",
        "",
        "Quick start:",
        "gots --help",
        "",
        "See README.md for more information"
    )

    $readmeLines | Out-File "$BUILD_DIR\PORTABLE_README.txt" -Encoding UTF8

    Compress-Archive -Path "$BUILD_DIR\*" -DestinationPath $zipPath -Force

    $zipSize = (Get-Item $zipPath).Length / 1MB
    Write-Host "[OK] Portable ZIP created: $zipName" -ForegroundColor Green
    Write-Host "Size: $([math]::Round($zipSize, 2)) MB" -ForegroundColor Gray
}

# -----------------------------
# Execute based on parameters
# -----------------------------
if ($All) {
    Build-Portable
} else {
    if ($Portable) {
        Build-Portable
    }

    if (-not ($Portable -or $MSI -or $NSIS)) {
        Build-Portable
    }
}

# -----------------------------
# Summary
# -----------------------------
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build directory: $BUILD_DIR" -ForegroundColor White
Write-Host "Distribution directory: $DIST_DIR" -ForegroundColor White
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Test executable:" -ForegroundColor Gray
Write-Host "   .\$BUILD_DIR\gots.exe --version" -ForegroundColor Gray
Write-Host "2. Extract and test portable ZIP" -ForegroundColor Gray
Write-Host ""
Write-Host "[OK] Build completed successfully" -ForegroundColor Green
