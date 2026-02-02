package app

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/eorojas/pickCountry/internal/data"
)

// Config holds configuration for the application logic.
type Config struct {
	MaxListSize int // Default window size
}

// State represents the current UI state to be sent to the frontend.
type State struct {
	List         []string `json:"list"`
	Selection    int      `json:"selection"` // Index relative to the window
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
	
	// displayItems holds the full, pre-formatted list of "Name (Code)"
	displayItems []DisplayItem 
	
	// filteredIndices holds the indices of displayItems that match the current filter
	filteredIndices []int

	filter       string
	selection    int // Absolute index into filteredIndices
	isFinal      bool
	selectedCode string
	windowSize   int
}

type DisplayItem struct {
	Text string
	Code string
	Name string // Original name key for map lookup
}

// NewManager creates a new state manager.
func NewManager(nameMap data.NameEntryMap, codeMap data.CountryCodeMap, cfg Config) *Manager {
	if cfg.MaxListSize <= 0 {
		cfg.MaxListSize = 20
	}

	// flatten and sort
	// We want to list every valid name alias.
	// Format: "Name (Code)"
	var items []DisplayItem
	for name, code := range nameMap {
		items = append(items, DisplayItem{
			Text: fmt.Sprintf("%s (%s)", name, code),
			Code: string(code),
			Name: name,
		})
	}
	
	// Sort by display text
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Text) < strings.ToLower(items[j].Text)
	})

	// Initial filtered list is everything
	indices := make([]int, len(items))
	for i := range items {
		indices[i] = i
	}

	return &Manager{
		config:          cfg,
		nameMap:         nameMap,
		codeMap:         codeMap,
		displayItems:    items,
		filteredIndices: indices,
		filter:          "",
		selection:       0,
		windowSize:      cfg.MaxListSize,
	}
}

// SetWindowSize allows dynamic updating of the view window size
func (m *Manager) SetWindowSize(size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if size > 0 {
		m.windowSize = size
	}
}

// GetState returns the current state snapshot.
func (m *Manager) GetState() State {
	m.mu.Lock()
	defer m.mu.Unlock()

	totalMatches := len(m.filteredIndices)
	if totalMatches == 0 {
		return State{
			List:      []string{},
			Selection: 0,
			Filter:    m.filter,
			IsFinal:   m.isFinal,
		}
	}

	// Clamp selection
	if m.selection >= totalMatches {
		m.selection = totalMatches - 1
	}
	if m.selection < 0 {
		m.selection = 0
	}

	// Calculate Window
	// We want the selection to be roughly in the middle, or just ensure it's visible.
	// Simple strategy: Try to center selection.
	halfWindow := m.windowSize / 2
	start := m.selection - halfWindow
	if start < 0 {
		start = 0
	}
	end := start + m.windowSize
	if end > totalMatches {
		end = totalMatches
		// Adjust start if we hit the end
		start = end - m.windowSize
		if start < 0 {
			start = 0
		}
	}

	// Build windowed list
	windowList := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		originalIndex := m.filteredIndices[i]
		windowList = append(windowList, m.displayItems[originalIndex].Text)
	}

	// Selection index relative to the window
	relativeSelection := m.selection - start

	return State{
		List:         windowList,
		Selection:    relativeSelection,
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
		if key == "Enter" {
			m.reset()
			return
		}
	}

	totalMatches := len(m.filteredIndices)

	switch key {
	case "ArrowDown":
		if totalMatches > 0 && m.selection < totalMatches-1 {
			m.selection++
		}
	case "ArrowUp":
		if totalMatches > 0 && m.selection > 0 {
			m.selection--
		}
	case "PageDown":
		if totalMatches > 0 {
			m.selection += m.windowSize
			if m.selection >= totalMatches {
				m.selection = totalMatches - 1
			}
		}
	case "PageUp":
		if totalMatches > 0 {
			m.selection -= m.windowSize
			if m.selection < 0 {
				m.selection = 0
			}
		}
	case "ArrowLeft":
		// Backspace behavior for filter? Or strict navigation?
		// Requirement: "Left arrow goes back one level."
		// In a flat list model, this usually deletes the last char of filter.
		if len(m.filter) > 0 {
			m.updateFilter(m.filter[:len(m.filter)-1])
		}
	case "ArrowRight":
		// Select current
		if totalMatches > 0 {
			m.finalizeSelection()
		}

	case "Enter":
		if totalMatches > 0 {
			m.finalizeSelection()
		}
	case "Backspace":
		if len(m.filter) > 0 {
			m.updateFilter(m.filter[:len(m.filter)-1])
		}
	default:
		if len(key) == 1 {
			// Printable character
			m.updateFilter(m.filter + key)
		}
	}
}

func (m *Manager) updateFilter(newFilter string) {
	m.filter = newFilter
	prefix := strings.ToLower(m.filter)

	// Re-calculate filteredIndices
	// Optimization: If appending, we could filter the existing subset.
	// But sticking to full scan is safer and fast enough for <1000 items.
	
	var newIndices []int
	for i, item := range m.displayItems {
		// Contains or Prefix? "Input Matcher... show matches as the user types name"
		// Usually prefix for country pickers.
		if strings.HasPrefix(strings.ToLower(item.Text), prefix) {
			newIndices = append(newIndices, i)
		}
	}
	m.filteredIndices = newIndices
	m.selection = 0
}

func (m *Manager) finalizeSelection() {
	if m.selection >= 0 && m.selection < len(m.filteredIndices) {
		idx := m.filteredIndices[m.selection]
		item := m.displayItems[idx]
		
		m.filter = item.Name // Set filter to actual name? Or Display Text?
		// User requirement: "Output Method: Display the country name on web page."
		// Let's use the display text (Name + Code) or just Name?
		// Previously we used Name.
		// "When a country name is shown the country code shoud be shown in parenthesis."
		// This likely refers to the list view.
		// For final selection, let's keep it consistent.
		m.filter = item.Text 
		m.selectedCode = item.Code
		m.isFinal = true
	}
}

func (m *Manager) reset() {
	m.updateFilter("")
	m.selection = 0
	m.isFinal = false
	m.selectedCode = ""
}