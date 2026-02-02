package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/eorojas/pickCountry/internal/data"
)

func main() {
	fmt.Println("Starting pickCountry...")

	alphaData, err := loadJSON[data.AlphaCountries]("alpha_countries.json")
	if err != nil {
		log.Fatalf("Error loading alpha_countries.json: %v", err)
	}

	codeData, err := loadJSON[data.CountryCodes]("country_codes.json")
	if err != nil {
		log.Fatalf("Error loading country_codes.json: %v", err)
	}

	nameMap, codeMap, err := data.BuildMapsFromJSON(alphaData, codeData)
	if err != nil {
		log.Fatalf("Data integrity error: %v", err)
	}

	fmt.Printf("Successfully loaded %d countries and %d name entries.\n", len(codeMap), len(nameMap))
	
	// Example lookup
	if code, ok := nameMap["Taiwan, Province of China"]; ok {
		fmt.Printf("Lookup 'Taiwan, Province of China' -> %s\n", code)
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