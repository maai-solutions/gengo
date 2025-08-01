package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	input   string
	cursor  int
	history []string
}

func initialModel() model {
	return model{
		input:   "",
		cursor:  0,
		history: []string{},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			command := strings.TrimSpace(m.input)
			if command != "" {
				m.history = append(m.history, fmt.Sprintf("> %s", command))
				
				// Handle commands
				switch command {
				case "/exit":
					return m, tea.Quit
				default:
					m.history = append(m.history, fmt.Sprintf("Unknown command: %s", command))
					m.history = append(m.history, "Available commands: /exit")
				}
			}
			m.input = ""
			m.cursor = 0
		case "backspace":
			if m.cursor > 0 {
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.input) {
				m.cursor++
			}
		default:
			// Insert character at cursor position
			if len(msg.String()) == 1 {
				m.input = m.input[:m.cursor] + msg.String() + m.input[m.cursor:]
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var s strings.Builder
	
	s.WriteString("GenGo Interactive CLI\n")
	s.WriteString("Type '/exit' to quit or 'Ctrl+C'\n")
	s.WriteString(strings.Repeat("─", 40) + "\n\n")
	
	// Show command history
	for _, line := range m.history {
		s.WriteString(line + "\n")
	}
	
	// Show current input with cursor
	s.WriteString("> ")
	for i, r := range m.input {
		if i == m.cursor {
			s.WriteString("│")
		}
		s.WriteString(string(r))
	}
	if m.cursor >= len(m.input) {
		s.WriteString("│")
	}
	
	return s.String()
}
