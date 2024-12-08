package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type StaticJson struct {
	ProjectName string   `json:"projectName"`
	Pages       []string `json:"pages"`
}

// Templates for layout and page files
const mainTemplate = `package main

import (
	"context"
	"log"
	"os"
	"io"
	"path/filepath"
	{{ .ViewsImports }}
	"github.com/a-h/templ"
)

func createFile(name string) (*os.File, error) {
	dirPath := filepath.Dir(name)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}
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

func copyStaticFiles(src, dst string) error {
    if _, err := os.Stat(src); os.IsNotExist(err) {
        return err
    }

    if _, err := os.Stat(dst); os.IsNotExist(err) {
        err := os.MkdirAll(dst, os.ModePerm)
        if err != nil {
            return err
        }
    }

    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        relPath, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }

        destPath := filepath.Join(dst, relPath)

        if info.IsDir() {
            // Create the directory if it does not exist
            if _, err := os.Stat(destPath); os.IsNotExist(err) {
                err := os.MkdirAll(destPath, info.Mode())
                if err != nil {
                    return err
                }
            }
            return nil
        }

        srcFile, err := os.Open(path)
        if err != nil {
            return err
        }
        defer srcFile.Close()

        destFile, err := os.Create(destPath)
        if err != nil {
            return err
        }
        defer destFile.Close()

        _, err = io.Copy(destFile, srcFile)
        if err != nil {
            return err
        }

        return nil
    })
}

func main() {
	{{ .Templates }}

	outDir := "output"
	static := "public"

	for filename, component := range templates {

		var filePath string
		if filename == "home/index.html" {
			filePath = filepath.Join(outDir, "index.html")
		} else {
			filePath = filepath.Join(outDir, filename)
		}

		file, err := createFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		if err := renderTemplateToFile(component, context.Background(), file); err != nil {
			log.Fatal(err)
		}
	}

	if err := copyStaticFiles(static, outDir); err != nil {
		log.Fatal(err)
	}
}`

const layoutTemplate = `package layout
templ Layout(title string) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
			<link rel="stylesheet" href="/style/styles.css">
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
import "{{ .ProjectName }}/views/layout"

templ Home() {
	@layout.Layout("Home") {
		<h1>Welcome to the home page</h1>
	}
}`

const genericPageTemplate = `package {{ .PageName }}
import "{{ .ProjectName }}/views/layout"

templ {{ .CPageName }}() {
	@layout.Layout("{{ .PageName }}") {
		<h1> Welcome to the {{ .PageName }} page</h1>
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
		compileProject()
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

	//go get github.com/a-h/templ
	cmd2 := exec.Command("go", "get", "github.com/a-h/templ")
	err2 := cmd2.Run()
	if err2 != nil {
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

	addProjectNameToStaticJSON(projName)

	fmt.Println("Project setup completed successfully. make sure to run 'go mod tidy' to update the go.mod file.")
}

func contains(slice []string, item string) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

// addPage creates a new page template
func addPage(pageName string) {

	caser := cases.Title(language.English)
	CpageName := caser.String(pageName)
	fileContent, err := os.ReadFile("static.json")
	if err != nil {
		fmt.Printf("Error reading static.json: %v\n", err)
		return
	}

	var data StaticJson
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		fmt.Printf("Error decoding static.json: %v\n", err)
		return
	}

	//check if the page already exists then return
	if contains(data.Pages, pageName) {
		return
	}

	var filePath string
	var tmplContent string
	var tmplName string

	if pageName == "home" {
		filePath = "views/home.templ"
		tmplContent = homeTemplate
		tmplName = "home"
	} else {
		dirPath := filepath.Join("views", pageName)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dirPath, err)
			return
		}
		filePath = filepath.Join(dirPath, pageName+".templ")
		tmplContent = genericPageTemplate
		tmplName = "generic"
	}

	tmpl, err := template.New(tmplName).Parse(tmplContent)
	if err != nil {
		fmt.Printf("Template parsing error: %v\n", err)
		return
	}

	project := data.ProjectName
	pageData := struct {
		ProjectName string
		PageName    string
		CPageName   string
	}{
		ProjectName: project,
		PageName:    pageName,
		CPageName:   CpageName,
	}

	var b strings.Builder
	err = tmpl.Execute(&b, pageData)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	writeFile(filePath, b.String())

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
	file, err := os.ReadFile("static.json")
	if err != nil {
		fmt.Printf("Error reading static.json: %v\n", err)
		return
	}

	var data StaticJson

	//read the file content and unmarshal it into the data struct
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Printf("Error decoding static.json: %v", err)
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

func addProjectNameToStaticJSON(projectName string) {
	// Open the static.json file
	file, err := os.ReadFile("static.json")
	if err != nil {
		fmt.Printf("Error reading static.json: %v\n", err)
		return
	}

	var data StaticJson

	//read the file content and unmarshal it into the data struct
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Printf("Error decoding static.json: %v", err)
	}

	// Add the new page to the data struct
	data.ProjectName = projectName

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
	fmt.Printf("Project name %s added to static.json.\n", projectName)
}

func compileProject() {

	data, err := os.ReadFile("static.json")
	if err != nil {
		fmt.Printf("Error reading static.json: %v\n", err)
		return
	}

	var pages StaticJson
	err = json.Unmarshal(data, &pages)
	if err != nil {
		fmt.Printf("Error decoding static.json: %v\n", err)
		return
	}

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
		temp += fmt.Sprintf(`"%s/index.html": %s.%s(),`, page, page, cPage)
		if page != "home" {
			importTemp += fmt.Sprintf("\t\"%s/views/%s\"\n", pages.ProjectName, page)
		} else {
			importTemp += fmt.Sprintf("\t\"%s/views\"\n", pages.ProjectName)
		}
	}

	templates = strings.Replace(templates, "{{ .replace }}", temp, -1)

	tmpl, err := template.New("maintemplate").Parse(mainTemplate)
	if err != nil {
		fmt.Printf("Template parsing error: %v\n", err)
		return
	}

	mainGo := struct {
		ViewsImports string
		Templates    string
	}{
		ViewsImports: importTemp,
		Templates:    templates,
	}

	file, err := os.Create("main.go")
	if err != nil {
		fmt.Printf("Error creating file main.go: %v\n", err)
		return
	}
	defer file.Close()

	tmpl.Execute(file, mainGo)

	cmd := exec.Command("go", "mod", "tidy")
	err1 := cmd.Run()
	if err1 != nil {
		fmt.Printf("Error tidying file: %v\n", err1)
		return
	}

	cmd2 := exec.Command("templ", "generate")
	err2 := cmd2.Run()
	if err2 != nil {
		fmt.Printf("Error creating go.mod file: %v\n", err2)
		return
	}
	cmd3 := exec.Command("go", "run", ".")
	err3 := cmd3.Run()
	if err3 != nil {
		fmt.Printf("Error creating go.mod file: %v\n", err3)
		return
	}

}

func watchProject() {
	//the concept of this function is to constantly watch the public and views directory for changes
	//we will call recompileProject whenever a change is detected
	//we know when a change is detected comparing the hashes of folders
	// every 1 second, we compare the current hash to the previous hash
	// if they are different, we call recompileProject
}
