# Test npm packaging for Electron app
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  ClawPet Electron - npm Package Test" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

# 1. Clean old files
Write-Host "[1/6] Cleaning old files..." -ForegroundColor Yellow
if (Test-Path "dist") {
    Remove-Item -Recurse -Force "dist"
    Write-Host "  OK - Deleted old dist directory" -ForegroundColor Green
}

if (Test-Path "node_modules") {
    Write-Host "  Removing old node_modules (this may take a while)..." -ForegroundColor Gray
    Remove-Item -Recurse -Force "node_modules"
    Write-Host "  OK - Deleted old node_modules" -ForegroundColor Green
}

# 2. Install dependencies with npm
Write-Host "`n[2/6] Installing dependencies with npm..." -ForegroundColor Yellow
npm install
Write-Host "  OK - npm install completed" -ForegroundColor Green

# 3. Build Next.js
Write-Host "`n[3/6] Building Next.js..." -ForegroundColor Yellow
npm run build
Write-Host "  OK - Next.js build completed" -ForegroundColor Green

# 4. Copy Go binaries
Write-Host "`n[4/6] Copying Go binaries..." -ForegroundColor Yellow
if (Test-Path "..\..\dist\picoclaw.exe") {
    Copy-Item "..\..\dist\picoclaw.exe" "picoclaw.exe" -Force
    Write-Host "  OK - Copied picoclaw.exe" -ForegroundColor Green
} else {
    Write-Host "  WARN - picoclaw.exe not found in ..\..\dist" -ForegroundColor Magenta
}

if (Test-Path "..\..\dist\picoclaw-web.exe") {
    Copy-Item "..\..\dist\picoclaw-web.exe" "picoclaw-web.exe" -Force
    Write-Host "  OK - Copied picoclaw-web.exe" -ForegroundColor Green
} else {
    Write-Host "  WARN - picoclaw-web.exe not found in ..\..\dist" -ForegroundColor Magenta
}

# 5. Package Electron app
Write-Host "`n[5/6] Packaging Electron app..." -ForegroundColor Yellow
npm run package
Write-Host "  OK - Electron packaging completed" -ForegroundColor Green

# 6. Verify package
Write-Host "`n[6/6] Verifying package..." -ForegroundColor Yellow

$zipFile = Get-ChildItem "dist\*.zip" -ErrorAction SilentlyContinue | Select-Object -First 1
if ($zipFile) {
    Write-Host "  OK - Found package: $($zipFile.Name)" -ForegroundColor Green
    Write-Host "  Package size: $([math]::Round($zipFile.Length / 1MB, 2)) MB" -ForegroundColor Cyan
    
    # Extract for testing
    $testDir = "dist\test-run"
    if (Test-Path $testDir) {
        Remove-Item -Recurse -Force $testDir
    }
    Expand-Archive -Path $zipFile.FullName -DestinationPath $testDir
    
    Write-Host "`n  Verifying critical files:" -ForegroundColor Cyan
    
    # Check .next directory
    if (Test-Path "dist\test-run\resources\app.asar.unpacked\.next") {
        Write-Host "    OK - .next directory" -ForegroundColor Green
    } else {
        Write-Host "    FAIL - .next directory not found" -ForegroundColor Red
    }
    
    # Check Next.js CLI
    if (Test-Path "dist\test-run\resources\app.asar.unpacked\node_modules\next\dist\bin\next") {
        Write-Host "    OK - Next.js CLI" -ForegroundColor Green
    } else {
        Write-Host "    FAIL - Next.js CLI not found" -ForegroundColor Red
    }
    
    # Check main.js
    if (Test-Path "dist\test-run\resources\app.asar.unpacked\electron\main.js") {
        Write-Host "    OK - main.js" -ForegroundColor Green
    } else {
        Write-Host "    FAIL - main.js not found" -ForegroundColor Red
    }
    
    # Check picoclaw.exe
    if (Test-Path "dist\test-run\picoclaw.exe") {
        Write-Host "    OK - picoclaw.exe" -ForegroundColor Green
    } else {
        Write-Host "    FAIL - picoclaw.exe not found" -ForegroundColor Red
    }
    
    Write-Host "`n  Test complete. You can now run the app to test." -ForegroundColor Cyan
} else {
    Write-Host "  FAIL - No package file found" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Test Complete" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan
