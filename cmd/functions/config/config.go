package config

import (
	"github.com/rudyrdx/Go-Static/functions/helpers"
	"encoding/json"
	"fmt"
	"os"
)

type StaticJson struct {
	Pages       []string `json:"pages"`
	ProjectName string   `json:"projectName"`
	Tided       bool     `json:"tided"`
}

// updateJson updates the static.json file with the provided changes.
func UpdateJson(updates StaticJson) error {
	filePath := "static.json"

	var data StaticJson

	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// If the file doesn't exist, create it with the updates as the initial content
		data = updates
	} else {
		// If the file exists, read and update it
		file, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading static.json: %v\n", err)
			return err
		}

		// Unmarshal the existing content into the data struct
		if err := json.Unmarshal(file, &data); err != nil {
			fmt.Printf("Error decoding static.json: %v\n", err)
			return err
		}

		// Apply the updates
		if len(updates.Pages) > 0 {
			data.Pages = append(data.Pages, updates.Pages...)
			data.Pages = helpers.RemoveDuplicates(data.Pages)
		}
		if updates.ProjectName != "" {
			data.ProjectName = updates.ProjectName
		}
		if updates.Tided {
			data.Tided = updates.Tided
		}
	}

	// Marshal the updated struct back to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error creating static.json file: %v\n", err)
		return err
	}

	// Write the JSON data back to the file
	err = helpers.WriteFile(filePath, string(jsonData))
	if err != nil {
		fmt.Printf("Error writing to static.json: %v\n", err)
		return err
	}

	fmt.Printf("static.json updated successfully.\n")
	return nil
}

func ReadJson(filePath string) (*StaticJson, error) {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", filePath)
	}

	// Read the file content
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// Unmarshal the JSON content into a StaticJson object
	var data StaticJson
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, fmt.Errorf("error decoding JSON from file %s: %v", filePath, err)
	}

	return &data, nil
}
