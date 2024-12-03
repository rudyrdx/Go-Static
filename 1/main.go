package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type StaticJson struct {
	Pages []string `json:"pages"`
}

// Templates for layout and page files
const mainTemplate = `package main

import (
	"context"
	"log"
	"os"

	"github.com/a-h/templ"
)

func createFile(name string) (*os.File, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func renderTemplateToFile(component templ.Component, ctx context.Context, f *os.File) error {
	defer f.Close()
	return component.Render(ctx, f)
}

func main() {
	// in the output folder
	templates := map[string]templ.Component{
		"index.html":    Home(),
		"about/index.html":   About(),
		"/contacts/index.html": Contacts(),
	}

	for filename, component := range templates {
		file, err := createFile(filename)
		if err != nil {
			log.Fatal(err)
		}

		if err := renderTemplateToFile(component, context.Background(), file); err != nil {
			log.Fatal(err)
		}
	}
}`

const layoutTemplate = `package layout
templ Layout(title string) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
			<link rel="stylesheet" href="styles.css">
			<title>{ title }</title>
		</head>
		<body>
			<header></header>
			{ children... }
			<footer></footer>
		</body>
	</html>
}`

const homeTemplate = `package home
templ Home() {
	@layout.Layout("Home") {
		<!-- Add content here -->
	}
}`

const genericPageTemplate = `package {{ .PageName }}
templ {{ .PageName }}() {
	@layout.Layout("{{ .PageName }}") {
		<!-- Add content here -->
	}
}`

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: static [setup|add <page name>]")
		return
	}

	command := os.Args[1]
	switch command {
	case "setup":
		if len(os.Args) < 3 {
			fmt.Println("Please provide the project name: static setup <project name>")
			return
		}
		projectName := os.Args[2]
		setupProject(projectName)
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a page name: static add <page name>")
			return
		}
		pageName := os.Args[2]
		addPage(pageName)
	case "compile":

	default:
		fmt.Println("Unknown command. Use 'setup' or 'add <page name>'.")
	}
}

// setupProject initializes the project structure
func setupProject(projName string) {
	directories := []string{
		"output",
		"views/layout",
		"public/style",
	}

	for _, dir := range directories {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	//setup main.go
	cmd := exec.Command("go", "mod", "init", projName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error creating go.mod file: %v\n", err)
		return
	}

	// Create layout.templ
	writeFile("views/layout/layout.templ", layoutTemplate)

	// Create styles.css
	writeFile("public/style/styles.css", "/* Add your CSS here */")

	pagesData := StaticJson{
		Pages: []string{""},
	}

	// Create static.json
	jsonData, err := json.MarshalIndent(pagesData, "", "  ")
	if err != nil {
		fmt.Printf("Error creating static.json file: %v\n", err)
		return
	}

	writeFile("static.json", string(jsonData))

	fmt.Println("Project setup completed successfully. make sure to run 'go mod tidy' to update the go.mod file.")
}

// addPage creates a new page template
func addPage(pageName string) {
	var filePath string
	var content string

	if pageName == "home" {
		filePath = "views/home.templ"
		content = homeTemplate
	} else {
		dirPath := filepath.Join("views", pageName)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dirPath, err)
			return
		}
		filePath = filepath.Join(dirPath, pageName+".templ")

		tmpl, err := template.New("generic").Parse(genericPageTemplate)
		if err != nil {
			fmt.Printf("Template parsing error: %v\n", err)
			return
		}

		// Create the file with dynamic content
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filePath, err)
			return
		}
		defer file.Close()

		data := struct{ PageName string }{PageName: pageName}
		tmpl.Execute(file, data)
		fmt.Printf("Page %s created successfully.\n", pageName)
		addPageToStaticJSON(pageName)
		return
	}

	// Create the home template
	writeFile(filePath, content)
	fmt.Printf("Page %s created successfully.\n", pageName)
	addPageToStaticJSON(pageName)
}

// writeFile writes content to the specified file
func writeFile(filePath, content string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", filePath, err)
	}
}

func addPageToStaticJSON(pageName string) {
	// Open the static.json file
	file, err := os.Open("static.json")
	if err != nil {
		fmt.Printf("Error opening file static.json: %v\n", err)
		return
	}
	defer file.Close()

	var data StaticJson

	//read the file content and unmarshal it into the data struct
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Printf("Error decoding static.json: %v\n", err)
		return
	}

	// Add the new page to the data struct
	data.Pages = append(data.Pages, pageName)

	// Marshal the data struct back to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error creating static.json file: %v\n", err)
		return
	}

	// Write the JSON data back to the file
	err = os.WriteFile("static.json", jsonData, os.ModePerm)
	if err != nil {
		fmt.Printf("Error writing to static.json: %v\n", err)
	}
	fmt.Printf("Page %s added to static.json.\n", pageName)
}
