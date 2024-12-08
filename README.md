# Go-Static
## Overview

This project is a static site generator designed to streamline the development of static web pages using pre-defined templates [templ](https://github.com/a-h/templ) and a structured file organization. It simplifies the process of setting up a project, adding new pages, and compiling them into a static output that can be easily served by a web server.

## Demo
[![YouTube Video](https://img.youtube.com/vi/KvXslPeZiEA/0.jpg)](https://www.youtube.com/watch?v=KvXslPeZiEA)

## Features

- **Automated Project Setup**: Quickly create a structured project directory with necessary folders for views, layouts, and public assets.
- **Easy Page Addition**: Easily add new pages with pre-baked templates using the templ language to maintain consistency across the site.
- **Static Compilation**: Compile all templates into Go source files, handle asset integration, and generate static output files ready to be served.

## Installation

Ensure you have Go installed on your machine as it is required for compiling templates. Clone the project repository and navigate to the project directory.

```bash
git clone https://github.com/rudyrdx/Go-Static
cd cmd
make build
```

## Usage

### 1. Set Up the Project Directory

To set up the initial project structure, run:

```bash
static setup
```

This command will create the following directory structure:

```
root/
    output/
    views/
        layout/
            layout.templ
    public/
        style/
            styles.css
```

### 2. Add a New Page

To add a new page, use the command:

```bash
static add <page-name>
```

- If `<page-name>` is `home`, a file named `home.templ` will be created directly in the `views` folder.
- For other `<page-name>`, a new directory with the page's name will be created inside `views`, containing a template file in the format of `Home.templ`.

### Example

For a `home` page, the structure would be:

```
views/
    home.templ
```

Content of `home.templ`:

```templ
package home

templ Home() {
    @layout.Layout("Home") {
        // Insert page content here.
    }
}
```

### 3. Compile the Templates

To compile the templates and generate output files, run:

```bash
static compile
```

This will perform the following actions:

- Execute the `templ` generate command.
- Generate Go source files for all directories inside `views`.
- Reference these compiled templates in a `main.go` file.
- Output static files and assets into the `output` folder where a server will be set up for preview.

## Viewing the Site

After compiling, the output folder will consist of all the output html files and the styles or scripts mentioned in the public dir. for now you can use any fileserver like vs-code go live extension or any http server

## Contributing

Contributions to improve the project are welcome. Please create a fork of the repository and submit a pull request for any changes you wish to make.

## License

This project is licensed under the MIT License. See the `LICENSE` file for more information. 

## Contact

For any questions, feel free to open an issue in the repository or contact us via [email](mailto:rudyrdx21@gmail.com).
