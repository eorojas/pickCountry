# Gemini Project Context: pickCountry

## Project Overview
### Pick a country
From a list of countries create a web page that efficiently picks a country.

#### Components
- **Country Data Base:**
  A JSON list of countries with alternate spellings.
- **Input Matcher (Frontend):**
  - **1st stage:** Show matches as the user types name.
    - List shortens as typed.
    - If list > 20 items, it scrolls.
  - **Interaction:**
    - User types letters.
    - Up/Down arrows to select letter or country.
    - Right arrow/Enter to make selection.
    - Left arrow goes back one level.
  - **Receiver:**
    - Current list sent to a web page supporting list scrolling.
  - **Future stages:** Phonic matching, Spoken input.
- **Output Method:**
  Display the country name on the web page. Enter resets the process.

#### Configuration (Data Sources)
- **`alpha_countries.json`**: Countries indexed by first letter (A-Z).
- **`country_codes.json`**: Countries indexed by 2-letter ISO code.
- **`countries.json`**: Flat list of countries.

## Current Status
**Date:** February 1, 2026

### Data Validation & Preparation
- **Encoding**: Files (`alpha_countries.json`, `country_codes.json`, `countries.json`) were detected as ISO-8859-1 and successfully converted to UTF-8.
- **Validation**: All JSON files are now valid.
    - `alpha_countries.json`: 25 groups (Keys A-Z).
    - `country_codes.json`: 249 entries.
    - `countries.json`: 252 entries.

## Goals
1. **Primary:** Implement a Go program to read JSON data and serve it to a web page (default localhost:8081).
2. **Receiver:** An HTTPS receiver for the frontend input.

## Technology Stack
- **Language:** Go (Latest stable)
- **Frontend:** HTML/JS (No specific framework mandated, keeping it lightweight).

## Coding Preferences
- **Style:** Idiomatic Go (Effective Go).
- **Comments:** "Why" over "What".
- **Error Handling:** Explicit checks.

## Roadmap
- [x] **Data Prep**: Validate and fix JSON encoding/schema.
- [x] **Setup**: Initialize Go module (`go mod init`).
- [x] **Core Logic**: Implement data structures and map building (NameEntryMap, CountryCodeMap).
- [x] **Backend**: Create basic Go server to serve data.
- [x] **Frontend**: Create the web interface for the Input Matcher.
- [ ] **Testing**: Add unit tests for App Logic and Server.
- [ ] **Refinement**: Refactor and polish.

## Useful Commands
- `go run main.go` ( Anticipated )
- `python3 scripts/validate_json.py` ( If we save the script )

## Exploration & Research
- Need to determine best way to handle the "push to web page" requirement. Websockets or simple HTTP API?