# GitHub Push Script - Fixed version (no Chinese encoding issues)
# Usage: .\push_to_github.ps1 -Token "your GitHub Token"

param(
    [Parameter(Mandatory=$true)]
    [string]$Token
)

Write-Host "Pushing code to GitHub..." -ForegroundColor Green
Set-Location "D:\gohome\demo1-gozero\ns-tracking-go"

# Configure Git with token
$repoUrl = "https://Kirito-al:$Token@github.com/Kirito-al/ns-tracking-go.git"
git remote set-url origin $repoUrl

try {
    git push -u origin main
    if ($LASTEXITCODE -eq 0) {
        Write-Host "SUCCESS! Code pushed to GitHub." -ForegroundColor Green
        Write-Host "Repository: https://github.com/Kirito-al/ns-tracking-go" -ForegroundColor Cyan
    }
} catch {
    Write-Host "Push failed. Please check:" -ForegroundColor Red
    Write-Host "1. GitHub repository exists" -ForegroundColor Yellow
    Write-Host "2. Token has 'repo' permission" -ForegroundColor Yellow
}

# Clean up token from remote URL
git remote set-url origin https://github.com/Kirito-al/ns-tracking-go.git
Write-Host "Done." -ForegroundColor Green