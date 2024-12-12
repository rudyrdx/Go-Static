package main

import (
	"1/1/functions/config"
	"1/1/functions/setup"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: static [setup|add <page name>]")
		return
	}

	var staticjson config.StaticJson

	command := os.Args[1]
	switch command {
	case "setup":
		if len(os.Args) < 3 {
			fmt.Println("Please provide the project name: static setup <project name>")
			return
		}
		projectName := os.Args[2]
		setup.SetupProject(projectName, staticjson)
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a page name: static add <page name>")
			return
		}
		pageName := os.Args[2]
		setup.AddPage(pageName)
	case "compile":
		setup.CompileProject()
	default:
		fmt.Println("Unknown command. Use 'setup' or 'add <page name>'.")
	}
}

func watchProject() {
	//the concept of this function is to constantly watch the public and views directory for changes
	//we will call recompileProject whenever a change is detected
	//we know when a change is detected comparing the hashes of folders
	// every 1 second, we compare the current hash to the previous hash
	// if they are different, we call recompileProject
}
