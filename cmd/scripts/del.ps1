
# Define the executable and base directory
$exePath = ".\static.exe"

# Setup the project
Write-Host "Deleting up the project..."
rm -r *

# Define the output directory
$outputDir = ".\output"

# Check if the output directory exists and delete its contents
if (Test-Path $outputDir) {
    Write-Host "Deleting contents of the output directory..."
    rm -r "$outputDir\*"
}