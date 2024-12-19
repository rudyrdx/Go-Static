package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rudyrdx/Go-Static/boilerplate"
	"github.com/rudyrdx/Go-Static/functions/config"
	"github.com/rudyrdx/Go-Static/functions/helpers"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func SetupProject(projName string, pagesData config.StaticJson) error {
	directories := []string{
		"output",
		"views/layout",
		"public/style",
	}

	// Create required directories
	if err := createDirectories(directories); err != nil {
		return err
	}

	// Initialize Go project and fetch dependencies
	if err := initializeGoProject(projName); err != nil {
		return err
	}

	// Create necessary files
	if err := createFiles(); err != nil {
		return err
	}

	// Update the static.json file
	pagesData.Pages = []string{""}
	pagesData.ProjectName = projName
	pagesData.Tided = false
	if err := config.UpdateJson(pagesData); err != nil {
		fmt.Printf("Error updating static.json: %v\n", err)
		return err
	}

	fmt.Println("Project setup completed successfully. Run 'go mod tidy' to update the go.mod file.")
	return nil
}

// Helper function to create directories
func createDirectories(directories []string) error {
	for _, dir := range directories {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return err
		}
	}
	return nil
}

// Helper function to initialize a Go project
func initializeGoProject(projName string) error {
	commands := []struct {
		name string
		args []string
	}{
		{"go", []string{"mod", "init", projName}},
		{"go", []string{"get", "github.com/a-h/templ"}},
	}

	for _, cmd := range commands {
		if err := exec.Command(cmd.name, cmd.args...).Run(); err != nil {
			fmt.Printf("Error running '%s %v': %v\n", cmd.name, cmd.args, err)
			return err
		}
	}
	return nil
}

// Helper function to create default files
func createFiles() error {
	files := map[string]string{
		"views/layout/layout.templ": boilerplate.LayoutTemplate,
		"public/style/styles.css":   "/* Add your CSS here */",
	}

	for path, content := range files {
		if err := helpers.WriteFile(path, content); err != nil {
			fmt.Printf("Error creating file %s: %v\n", path, err)
			return err
		}
	}
	return nil
}

func AddPage(pageName string) {
	caser := cases.Title(language.English)
	CpageName := caser.String(pageName)

	// Read and unmarshal the static.json file
	data, err := config.ReadJson("static.json")
	if err != nil {
		fmt.Printf("Error reading static.json: %v\n", err)
		return
	}

	// Check if the page already exists in the JSON
	if helpers.Contains(data.Pages, pageName) {
		fmt.Printf("Page %s already exists.\n", pageName)
		return
	}

	// Determine file paths and template content based on page type
	var filePath string
	var tmplContent string
	var tmplName string

	if pageName == "home" {
		filePath = "views/home.templ"
		tmplContent = boilerplate.HomeTemplate
		tmplName = "home"
	} else {
		// Create a new directory for non-home pages
		dirPath := filepath.Join("views", pageName)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dirPath, err)
			return
		}
		filePath = filepath.Join(dirPath, pageName+".templ")
		tmplContent = boilerplate.GenericPageTemplate
		tmplName = "generic"
	}

	// Parse the template
	tmpl, err := template.New(tmplName).Parse(tmplContent)
	if err != nil {
		fmt.Printf("Template parsing error: %v\n", err)
		return
	}

	// Prepare data for the template
	pageData := struct {
		ProjectName string
		PageName    string
		CPageName   string
	}{
		ProjectName: data.ProjectName,
		PageName:    pageName,
		CPageName:   CpageName,
	}

	// Execute the template and write the result to the file
	var b strings.Builder
	if err := tmpl.Execute(&b, pageData); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	// Write the generated content to the file
	helpers.WriteFile(filePath, b.String())
	fmt.Printf("Page %s created successfully.\n", pageName)

	// Update the static.json with the new page
	data.Pages = append(data.Pages, pageName)
	if err := config.UpdateJson(*data); err != nil {
		fmt.Printf("Error updating static.json: %v\n", err)
		return
	}
}

func CompileProject() {
	pages, err := config.ReadJson("static.json")
	if err != nil {
		fmt.Printf("Error reading static.json: %v\n", err)
		return
	}

	// Template for mapping templates
	var templates = `templates := map[string]templ.Component{
		{{ .replace }}
	}
	`

	var temp string
	var importTemp string
	caser := cases.Title(language.English)

	for _, page := range pages.Pages {
		if page == "" {
			continue
		}
		cPage := caser.String(page)

		// Add template mapping for the page
		temp += fmt.Sprintf(`"%s/index.html": %s.%s(),`, page, page, cPage)

		// Handle imports properly
		if page != "home" {
			importTemp += fmt.Sprintf("\t\"%s/views/%s\"\n", pages.ProjectName, filepath.ToSlash(page))
		} else {
			importTemp += fmt.Sprintf("\t\"%s/views\"\n", pages.ProjectName)
		}
	}

	// Replace the placeholder in the templates
	templates = strings.Replace(templates, "{{ .replace }}", temp, -1)

	// Parse the main template
	tmpl, err := template.New("maintemplate").Parse(boilerplate.MainTemplate)
	if err != nil {
		fmt.Printf("Template parsing error: %v\n", err)
		return
	}

	// Populate data for the main.go template
	mainGo := struct {
		ViewsImports string
		Templates    string
	}{
		ViewsImports: importTemp,
		Templates:    templates,
	}

	// Create or overwrite the main.go file
	file, err := os.Create("main.go")
	if err != nil {
		fmt.Printf("Error creating file main.go: %v\n", err)
		return
	}
	defer file.Close()

	// Execute the template and write to file
	if err := tmpl.Execute(file, mainGo); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	// Run Go commands
	commands := []struct {
		Cmd  string
		Args []string
	}{
		{"go", []string{"mod", "tidy"}},
		{"templ", []string{"generate"}},
		{"go", []string{"run", "."}},
	}

	for _, cmd := range commands {

		if cmd.Args[0] == "mod" {
			if pages.Tided {
				continue
			} else {
				pages.Tided = true
				if err := config.UpdateJson(*pages); err != nil {
					fmt.Printf("Error updating static.json: %v\n", err)
					return
				}
			}
		}
		if err := helpers.RunCommand(cmd.Cmd, cmd.Args...); err != nil {
			fmt.Printf("Error running '%s %v': %v\n", cmd.Cmd, cmd.Args, err)
			return
		}
	}

	fmt.Println("Project compiled successfully.")
}
