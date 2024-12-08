
# Define the executable and base directory
$exePath = ".\static.exe"

# Setup the project
Write-Host "Setting up the project..."
& $exePath setup "1/1"
Write-Host "Project setup completed."

# Add pages to the project
$pages = @("home", "about", "contact", "knowmore", "sheshells", "sheshells2", "sheshells3", "sheshells4", "sheshells5", "sheshells6", "sheshells7", "sheshells8", "sheshells9", "sheshells10", "sheshells11", "sheshells12", "sheshells13", "sheshells14", "sheshells15", "sheshells16", "sheshells17", "sheshells18", "sheshells19", "sheshells20", "sheshells21", "sheshells22", "sheshells23", "sheshells24", "sheshells25", "sheshells26", "sheshells27", "sheshells28", "sheshells29", "sheshells30", "sheshells31", "sheshells32", "sheshells33", "sheshells34", "sheshells35", "sheshells36", "sheshells37", "sheshells38", "sheshells39", "sheshells40", "sheshells41", "sheshells42", "sheshells43", "sheshells44", "sheshells45", "sheshells46", "sheshells47", "sheshells48", "sheshells49", "sheshells50", "sheshells51", "sheshells52", "sheshells53", "sheshells54", "sheshells55", "sheshells56", "sheshells57", "sheshells58", "sheshells59", "sheshells60", "sheshells61", "sheshells62", "sheshells63", "sheshells64", "sheshells65", "sheshells66", "sheshells67", "sheshells68", "sheshells69", "sheshells70", "sheshells71", "sheshells72", "sheshells73", "sheshells74", "sheshells75", "sheshells76", "sheshells77", "sheshells78", "sheshells79", "sheshells80", "sheshells81", "sheshells82", "sheshells83", "sheshells84")
foreach ($page in $pages) {
    Write-Host "Adding page '$page'..."
    & $exePath add $page
    Write-Host "Page '$page' added."
}

# Compile the project
Write-Host "Compiling the project..."
& $exePath compile
Write-Host "Project compiled successfully."