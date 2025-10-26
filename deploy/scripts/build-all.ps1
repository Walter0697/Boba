# PowerShell build script for creating multi-platform binaries
# This script builds BOBA for all supported platforms on Windows

$ErrorActionPreference = "Stop"

# Get version from git tag or use dev version
try {
    $VERSION = git describe --tags --always --dirty 2>$null
    if (-not $VERSION) { $VERSION = "dev" }
} catch {
    $VERSION = "dev"
}

$BUILD_TIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

try {
    $GIT_COMMIT = git rev-parse HEAD 2>$null
    if (-not $GIT_COMMIT) { $GIT_COMMIT = "unknown" }
} catch {
    $GIT_COMMIT = "unknown"
}

Write-Host "Building BOBA v$VERSION" -ForegroundColor Green
Write-Host "Build time: $BUILD_TIME"
Write-Host "Git commit: $GIT_COMMIT"
Write-Host ""

# Create dist directory
$DIST_DIR = "dist"
if (Test-Path $DIST_DIR) {
    Remove-Item -Recurse -Force $DIST_DIR
}
New-Item -ItemType Directory -Path $DIST_DIR | Out-Null

# Build flags
$LDFLAGS = "-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"

# Platforms to build
$PLATFORMS = @{
    "linux-amd64" = @{ GOOS = "linux"; GOARCH = "amd64" }
    "linux-arm64" = @{ GOOS = "linux"; GOARCH = "arm64" }
    "darwin-amd64" = @{ GOOS = "darwin"; GOARCH = "amd64" }
    "darwin-arm64" = @{ GOOS = "darwin"; GOARCH = "arm64" }
    "windows-amd64" = @{ GOOS = "windows"; GOARCH = "amd64" }
}

# Build for each platform
foreach ($platform in $PLATFORMS.Keys) {
    $GOOS = $PLATFORMS[$platform].GOOS
    $GOARCH = $PLATFORMS[$platform].GOARCH
    
    $output_name = "boba-$platform"
    if ($GOOS -eq "windows") {
        $output_name = "$output_name.exe"
    }
    
    Write-Host "Building for $GOOS/$GOARCH..." -ForegroundColor Yellow
    
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:CGO_ENABLED = "0"
    
    try {
        go build -ldflags="$LDFLAGS" -o "$DIST_DIR\$output_name" .
        Write-Host "✓ Built $output_name" -ForegroundColor Green
    } catch {
        Write-Host "Failed to build for $GOOS/$GOARCH" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "Creating archives..." -ForegroundColor Green

# Create archives
Push-Location $DIST_DIR

Get-ChildItem "boba-*" | ForEach-Object {
    if ($_.Name -like "*.exe") {
        # Windows: create zip
        $zip_name = $_.Name -replace '\.exe$', '.zip'
        Compress-Archive -Path $_.Name -DestinationPath $zip_name -Force
        Write-Host "✓ Created $zip_name" -ForegroundColor Green
    } else {
        # Unix: create tar.gz (requires tar command)
        $tar_name = "$($_.Name).tar.gz"
        if (Get-Command tar -ErrorAction SilentlyContinue) {
            tar -czf $tar_name $_.Name
            Write-Host "✓ Created $tar_name" -ForegroundColor Green
        } else {
            Write-Host "⚠ tar command not found, skipping $tar_name" -ForegroundColor Yellow
        }
    }
}

# Generate checksums
Write-Host ""
Write-Host "Generating checksums..." -ForegroundColor Green

$checksums = @()
Get-ChildItem "*.tar.gz", "*.zip" -ErrorAction SilentlyContinue | ForEach-Object {
    $hash = (Get-FileHash -Path $_.Name -Algorithm SHA256).Hash.ToLower()
    $checksums += "$hash  $($_.Name)"
}

$checksums | Out-File -FilePath "checksums.txt" -Encoding ASCII
Write-Host "✓ Created checksums.txt" -ForegroundColor Green

Pop-Location

Write-Host ""
Write-Host "Build complete!" -ForegroundColor Green
Write-Host "Artifacts are in the $DIST_DIR\ directory:"
Get-ChildItem $DIST_DIR | Format-Table Name, Length, LastWriteTime

Write-Host ""
Write-Host "To test a binary:" -ForegroundColor Yellow
Write-Host "  .\$DIST_DIR\boba-windows-amd64.exe --version"
Write-Host ""
Write-Host "To create a release:" -ForegroundColor Yellow
Write-Host "  git tag -a v$VERSION -m 'Release v$VERSION'"
Write-Host "  git push origin v$VERSION"
