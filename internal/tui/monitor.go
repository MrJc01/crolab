// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crolab/core/internal/cli"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1a2e")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0")).
			Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ADE80"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171"))

	onlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ADE80")).Bold(true)

	offlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171")).Bold(true)

	logPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 1)

	logTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#38BDF8"))
)

// --- Messages ---

type pingResultMsg struct {
	name   string
	online bool
	latMs  int64
}

type tickMsg struct{}

// --- Model ---

type viewMode int

const (
	modeNormal viewMode = iota
	modeAdd
)

type Model struct {
	servers     []cli.ServerConfig
	defaultName string
	cursor      int
	width       int
	height      int
	message     string
	messageErr  bool
	mode        viewMode
	pingStatus  map[string]string // "online 12ms" or "offline"
	logs        []string          // recent log lines

	// Add form
	addInputs []textinput.Model
	addFocus  int
}

func NewModel() Model {
	servers, defaultName, _ := cli.ListServers()

	// Create text inputs for add form
	nameInput := textinput.New()
	nameInput.Placeholder = "nome (ex: meu-gpu)"
	nameInput.CharLimit = 30

	addrInput := textinput.New()
	addrInput.Placeholder = "ip:porta (ex: 192.168.1.5:4422)"
	addrInput.CharLimit = 50

	tokenInput := textinput.New()
	tokenInput.Placeholder = "token (ou vazio)"
	tokenInput.CharLimit = 64

	provInput := textinput.New()
	provInput.Placeholder = "provider (local/vastai/runpod)"
	provInput.CharLimit = 20

	return Model{
		servers:     servers,
		defaultName: defaultName,
		pingStatus:  make(map[string]string),
		logs:        []string{"Monitor iniciado."},
		addInputs:   []textinput.Model{nameInput, addrInput, tokenInput, provInput},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(pingAllCmd(m.servers), tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func pingAllCmd(servers []cli.ServerConfig) tea.Cmd {
	var cmds []tea.Cmd
	for _, s := range servers {
		s := s
		cmds = append(cmds, func() tea.Msg {
			return pingServer(s)
		})
	}
	return tea.Batch(cmds...)
}

func pingServer(s cli.ServerConfig) pingResultMsg {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, s.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return pingResultMsg{name: s.Name, online: false}
	}
	conn.Close()
	latMs := time.Since(start).Milliseconds()
	return pingResultMsg{name: s.Name, online: true, latMs: latMs}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case pingResultMsg:
		if msg.online {
			m.pingStatus[msg.name] = fmt.Sprintf("● %dms", msg.latMs)
		} else {
			m.pingStatus[msg.name] = "○ offline"
		}

	case tickMsg:
		return m, tea.Batch(pingAllCmd(m.servers), tickCmd())

	case tea.KeyMsg:
		if m.mode == modeAdd {
			return m.updateAddMode(msg)
		}
		return m.updateNormalMode(msg)
	}

	return m, nil
}

func (m Model) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
		return m, tea.Quit

	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.servers)-1 {
			m.cursor++
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if len(m.servers) > 0 {
			name := m.servers[m.cursor].Name
			cli.SetDefault(name)
			m.defaultName = name
			m.message = fmt.Sprintf("Default → %s", name)
			m.messageErr = false
			m.addLog("Default alterado para " + name)
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("d", "delete"))):
		if len(m.servers) > 0 {
			name := m.servers[m.cursor].Name
			if err := cli.RemoveServer(name); err != nil {
				m.message = fmt.Sprintf("Erro: %v", err)
				m.messageErr = true
			} else {
				m.addLog("Removido: " + name)
				m.message = fmt.Sprintf("Removido: %s", name)
				m.messageErr = false
				m.servers, m.defaultName, _ = cli.ListServers()
				if m.cursor >= len(m.servers) && m.cursor > 0 {
					m.cursor--
				}
			}
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
		m.mode = modeAdd
		m.addFocus = 0
		for i := range m.addInputs {
			m.addInputs[i].Reset()
		}
		m.addInputs[0].Focus()
		return m, m.addInputs[0].Cursor.BlinkCmd()

	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		m.servers, m.defaultName, _ = cli.ListServers()
		m.message = "Lista atualizada"
		m.messageErr = false
		return m, pingAllCmd(m.servers)
	}

	return m, nil
}

func (m Model) updateAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.mode = modeNormal
		return m, nil

	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
		m.addInputs[m.addFocus].Blur()
		m.addFocus = (m.addFocus + 1) % len(m.addInputs)
		m.addInputs[m.addFocus].Focus()
		return m, m.addInputs[m.addFocus].Cursor.BlinkCmd()

	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		name := m.addInputs[0].Value()
		addr := m.addInputs[1].Value()
		token := m.addInputs[2].Value()
		prov := m.addInputs[3].Value()
		if prov == "" {
			prov = "local"
		}

		if name == "" || addr == "" {
			m.message = "Nome e Endereço são obrigatórios"
			m.messageErr = true
			return m, nil
		}

		if err := cli.AddServer(name, addr, token, prov, 1); err != nil {
			m.message = fmt.Sprintf("Erro: %v", err)
			m.messageErr = true
		} else {
			m.addLog("Adicionado: " + name + " (" + addr + ")")
			m.message = fmt.Sprintf("✓ %s adicionado", name)
			m.messageErr = false
			m.servers, m.defaultName, _ = cli.ListServers()
		}
		m.mode = modeNormal
		return m, pingAllCmd(m.servers)
	}

	// Forward to focused input
	var cmd tea.Cmd
	m.addInputs[m.addFocus], cmd = m.addInputs[m.addFocus].Update(msg)
	return m, cmd
}

func (m *Model) addLog(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > 12 {
		m.logs = m.logs[len(m.logs)-12:]
	}
}

func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("  CROLAB MONITOR  ") + "\n\n")

	if m.mode == modeAdd {
		return b.String() + m.viewAddForm()
	}

	// Header
	header := fmt.Sprintf("  %-3s %-16s %-10s %-5s %-22s %s",
		"", "NOME", "PROVIDER", "PRIO", "ENDEREÇO", "STATUS")
	b.WriteString(headerStyle.Render(header) + "\n")
	b.WriteString(dimStyle.Render("  "+strings.Repeat("─", 72)) + "\n")

	if len(m.servers) == 0 {
		b.WriteString(dimStyle.Render("\n  Nenhum servidor. Pressione [A] para adicionar.\n"))
	}

	for i, s := range m.servers {
		marker := "  "
		if s.Name == m.defaultName {
			marker = "★ "
		}

		statusStr := dimStyle.Render("...")
		if ps, ok := m.pingStatus[s.Name]; ok {
			if strings.Contains(ps, "offline") {
				statusStr = offlineStyle.Render(ps)
			} else {
				statusStr = onlineStyle.Render(ps)
			}
		}

		line := fmt.Sprintf("%s%-16s %-10s %-5d %-22s",
			marker, s.Name, s.Provider, s.Priority, s.Address)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render(line) + " " + statusStr + "\n")
		} else {
			b.WriteString(normalStyle.Render(line) + " " + statusStr + "\n")
		}
	}

	// Message
	b.WriteString("\n")
	if m.message != "" {
		if m.messageErr {
			b.WriteString(errorStyle.Render("  ✗ "+m.message) + "\n")
		} else {
			b.WriteString(successStyle.Render("  ✓ "+m.message) + "\n")
		}
	}

	// Log panel
	b.WriteString("\n")
	logTitle := logTitleStyle.Render("  Logs")
	b.WriteString(logTitle + "\n")
	b.WriteString(dimStyle.Render("  "+strings.Repeat("─", 50)) + "\n")
	startIdx := 0
	if len(m.logs) > 6 {
		startIdx = len(m.logs) - 6
	}
	for _, l := range m.logs[startIdx:] {
		b.WriteString(dimStyle.Render("  │ "+l) + "\n")
	}

	// Status bar
	b.WriteString("\n")
	slots := fmt.Sprintf("Servers: %d", len(m.servers))
	help := "↑↓ navegar  Enter=default  D=remover  A=adicionar  R=refresh  Q=sair"
	b.WriteString(statusBarStyle.Render(fmt.Sprintf("  %s  │  %s  ", slots, help)))

	return b.String()
}

func (m Model) viewAddForm() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("  Adicionar Servidor") + "\n\n")

	labels := []string{"Nome:", "Endereço:", "Token:", "Provider:"}
	for i, label := range labels {
		b.WriteString(normalStyle.Render("  "+label) + "\n")
		b.WriteString("  " + m.addInputs[i].View() + "\n\n")
	}

	b.WriteString(dimStyle.Render("  Tab=próximo campo  Enter=salvar  Esc=cancelar"))
	return b.String()
}

// Run starts the TUI monitor.
func Run() error {
	cli.InitConfig()
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
