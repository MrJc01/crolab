// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/crolab/core/internal/cli"
)

var (
	selectorTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED"))

	selectorSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1a1a2e")).
				Background(lipgloss.Color("#4ADE80")).
				Padding(0, 1)

	selectorNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E2E8F0")).
				Padding(0, 1)

	selectorDimStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#64748B"))
)

type SelectorModel struct {
	servers  []cli.ServerConfig
	cursor   int
	chosen   *cli.ServerConfig
	canceled bool
}

func NewSelectorModel(servers []cli.ServerConfig) SelectorModel {
	return SelectorModel{servers: servers}
}

func (m SelectorModel) Init() tea.Cmd { return nil }

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if m.cursor < len(m.servers)-1 {
				m.cursor++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			s := m.servers[m.cursor]
			m.chosen = &s
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			m.canceled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SelectorModel) View() string {
	s := selectorTitleStyle.Render("Selecione o servidor destino:") + "\n\n"

	for i, srv := range m.servers {
		line := fmt.Sprintf("%-16s %-10s prio:%-3d %s", srv.Name, srv.Provider, srv.Priority, srv.Address)
		if i == m.cursor {
			s += selectorSelectedStyle.Render("▸ "+line) + "\n"
		} else {
			s += selectorNormalStyle.Render("  "+line) + "\n"
		}
	}

	s += "\n" + selectorDimStyle.Render("Enter=selecionar  Esc=cancelar")
	return s
}

// SelectServer presents an interactive TUI selector and returns the chosen server.
// Returns nil if the user cancels.
func SelectServer(servers []cli.ServerConfig) *cli.ServerConfig {
	m := NewSelectorModel(servers)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil
	}
	final := result.(SelectorModel)
	if final.canceled {
		return nil
	}
	return final.chosen
}
