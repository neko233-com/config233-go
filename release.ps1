Write-Host "Config233-Go Auto Release Script" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green

# Prompt for version
$Version = Read-Host "Enter version tag (e.g., v1.0.0)"
if ([string]::IsNullOrWhiteSpace($Version)) {
    Write-Error "Version cannot be empty"
    exit 1
}

Write-Host "Releasing version $Version"
Write-Host ""

# Check git status
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
Write-Host "The module will be available at: https://pkg.go.dev/config233-go@$Version"
Write-Host ""