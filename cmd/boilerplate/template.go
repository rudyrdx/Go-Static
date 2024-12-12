package boilerplate

const MainTemplate = `package main

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

// purpose of the template is to create the layout and home page templates
const LayoutTemplate = `package layout
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

// homeTemplate is the template for the home page
const HomeTemplate = `package home
import "{{ .ProjectName }}/views/layout"

templ Home() {
	@layout.Layout("Home") {
		<h1>Welcome to the home page</h1>
	}
}`

// genericPageTemplate is the template for the generic page
const GenericPageTemplate = `package {{ .PageName }}
import "{{ .ProjectName }}/views/layout"

templ {{ .CPageName }}() {
	@layout.Layout("{{ .PageName }}") {
		<h1> Welcome to the {{ .PageName }} page</h1>
	}
}`
