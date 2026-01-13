# GoTS Runtime - Quick Setup Script

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "GoTS Runtime - Quick Setup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "Creating directory structure..." -ForegroundColor Yellow

$directories = @(
    "cmd\gots",
    "stdlib",
    "assets",
    "build",
    "dist"
)

foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "  [OK] Created: $dir" -ForegroundColor Green
    } else {
        Write-Host "  [WARN] Already exists: $dir" -ForegroundColor Gray
    }
}

# Check main.go
if (Test-Path "cmd\gots\main.go") {
    Write-Host ""
    Write-Host "[OK] cmd\gots\main.go already exists" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "[WARN] cmd\gots\main.go not found" -ForegroundColor Yellow
}

# stdlib README
if (-not (Test-Path "stdlib\README.md")) {
    Write-Host ""
    Write-Host "Creating stdlib README..." -ForegroundColor Yellow

    $stdlibReadme = @(
        "# GoTS Standard Library",
        "",
        "This directory contains the TypeScript standard library for the GoTS runtime.",
        "",
        "## Structure (Planned)",
        "",
        "stdlib/",
        "  core/   - Core runtime functions",
        "  http/   - HTTP client/server",
        "  fs/     - File system operations",
        "  crypto/ - Cryptography",
        "  util/   - Utilities",
        "",
        "## Status",
        "",
        "This is a placeholder. The actual standard library will be implemented later.",
        "This directory must exist for the runtime to start correctly."
    )

    $stdlibReadme | Out-File -FilePath "stdlib\README.md" -Encoding UTF8
    Write-Host "  [OK] Created: stdlib\README.md" -ForegroundColor Green
}

# .gitignore
if (-not (Test-Path ".gitignore")) {
    Write-Host ""
    Write-Host "Creating .gitignore..." -ForegroundColor Yellow

    $gitignoreContent = @(
        "*.exe",
        "*.dll",
        "*.so",
        "*.dylib",
        "gots",
        "",
        "build/",
        "dist/",
        "",
        "*.test",
        "*.out",
        "coverage.out",
        "",
        ".vscode/",
        ".idea/",
        "*.swp",
        "*.swo",
        "*~",
        "",
        ".DS_Store",
        "Thumbs.db",
        "",
        "*.tmp",
        "*.log"
    )

    $gitignoreContent | Out-File -FilePath ".gitignore" -Encoding UTF8
    Write-Host "  [OK] Created: .gitignore" -ForegroundColor Green
}

# LICENSE
if (-not (Test-Path "LICENSE")) {
    Write-Host ""
    Write-Host "Creating LICENSE..." -ForegroundColor Yellow

    $licenseContent = @(
        "MIT License",
        "",
        "Copyright (c) 2025 GoTS Team",
        "",
        "Permission is hereby granted, free of charge, to any person obtaining a copy",
        "of this software and associated documentation files to deal in the Software",
        "without restriction, including without limitation the rights to use, copy,",
        "modify, merge, publish, distribute, sublicense, and/or sell copies.",
        "",
        "THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND."
    )

    $licenseContent | Out-File -FilePath "LICENSE" -Encoding UTF8
    Write-Host "  [OK] Created: LICENSE" -ForegroundColor Green
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
