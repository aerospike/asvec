package interactive

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aerospike/avs-client-go"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#7D56F4")).Bold(true).
				SetString("▶ ")

	highlightedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render
)

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
	Query  key.Binding
	Quit   key.Binding
}

// DefaultKeyMap returns a set of default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Query: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "query"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

// QueryModel represents the state of the query application
type QueryModel struct {
	neighbors       []*avs.Neighbor
	cursor          int
	selected        int // -1 means no selection
	queryRequested  bool
	detailsViewport viewport.Model
	keyMap          KeyMap
	width           int
	height          int
	maxDataKeys     int
	maxDataColWidth int
	includeFields   []string
}

// NewQueryModel creates a new model with the given neighbors
func NewQueryModel(
	neighbors []*avs.Neighbor,
	maxDataKeys int,
	maxDataColWidth int,
	includeFields []string,
) *QueryModel {
	// Sort neighbors by distance for consistent display
	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].Distance < neighbors[j].Distance
	})

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		PaddingRight(2)

	return &QueryModel{
		neighbors:       neighbors,
		cursor:          0,
		selected:        -1,
		detailsViewport: vp,
		keyMap:          DefaultKeyMap(),
		maxDataKeys:     maxDataKeys,
		maxDataColWidth: maxDataColWidth,
		includeFields:   includeFields,
	}
}

// Init initializes the model
func (m *QueryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and user input
func (m *QueryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Query):
			if m.selected != -1 {
				// Set flag to indicate we want to query with this record
				m.queryRequested = true
				return m, tea.Quit
			}

		case key.Matches(msg, m.keyMap.Back):
			if m.selected != -1 {
				// Go back to list view
				m.selected = -1
				return m, nil
			}
			// Otherwise quit
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Up):
			if m.selected == -1 && m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keyMap.Down):
			if m.selected == -1 && m.cursor < len(m.neighbors)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keyMap.Select):
			if m.selected == -1 {
				m.selected = m.cursor
				// Update viewport content for the selected neighbor
				m.detailsViewport.SetContent(m.formatNeighborDetails(m.neighbors[m.selected]))
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.selected != -1 {
			// Adjust viewport height
			headerHeight := 3 // Allow space for title and help
			footerHeight := 2
			m.detailsViewport.Width = msg.Width - 4 // Allow for borders
			m.detailsViewport.Height = msg.Height - headerHeight - footerHeight
			m.detailsViewport.SetContent(m.formatNeighborDetails(m.neighbors[m.selected]))
		}
	}

	// Handle viewport scrolling when in details view
	if m.selected != -1 {
		m.detailsViewport, cmd = m.detailsViewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the UI
func (m *QueryModel) View() string {
	if len(m.neighbors) == 0 {
		return "No results found"
	}

	if m.selected != -1 {
		// Render details view
		neighbor := m.neighbors[m.selected]
		title := titleStyle.Render(fmt.Sprintf(" Result Details: %v ", neighbor.Key))

		return fmt.Sprintf("%s\n%s\n%s",
			title,
			m.detailsViewport.View(),
			helpStyle("↑/↓: scroll • q: query with this record • esc: back • ctrl+c: quit"),
		)
	}

	// Render list view
	title := titleStyle.Render(" Query Results ")

	var b strings.Builder
	b.WriteString(title + "\n\n")

	// List neighbors
	for i, neighbor := range m.neighbors {
		var line string
		if i == m.cursor {
			prefix := selectedItemStyle.Render("")
			keyInfo := highlightedStyle.Render(fmt.Sprintf("%v", neighbor.Key))
			line = fmt.Sprintf("%s%s", prefix, keyInfo)
		} else {
			line = itemStyle.Render(fmt.Sprintf("%v", neighbor.Key))
		}

		// Add additional information
		setInfo := ""
		if neighbor.Set != nil {
			setInfo = fmt.Sprintf(" (Set: %s)", *neighbor.Set)
		}
		line += fmt.Sprintf("%s - Distance: %.4f", setInfo, neighbor.Distance)

		b.WriteString(line + "\n")
	}

	// Help text
	b.WriteString("\n" + helpStyle("↑/↓: navigate • enter: select • ctrl+c: quit"))
	b.WriteString("\n" + infoStyle.Render(fmt.Sprintf("Found %d results", len(m.neighbors))))
	b.WriteString("\n" + infoStyle.Render("Select any result to view details, then press 'q' to find similar records"))
	b.WriteString("\n" + infoStyle.Render("You can chain multiple queries to explore the vector space"))

	return b.String()
}

// formatNeighborDetails creates a formatted string with all neighbor details
func (m *QueryModel) formatNeighborDetails(neighbor *avs.Neighbor) string {
	var b strings.Builder

	// Basic information
	b.WriteString(fmt.Sprintf("Namespace: %s\n", neighbor.Namespace))

	if neighbor.Set != nil {
		b.WriteString(fmt.Sprintf("Set: %s\n", *neighbor.Set))
	}

	b.WriteString(fmt.Sprintf("Key: %v\n", neighbor.Key))
	b.WriteString(fmt.Sprintf("Distance: %.6f\n", neighbor.Distance))
	b.WriteString(fmt.Sprintf("Generation: %d\n", neighbor.Record.Generation))

	if neighbor.Record.Expiration != nil {
		b.WriteString(fmt.Sprintf("Expiration: %v\n", *neighbor.Record.Expiration))
	}

	b.WriteString("\n" + subtitleStyle.Render("Record Data:") + "\n\n")

	// Format the record data
	keys := make([]string, 0, len(neighbor.Record.Data))

	// If we have specific fields to include, use them
	if len(m.includeFields) > 0 {
		for _, key := range m.includeFields {
			if _, exists := neighbor.Record.Data[key]; exists {
				keys = append(keys, key)
			}
		}
	} else {
		// Otherwise get all keys
		for key := range neighbor.Record.Data {
			keys = append(keys, key)
		}
	}

	// Sort keys for consistent display
	sort.Strings(keys)

	// Display keys up to maxDataKeys if limit is set
	displayCount := len(keys)
	if m.maxDataKeys > 0 && m.maxDataKeys < displayCount {
		displayCount = m.maxDataKeys
	}

	for i, key := range keys {
		if i >= displayCount && m.maxDataKeys > 0 {
			b.WriteString("...\n")
			break
		}

		value := formatValueForDisplay(neighbor.Record.Data[key], m.maxDataColWidth)
		b.WriteString(fmt.Sprintf("%s: %s\n", highlightedStyle.Render(key), value))
	}

	return b.String()
}

// formatValueForDisplay formats a value for display, handling vectors and truncating if necessary
func formatValueForDisplay(value interface{}, maxWidth int) string {
	str := fmt.Sprintf("%v", value)

	// Handle vector display
	if val, ok := value.([]interface{}); ok {
		// It's a vector or array
		if len(val) > 10 {
			// Truncate large vectors
			strValues := make([]string, 10)
			for i := 0; i < 10; i++ {
				strValues[i] = fmt.Sprintf("%v", val[i])
			}
			str = fmt.Sprintf("[%s, ... (%d more)]", strings.Join(strValues, ", "), len(val)-10)
		}
	}

	// Truncate long strings
	if maxWidth > 0 && len(str) > maxWidth {
		return str[:maxWidth-3] + "..."
	}

	return str
}

// StartInteractiveQuery launches the interactive TUI for query results
func StartInteractiveQuery(
	neighbors []*avs.Neighbor,
	maxDataKeys int,
	maxDataColWidth int,
	includeFields []string,
) (*avs.Neighbor, error) {
	model := NewQueryModel(neighbors, maxDataKeys, maxDataColWidth, includeFields)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	if err != nil {
		return nil, err
	}

	// Check if the user requested a new query with a selected record
	if selected, ok := model.GetSelectedVectorForQuery(); ok {
		return selected, nil
	}

	return nil, nil
}

// GetSelectedVectorForQuery returns the selected neighbor's vector for a new query
func (m *QueryModel) GetSelectedVectorForQuery() (*avs.Neighbor, bool) {
	if m.selected == -1 || !m.queryRequested {
		return nil, false
	}

	return m.neighbors[m.selected], true
}
