param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("green", "installer", "installer-full")]
    [string]$Mode
)

$ErrorActionPreference = "Stop"
$ProjectRoot = $PSScriptRoot

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  OBS Client Build Tool" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

function Find-InnoSetup {
    $paths = @(
        "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe",
        "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
        "C:\Program Files\Inno Setup 6\ISCC.exe",
        "$env:LOCALAPPDATA\Programs\Inno Setup\ISCC.exe",
        "C:\Program Files (x86)\Inno Setup\ISCC.exe",
        "C:\Program Files\Inno Setup\ISCC.exe"
    )
    
    foreach ($path in $paths) {
        if (Test-Path $path) {
            return $path
        }
    }
    
    return $null
}

switch ($Mode) {
    "green" {
        Write-Host ""
        Write-Host "Building green version..." -ForegroundColor Green
        
        Write-Host "1. Building with wails..." -ForegroundColor Gray
        wails build
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Wails build failed!" -ForegroundColor Red
            exit 1
        }
        
        Write-Host ""
        Write-Host "Green version build succeeded!" -ForegroundColor Green
        Write-Host "Output: $ProjectRoot\build\bin\standalone\obs-client.exe" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Usage: Double-click obs-client.exe to run" -ForegroundColor Yellow
    }
    
    "installer" {
        Write-Host ""
        Write-Host "Building installer version..." -ForegroundColor Green
        
        Write-Host "1. Building with wails..." -ForegroundColor Gray
        wails build
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Wails build failed!" -ForegroundColor Red
            exit 1
        }
        
        Write-Host "2. Checking Inno Setup..." -ForegroundColor Gray
        $innoSetupPath = Find-InnoSetup
        if (-not $innoSetupPath) {
            Write-Host "Inno Setup not found!" -ForegroundColor Yellow
            Write-Host "Installing via winget..." -ForegroundColor Gray
            
            try {
                winget install --id JRSoftware.InnoSetup -e -s winget
            } catch {
                Write-Host "Winget install failed!" -ForegroundColor Red
                Write-Host "Please install Inno Setup manually: https://jrsoftware.org/isdl.php" -ForegroundColor Red
                exit 1
            }
            
            $innoSetupPath = Find-InnoSetup
            if (-not $innoSetupPath) {
                Write-Host "Inno Setup installation failed!" -ForegroundColor Red
                exit 1
            }
        }
        
        Write-Host "Found: $innoSetupPath" -ForegroundColor Gray
        
        Write-Host "3. Generating installer..." -ForegroundColor Gray
        & $innoSetupPath "$ProjectRoot\build\windows\installer\installer.iss"
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Installer generation failed!" -ForegroundColor Red
            exit 1
        }
        
        Write-Host ""
        Write-Host "Installer version build succeeded!" -ForegroundColor Green
        Write-Host "Output: $ProjectRoot\build\bin\obs-client-setup-1.0.0.exe" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Usage: Double-click installer and follow the wizard" -ForegroundColor Yellow
    }
    
    "installer-full" {
        Write-Host ""
        Write-Host "Building installer version with WebView2..." -ForegroundColor Green
        
        Write-Host "1. Building with wails..." -ForegroundColor Gray
        wails build
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Wails build failed!" -ForegroundColor Red
            exit 1
        }
        
        Write-Host "2. Checking Inno Setup..." -ForegroundColor Gray
        $innoSetupPath = Find-InnoSetup
        if (-not $innoSetupPath) {
            Write-Host "Inno Setup not found!" -ForegroundColor Yellow
            Write-Host "Installing via winget..." -ForegroundColor Gray
            
            try {
                winget install --id JRSoftware.InnoSetup -e -s winget
            } catch {
                Write-Host "Winget install failed!" -ForegroundColor Red
                Write-Host "Please install Inno Setup manually: https://jrsoftware.org/isdl.php" -ForegroundColor Red
                exit 1
            }
            
            $innoSetupPath = Find-InnoSetup
            if (-not $innoSetupPath) {
                Write-Host "Inno Setup installation failed!" -ForegroundColor Red
                exit 1
            }
        }
        
        Write-Host "Found: $innoSetupPath" -ForegroundColor Gray
        
        Write-Host "3. Generating installer with WebView2..." -ForegroundColor Gray
        & $innoSetupPath "$ProjectRoot\build\windows\installer\installer-full.iss"
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Installer generation failed!" -ForegroundColor Red
            exit 1
        }
        
        Write-Host ""
        Write-Host "Full installer version build succeeded!" -ForegroundColor Green
        Write-Host "Output: $ProjectRoot\build\bin\obs-client-setup-full-1.0.0.exe" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Usage: Double-click installer and follow the wizard" -ForegroundColor Yellow
        Write-Host "This installer will automatically install WebView2 Runtime if not present" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan