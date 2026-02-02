package data

import (
	"testing"
)

func TestBuildMapsFromJSON_Success(t *testing.T) {
	alpha := AlphaCountries{
		"U": []AlphaEntry{{Name: "United States", Code: "US"}},
		"C": []AlphaEntry{{Name: "Canada", Code: "CA"}},
	}
	codes := CountryCodes{
		"US": {"United States", "USA", "America"},
		"CA": {"Canada"},
	}

	nameMap, codeMap, err := BuildMapsFromJSON(alpha, codes)
	if err != nil {
		t.Fatalf("BuildMapsFromJSON failed: %v", err)
	}

	if len(codeMap) != 2 {
		t.Errorf("Expected 2 codes, got %d", len(codeMap))
	}

	if nameMap["USA"] != "US" {
		t.Errorf("Expected USA -> US, got %s", nameMap["USA"])
	}
}

func TestBuildMapsFromJSON_Collision(t *testing.T) {
	alpha := AlphaCountries{
		"X": []AlphaEntry{{Name: "Collision", Code: "X1"}},
	}
	codes := CountryCodes{
		"X2": {"Collision"},
	}

	_, _, err := BuildMapsFromJSON(alpha, codes)
	if err == nil {
		t.Fatal("Expected error for name collision across files, got nil")
	}
}