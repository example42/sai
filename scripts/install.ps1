# SAI CLI Installation Script for Windows PowerShell

param(
    [string]$Version = "",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\sai",
    [switch]$Help,
    [switch]$AddToPath
)

# Configuration
$RepoOwner = "example42"
$RepoName = "sai"
$BinaryName = "sai.exe"

# Show help
if ($Help) {
    Write-Host "SAI CLI Installation Script for Windows"
    Write-Host ""
    Write-Host "Usage: .\install.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Version <version>    Install specific version (default: latest)"
    Write-Host "  -InstallDir <path>    Installation directory (default: $env:LOCALAPPDATA\Programs\sai)"
    Write-Host "  -AddToPath           Add installation directory to PATH"
    Write-Host "  -Help                Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\install.ps1                           # Install latest version"
    Write-Host "  .\install.ps1 -Version v1.0.0          # Install specific version"
    Write-Host "  .\install.ps1 -AddToPath               # Install and add to PATH"
    exit 0
}

# Logging functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { 
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Get latest release version from GitHub
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to get latest version from GitHub: $_"
        exit 1
    }
}

# Download and install SAI
function Install-Sai {
    param(
        [string]$Platform,
        [string]$Version
    )
    
    Write-Info "Installing SAI CLI version $Version for $Platform..."
    
    # Create installation directory
    if (!(Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    # Download URL
    $downloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/sai-$Version-$Platform.zip"
    $tempFile = [System.IO.Path]::GetTempFileName() + ".zip"
    $tempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
    
    try {
        # Download release
        Write-Info "Downloading from $downloadUrl..."
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
        
        # Extract archive
        Write-Info "Extracting archive..."
        Expand-Archive -Path $tempFile -DestinationPath $tempDir -Force
        
        # Find and copy binary
        $binaryPath = Get-ChildItem -Path $tempDir -Name $BinaryName -Recurse | Select-Object -First 1
        if (!$binaryPath) {
            Write-Error "Binary not found in archive"
            exit 1
        }
        
        $sourcePath = Join-Path $tempDir $binaryPath.FullName
        $destPath = Join-Path $InstallDir $BinaryName
        
        Write-Info "Installing to $destPath..."
        Copy-Item -Path $sourcePath -Destination $destPath -Force
        
        Write-Success "SAI CLI installed successfully!"
    }
    catch {
        Write-Error "Installation failed: $_"
        exit 1
    }
    finally {
        # Cleanup
        if (Test-Path $tempFile) { Remove-Item $tempFile -Force }
        if (Test-Path $tempDir) { Remove-Item $tempDir -Recurse -Force }
    }
}

# Add to PATH
function Add-ToPath {
    param([string]$Directory)
    
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$Directory*") {
        Write-Info "Adding $Directory to user PATH..."
        $newPath = "$currentPath;$Directory"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Success "Added to PATH. Restart your terminal to use 'sai' command."
    } else {
        Write-Info "Directory already in PATH"
    }
}

# Verify installation
function Test-Installation {
    $binaryPath = Join-Path $InstallDir $BinaryName
    
    if (Test-Path $binaryPath) {
        Write-Success "SAI CLI is installed at $binaryPath"
        
        # Test if in PATH
        try {
            $null = Get-Command sai -ErrorAction Stop
            Write-Success "SAI CLI is available in PATH"
            $version = & sai version --short 2>$null
            Write-Info "Installed version: $version"
            Write-Info "Run 'sai --help' to get started"
        }
        catch {
            Write-Warning "SAI CLI installed but not found in PATH"
            Write-Info "Run directly: $binaryPath"
            Write-Info "Or add to PATH with: .\install.ps1 -AddToPath"
        }
    } else {
        Write-Error "Installation verification failed"
        exit 1
    }
}

# Main installation flow
function Main {
    Write-Info "SAI CLI Installation Script for Windows"
    Write-Info "======================================"
    
    # Check if already installed
    $existingPath = Join-Path $InstallDir $BinaryName
    if (Test-Path $existingPath) {
        try {
            $currentVersion = & $existingPath version --short 2>$null
            Write-Warning "SAI CLI is already installed (version: $currentVersion)"
        }
        catch {
            Write-Warning "SAI CLI is already installed (version: unknown)"
        }
        
        $response = Read-Host "Do you want to reinstall? [y/N]"
        if ($response -notmatch "^[Yy]$") {
            Write-Info "Installation cancelled"
            exit 0
        }
    }
    
    # Detect platform
    $arch = Get-Architecture
    $platform = "windows-$arch"
    Write-Info "Detected platform: $platform"
    
    # Get version
    if ($Version) {
        Write-Info "Using specified version: $Version"
    } else {
        Write-Info "Fetching latest release version..."
        $Version = Get-LatestVersion
        Write-Info "Latest version: $Version"
    }
    
    # Install SAI
    Install-Sai -Platform $platform -Version $Version
    
    # Add to PATH if requested
    if ($AddToPath) {
        Add-ToPath -Directory $InstallDir
    }
    
    # Verify installation
    Test-Installation
}

# Run main function
Main