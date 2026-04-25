# Simple setup script for GoondVR channels configuration (PowerShell)

Write-Host "🎬 GoondVR Channel Setup" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan
Write-Host ""

# Create conf directory if it doesn't exist
if (-not (Test-Path "conf")) {
    Write-Host "📁 Creating conf directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path "conf" -Force | Out-Null
}

# Check if channels.json already exists
if (Test-Path "conf/channels.json") {
    Write-Host "⚠️  conf/channels.json already exists!" -ForegroundColor Yellow
    Write-Host ""
    $response = Read-Host "Do you want to overwrite it? (y/N)"
    if ($response -notmatch "^[Yy]$") {
        Write-Host "❌ Setup cancelled. Your existing channels.json was not modified." -ForegroundColor Red
        exit 0
    }
}

# Copy example file
if (Test-Path "channels.json.example") {
    Write-Host "📋 Copying channels.json.example to conf/channels.json..." -ForegroundColor Yellow
    Copy-Item "channels.json.example" "conf/channels.json" -Force
    Write-Host "✅ Done!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📝 Next steps:" -ForegroundColor Cyan
    Write-Host "   1. Edit conf/channels.json and add your channels"
    Write-Host "   2. Run ./goondvr.exe to start recording"
    Write-Host ""
    Write-Host "📖 For detailed documentation, see CHANNELS.md" -ForegroundColor Cyan
} else {
    Write-Host "❌ Error: channels.json.example not found!" -ForegroundColor Red
    Write-Host "   Make sure you're running this script from the GoondVR directory."
    exit 1
}
