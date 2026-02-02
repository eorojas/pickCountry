package app

import (
	"sort"
	"strings"
	"sync"

	"github.com/eorojas/pickCountry/internal/data"
)

// Config holds configuration for the application logic.
type Config struct {
	MaxListSize int // N, default 20
}

// State represents the current UI state to be sent to the frontend.
type State struct {
	List         []string `json:"list"`
	Selection    int      `json:"selection"`
	Filter       string   `json:"filter"`
	IsFinal      bool     `json:"is_final"`
	SelectedCode string   `json:"selected_code,omitempty"`
}

// Manager handles the application state and logic.
type Manager struct {
	mu           sync.Mutex
	config       Config
	nameMap      data.NameEntryMap
	codeMap      data.CountryCodeMap
	sortedNames  []string // Pre-sorted list of all valid names/aliases
	
	filter       string
	selection    int
	isFinal      bool
	selectedCode string
}

// NewManager creates a new state manager.
func NewManager(nameMap data.NameEntryMap, codeMap data.CountryCodeMap, cfg Config) *Manager {
	if cfg.MaxListSize <= 0 {
		cfg.MaxListSize = 20
	}

	// Create a sorted list of all keys for easier filtering
	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}
	sort.Strings(names)

	return &Manager{
		config:      cfg,
		nameMap:     nameMap,
		codeMap:     codeMap,
		sortedNames: names,
		filter:      "",
		selection:   0,
	}
}

// GetState returns the current state snapshot.
func (m *Manager) GetState() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	list := m.getCurrentList()
	
	// Ensure selection is within bounds
	if m.selection >= len(list) {
		m.selection = len(list) - 1
	}
	if m.selection < 0 && len(list) > 0 {
		m.selection = 0
	}

	return State{
		List:         list,
		Selection:    m.selection,
		Filter:       m.filter,
		IsFinal:      m.isFinal,
		SelectedCode: m.selectedCode,
	}
}

// ProcessInput handles user input (char or navigation).
func (m *Manager) ProcessInput(key string, code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isFinal {
		// Reset on any input if currently in final state, or specifically Enter?
		// "Enter should reset the selection process"
		if key == "Enter" {
			m.reset()
			return
		}
		// Optional: Start over if typing? For now let's just stick to Enter resets.
	}

	list := m.getCurrentList()

	switch key {
	case "ArrowDown":
		if m.selection < len(list)-1 {
			m.selection++
		}
	case "ArrowUp":
		if m.selection > 0 {
			m.selection--
		}
	case "ArrowLeft":
		// "Left arrow goes back one level."
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.selection = 0
		}
	case "ArrowRight":
		// "Next selection is make by right arrow or letter, or <enter>."
		// Logic: If list item is a single letter (next char), append it.
		// If it's a full name, maybe select it?
		// For now, let's treat Right Arrow as "Auto-complete current selection" if it makes sense,
		// or just strictly follow "next selection".
		// If we are selecting a letter from the alphabet list, Right Arrow should 'type' that letter.
		if m.selection >= 0 && m.selection < len(list) {
			item := list[m.selection]
			// If item is a single letter, append it
			if len(item) == 1 {
				m.filter += item
				m.selection = 0
			} else {
				m.filter = item
				m.selectedCode = string(m.nameMap[item])
				m.isFinal = true
			}
		}

	case "Enter":
		if m.selection >= 0 && m.selection < len(list) {
			item := list[m.selection]
			if len(item) == 1 {
				m.filter += item
				m.selection = 0
			} else {
				// Finalize
				m.filter = item
				m.selectedCode = string(m.nameMap[item])
				m.isFinal = true
			}
		}
	case "Backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.selection = 0
		}
	default:
		if len(key) == 1 {
			// Printable character
			m.filter += key
			m.selection = 0 // Reset selection when filtering changes
		}
	}
}

func (m *Manager) reset() {
	m.filter = ""
	m.selection = 0
	m.isFinal = false
	m.selectedCode = ""
}

// getCurrentList calculates the list to display based on the current filter.
// Logic:
// 1. Filter sortedNames by prefix `m.filter`.
// 2. If count <= N, return the names.
// 3. If count > N, return the list of "next letters".
//    - "next letter" means the character at index `len(m.filter)` of the matching names.
func (m *Manager) getCurrentList() []string {
	// If filter is empty, return A-Z
	if m.filter == "" {
		return generateAlphabet()
	}

	// 1. Find all matches
	matches := []string{}
	prefix := strings.ToLower(m.filter)
	
	// Optimization: could use binary search since sortedNames is sorted
	for _, name := range m.sortedNames {
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			matches = append(matches, name)
		}
	}

	// If exact match found (and it's the only one, or we want to show it specifically?)
	// If matches is empty, we might want to handle that (no results).
	if len(matches) == 0 {
		return []string{}
	}

	// 2. Check size
	if len(matches) <= m.config.MaxListSize {
		// Case: List short enough to display fully
		// Note: We might want to preserve the original casing from sortedNames
		return matches
	}

	// 3. List too long, show next letters
	// "list of next letters appears"
	// We need unique next letters
	nextChars := make(map[string]struct{})
	filterLen := len(m.filter)
	
	for _, name := range matches {
		if len(name) > filterLen {
			// Get the next character (runes to be safe with utf8)
			runes := []rune(name)
			if len(runes) > filterLen {
				// We want the character relative to the filter length
				// But wait, the filter might ignore case, but the output should probably correspond
				// to the actual names. 
				// "E.g., if 'a' is typed the list of next letters appears"
				// If I typed "a", matches are "Afghanistan", "Albania".
				// Next letters are 'f', 'l'.
				// However, we need to match case-insensitively for the *filter*,
				// but the "next letter" display usually implies valid next inputs.
				// Let's normalize next letters to Upper Case for display? Or keep them as is?
				// Usually "next letter" menus use Upper Case or lowercase consistently.
				// Let's use the actual char from the name for now, but deduplicate case-insensitively?
				// "A" -> "f" (Afghanistan), "l" (Albania).
				
				// Using case-insensitive deduplication for the menu usually looks cleaner (A-Z list)
				// but names might have special chars.
				// Let's stick to the character from the name.
				
				// Actually, if I type 'a', and I see 'f', I expect to type 'f' next.
				// So if I type 'f', filter becomes "af".
				
				char := string(runes[filterLen])
				nextChars[strings.ToUpper(char)] = struct{}{}
			}
		}
	}
	
	result := make([]string, 0, len(nextChars))
	for c := range nextChars {
		result = append(result, c)
	}
	sort.Strings(result)
	return result
}

func generateAlphabet() []string {
	alpha := make([]string, 26)
	for i := 0; i < 26; i++ {
		alpha[i] = string(rune('A' + i))
	}
	return alpha
}
