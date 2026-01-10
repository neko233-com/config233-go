# Ensure we're in the script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "Config233-Go Auto Release Script" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green

# Check git status first
Write-Host "Checking git status..." -ForegroundColor Yellow
$gitStatus = git status --porcelain
if ($LASTEXITCODE -ne 0) {
    Write-Error "Git command failed"
    exit 1
}
if ($gitStatus) {
    Write-Error "Working directory is not clean. Please commit or stash changes."
    Write-Host $gitStatus
    exit 1
}

# Run tests
Write-Host "Running tests..." -ForegroundColor Yellow
go test ./tests
if ($LASTEXITCODE -ne 0) {
    Write-Error "Tests failed"
    exit 1
}

# Build
Write-Host "Building..." -ForegroundColor Yellow
go build ./pkg/config233
if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed"
    exit 1
}

# Now update version after all checks pass
# Read current version from version.txt
$versionFile = "version.txt"
$currentVersion = Get-Content $versionFile -Raw
$currentVersion = $currentVersion.Trim()

Write-Host "Current version: $currentVersion"

# Parse version (assume vX.Y.Z)
$versionPattern = '^v(\d+)\.(\d+)\.(\d+)$'
if ($currentVersion -notmatch $versionPattern) {
    Write-Error "Invalid version format in version.txt. Expected vX.Y.Z"
    exit 1
}

$major = [int]$matches[1]
$minor = [int]$matches[2]
$patch = [int]$matches[3]

# Increment patch version
$patch++
$newVersion = "v$major.$minor.$patch"

Write-Host "New version: $newVersion"

# Update version.txt
$newVersion | Out-File $versionFile -Encoding UTF8

Write-Host "Updated version.txt to $newVersion"
Write-Host ""

# Use the new version
$Version = $newVersion

Write-Host "Releasing version $Version"
Write-Host ""


# Create git tag
Write-Host "Creating git tag $Version..." -ForegroundColor Yellow
git tag -a $Version -m "Release $Version"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to create git tag"
    exit 1
}

# Push tag
Write-Host "Pushing tag to remote..." -ForegroundColor Yellow
git push origin $Version
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to push tag"
    exit 1
}

# Push main branch
Write-Host "Pushing main branch..." -ForegroundColor Yellow
git push origin main
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to push main branch"
    exit 1
}

Write-Host ""
Write-Host "Release $Version completed successfully!" -ForegroundColor Green
Write-Host "The module will be available at: https://pkg.go.dev/github.com/neko233-com/config233-go@$Version"
Write-Host ""