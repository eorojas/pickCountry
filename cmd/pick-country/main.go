package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/eorojas/pickCountry/internal/app"
	"github.com/eorojas/pickCountry/internal/data"
	"github.com/eorojas/pickCountry/internal/server"
)

func main() {
	port := flag.Int("port", 8081, "Port to run the server on")
	flag.Parse()

	fmt.Println("Starting pickCountry...")

	// 1. Load Data
	// Locate data files (simple heuristic for dev vs prod/test)
	alphaPath := "alpha_countries.json"

codesPath := "country_codes.json"

	if _, err := os.Stat(alphaPath); os.IsNotExist(err) {
		// Try stepping up one dir if running from cmd subfolder without correct cwd
		if _, err := os.Stat("../../" + alphaPath); err == nil {
			alphaPath = "../../" + alphaPath
		
codesPath = "../../" + codesPath
		}
	}

	alphaData, err := loadJSON[data.AlphaCountries](alphaPath)
	if err != nil {
		log.Fatalf("Error loading %s: %v", alphaPath, err)
	}

	codeData, err := loadJSON[data.CountryCodes](codesPath)
	if err != nil {
		log.Fatalf("Error loading %s: %v", codesPath, err)
	}

	nameMap, codeMap, err := data.BuildMapsFromJSON(alphaData, codeData)
	if err != nil {
		log.Fatalf("Data integrity error: %v", err)
	}

	fmt.Printf("Loaded %d countries.\n", len(codeMap))

	// 2. Initialize App State Manager
	cfg := app.Config{
		MaxListSize: 20,
	}
	manager := app.NewManager(nameMap, codeMap, cfg)

	// 3. Start Server
	srv := &server.Server{
		Manager: manager,
		Port:    *port,
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func loadJSON[T any](filename string) (T, error) {
	var target T
	content, err := os.ReadFile(filename)
	if err != nil {
		return target, err
	}
	err = json.Unmarshal(content, &target)
	return target, err
}
