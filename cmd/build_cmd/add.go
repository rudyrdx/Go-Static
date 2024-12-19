package buildcmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
)

// HandleAddCommand adds a new page to the project.
func HandleAddCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Please provide a page name: static add <page name>")
		return
	}
	pageName := os.Args[2]

	// Validate page name
	if strings.ContainsAny(pageName, " -./\\<>:\"|?*") || strings.HasSuffix(pageName, ".exe") {
		fmt.Println(Red + "Error: Invalid page name. Use only alphanumeric characters and avoid special symbols or '.exe' extensions." + Reset)
		return
	}

	fmt.Printf(Yellow+"You are about to add a page named '%s'. Confirm (y/n): "+Reset, pageName)
	reader := bufio.NewReader(os.Stdin)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(strings.ToLower(confirmation))

	// Validate confirmation input
	if confirmation != "y" && confirmation != "n" {
		fmt.Println(Red + "Invalid input. Please type 'y' for yes or 'n' for no." + Reset)
		return
	}

	if confirmation == "n" {
		fmt.Println(Yellow + "Page addition canceled." + Reset)
		return
	}

	// Add the page
	err := os.WriteFile(pageName+".html", []byte("<html><body><h1>"+pageName+"</h1></body></html>"), 0644)
	if err != nil {
		fmt.Println(Red+"Error creating page file:"+Reset, err)
		return
	}
	fmt.Println(Green + "Page created successfully. Running compile command..." + Reset)

	// Automatically run the compile command
	compileCmd := exec.Command(".\\static.exe", "compile")
	compileCmd.Stdout = os.Stdout
	compileCmd.Stderr = os.Stderr

	err = compileCmd.Run()
	if err != nil {
		fmt.Println(Red+"Error running compile command:"+Reset, err)
	} else {
		fmt.Println(Green + "Compile command executed successfully!" + Reset)
	}
}
