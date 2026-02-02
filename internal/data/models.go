package data

import (
	"fmt"
	"strings"
)

// CountryCode represents a two-letter ISO country code.
type CountryCode string

// NameEntry represents a mapping from a name (or alias) to a country code.
type NameEntry struct {
	Name string
	Code CountryCode
}

// Country represents the canonical country data.
type Country struct {
	Name string      `json:"name"`
	Code CountryCode `json:"code"`
}

// NameEntryMap maps a country name (or alias) to its CountryCode.
type NameEntryMap map[string]CountryCode

// CountryCodeMap maps a CountryCode to the canonical Country struct.
type CountryCodeMap map[CountryCode]*Country

// AlphaEntry is the structure used in alpha_countries.json
type AlphaEntry struct {
	Name string `json:"n"`
	Code string `json:"c"`
}

// AlphaCountries matches the structure of alpha_countries.json
type AlphaCountries map[string][]AlphaEntry

// CountryCodes matches the structure of country_codes.json
type CountryCodes map[string][]string

// BuildMapsFromJSON constructs the NameEntryMap and CountryCodeMap from the two JSON data structures.
func BuildMapsFromJSON(alpha AlphaCountries, codes CountryCodes) (NameEntryMap, CountryCodeMap, error) {
	nameMap := make(NameEntryMap)
	codeMap := make(CountryCodeMap)

	// 1. Process country_codes.json (Primary source for codes and aliases)
	for c, names := range codes {
		if len(names) == 0 {
			continue
		}
		code := CountryCode(strings.ToUpper(c))
		
		// First name in the list is treated as the canonical name
		codeMap[code] = &Country{
			Name: names[0],
			Code: code,
		}

		for _, name := range names {
			if err := addNameEntry(nameMap, name, code); err != nil {
				return nil, nil, err
			}
		}
	}

	// 2. Cross-reference with alpha_countries.json to ensure consistency and pick up any missing names
	for _, entries := range alpha {
		for _, entry := range entries {
			code := CountryCode(strings.ToUpper(entry.Code))
			
			// If we haven't seen this code in country_codes.json, add it
			if _, exists := codeMap[code]; !exists {
				codeMap[code] = &Country{
					Name: entry.Name,
					Code: code,
				}
			}

			// Add the name from alpha list if it's not already there
			if err := addNameEntry(nameMap, entry.Name, code); err != nil {
				// We don't return error here if it's the same code (already handled in addNameEntry)
				return nil, nil, err
			}
		}
	}

	return nameMap, codeMap, nil
}

func addNameEntry(m NameEntryMap, name string, code CountryCode) error {
	if existingCode, exists := m[name]; exists {
		if existingCode != code {
			return fmt.Errorf("name collision: '%s' is already mapped to %s, cannot map to %s", name, existingCode, code)
		}
		return nil 
	}
	m[name] = code
	return nil
}