# Define the executable and base directory
$exePath = ".\static.exe"

# Step 1: Setup the project
Write-Host "Setting up the project..." -ForegroundColor Green
& $exePath setup "1/1"
Write-Host "Project setup completed." -ForegroundColor Green

# Display the current directory contents
Write-Host "`nProject directory after setup:" -ForegroundColor Cyan
ls | Format-Table -AutoSize

# Step 2: Add pages to the project
Write-Host "`nAdding pages to the project..." -ForegroundColor Green
$pages = @("home", "about", "contact")
foreach ($page in $pages) {
    Write-Host "Adding page '$page'..." -ForegroundColor Yellow
    & $exePath add $page
    Write-Host "Page '$page' added successfully." -ForegroundColor Green
}

# Display the directory structure
Write-Host "`nProject directory structure after adding pages:" -ForegroundColor Cyan
tree /f | Out-String | Write-Host

# Step 3: Compile the project
Write-Host "`nCompiling the project..." -ForegroundColor Green
& $exePath compile
Write-Host "Project compiled successfully." -ForegroundColor Green

# Step 4: Serve the compiled output using Python HTTP server
Write-Host "`nStarting HTTP server to serve the compiled project..." -ForegroundColor Cyan
# Start-Process -NoNewWindow -FilePath "python" -ArgumentList "-m http.server --directory $outputDir 8000"

# Open the browser to display the served project
Start-Sleep -Seconds 2  # Allow server to start
Start-Process "http://localhost:8000"
& $exePath watch
Write-Host "`nDemo completed. Visit http://localhost:8000 to view the project." -ForegroundColor Green
