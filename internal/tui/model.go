package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kjaniec-dev/porthole/internal/config"
	"github.com/kjaniec-dev/porthole/internal/docker"
	"github.com/kjaniec-dev/porthole/internal/provider"
)

type tab int

const (
	tabRouters tab = iota
	tabServices
	tabCerts
	tabContainers
	tabCount
)

var tabNames = []string{"Routers", "Services", "Certs", "Containers"}

type tickMsg time.Time
type dataMsg struct {
	routers    []provider.Router
	services   []provider.Service
	certs      []provider.Certificate
	containers []docker.Container
	err        error
}

type Model struct {
	cfg          *config.Config
	prov         provider.Provider
	pollInterval time.Duration
	dockerClient *docker.Client
	dockerErr    error

	activeTab  tab
	width      int
	height     int
	filter     string
	filtering  bool
	showHelp   bool
	spinner    spinner.Model
	loading    bool
	lastErr    error
	lastUpdate time.Time

	routers    []provider.Router
	services   []provider.Service
	certs      []provider.Certificate
	containers []docker.Container

	routerTable    *Table
	serviceTable   *Table
	certTable      *Table
	containerTable *Table
}

func New(cfg *config.Config) *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorBlue)

	tableHeight := 20

	var prov provider.Provider
	var pollInterval time.Duration
	if cfg.Caddy.URL != "" {
		prov = provider.NewCaddyClient(cfg.Caddy)
		pollInterval = cfg.Caddy.PollInterval
	} else {
		prov = provider.NewTraefikClient(cfg.Traefik)
		pollInterval = cfg.Traefik.PollInterval
	}

	m := &Model{
		cfg:          cfg,
		prov:         prov,
		pollInterval: pollInterval,
		loading:      true,
		spinner:      sp,

		routerTable: NewTable([]Column{
			{Title: "RULE", Width: 40},
			{Title: "SERVICE", Width: 20},
			{Title: "TLS", Width: 5},
			{Title: "MIDDLEWARES", Width: 30},
			{Title: "STATUS", Width: 10},
		}, tableHeight),

		serviceTable: NewTable([]Column{
			{Title: "NAME", Width: 35},
			{Title: "TYPE", Width: 15},
			{Title: "SERVERS", Width: 30},
			{Title: "STATUS", Width: 10},
		}, tableHeight),

		certTable: NewTable([]Column{
			{Title: "DOMAIN", Width: 35},
			{Title: "ISSUER", Width: 25},
			{Title: "EXPIRES", Width: 20},
			{Title: "STATUS", Width: 12},
		}, tableHeight),

		containerTable: NewTable([]Column{
			{Title: "NAME", Width: 30},
			{Title: "IMAGE", Width: 30},
			{Title: "STATE", Width: 10},
			{Title: "UPTIME", Width: 15},
			{Title: "RESTARTS", Width: 10},
		}, tableHeight),
	}

	dc, err := docker.NewClient(cfg.Docker.Socket)
	if err != nil {
		m.dockerErr = err
	} else {
		m.dockerClient = dc
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchData(),
		m.pollTick(),
	)
}

func (m *Model) pollTick() tea.Cmd {
	return tea.Tick(m.pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		msg := dataMsg{}

		routers, err := m.prov.Routers()
		if err != nil {
			msg.err = err
			return msg
		}
		msg.routers = routers

		services, err := m.prov.Services()
		if err != nil {
			msg.err = err
			return msg
		}
		msg.services = services

		certs, _ := m.prov.Certificates()
		msg.certs = certs

		if m.dockerClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			containers, err := m.dockerClient.Containers(ctx)
			if err == nil {
				msg.containers = containers
			}
		}

		return msg
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeTables()

	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterKey(msg)
		}
		return m.handleKey(msg)

	case tickMsg:
		cmds = append(cmds, m.fetchData(), m.pollTick())

	case dataMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		if msg.err != nil {
			m.lastErr = msg.err
		} else {
			m.lastErr = nil
			m.routers = msg.routers
			m.services = msg.services
			m.certs = msg.certs
			m.containers = msg.containers
			m.rebuildTables()
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab", "right", "l":
		m.activeTab = (m.activeTab + 1) % tabCount
	case "shift+tab", "left", "h":
		m.activeTab = (m.activeTab + tabCount - 1) % tabCount
	case "1":
		m.activeTab = tabRouters
	case "2":
		m.activeTab = tabServices
	case "3":
		m.activeTab = tabCerts
	case "4":
		m.activeTab = tabContainers
	case "up", "k":
		m.activeTable().MoveUp()
	case "down", "j":
		m.activeTable().MoveDown()
	case "r":
		m.loading = true
		return m, m.fetchData()
	case "/":
		m.filtering = true
		m.filter = ""
	case "?":
		m.showHelp = !m.showHelp
	case "esc":
		m.showHelp = false
		m.filter = ""
	}
	return m, nil
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.filtering = false
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildTables()
		}
	default:
		if len(msg.String()) == 1 {
			m.filter += msg.String()
			m.rebuildTables()
		}
	}
	return m, nil
}

func (m *Model) activeTable() *Table {
	switch m.activeTab {
	case tabRouters:
		return m.routerTable
	case tabServices:
		return m.serviceTable
	case tabCerts:
		return m.certTable
	case tabContainers:
		return m.containerTable
	}
	return m.routerTable
}

func (m *Model) resizeTables() {
	contentHeight := m.height - 7
	if contentHeight < 5 {
		contentHeight = 5
	}
	for _, t := range []*Table{m.routerTable, m.serviceTable, m.certTable, m.containerTable} {
		t.Height = contentHeight
	}
}

func (m *Model) rebuildTables() {
	filter := strings.ToLower(m.filter)
	m.buildRouterRows(filter)
	m.buildServiceRows(filter)
	m.buildCertRows(filter)
	m.buildContainerRows(filter)
}

func (m *Model) buildRouterRows(filter string) {
	var rows []Row
	for _, r := range m.routers {
		tls := "—"
		if r.TLS != nil {
			tls = styleCellOK.Render("✓")
		}
		status := styleStatus(r.Status)
		mw := strings.Join(r.Middlewares, ", ")
		row := Row{r.Rule, r.Service, tls, mw, status}
		if filter == "" || rowMatchesFilter(row, filter) {
			rows = append(rows, row)
		}
	}
	m.routerTable.SetRows(rows)
}

func (m *Model) buildServiceRows(filter string) {
	var rows []Row
	for _, s := range m.services {
		var servers []string
		if s.LoadBalancer != nil {
			for _, srv := range s.LoadBalancer.Servers {
				servers = append(servers, srv.URL)
			}
		}
		status := styleStatus(s.Status)
		row := Row{s.Name, s.Type, strings.Join(servers, ", "), status}
		if filter == "" || rowMatchesFilter(row, filter) {
			rows = append(rows, row)
		}
	}
	m.serviceTable.SetRows(rows)
}

func (m *Model) buildCertRows(filter string) {
	now := time.Now()
	var rows []Row
	for _, c := range m.certs {
		domain := c.Domain
		if domain == "" && c.Subject.CommonName != "" {
			domain = c.Subject.CommonName
		}
		issuer := c.Issuer.CommonName
		remaining := c.NotAfter.Sub(now)
		days := int(remaining.Hours() / 24)
		expiresStr := fmt.Sprintf("%s (%dd)", c.NotAfter.Format("2006-01-02"), days)

		var certStatus string
		switch {
		case days < 0:
			certStatus = styleCellErr.Render("EXPIRED")
		case days < 30:
			certStatus = styleCellErr.Render(fmt.Sprintf("! %dd", days))
		case days < 90:
			certStatus = styleCellWarn.Render(fmt.Sprintf("~ %dd", days))
		default:
			certStatus = styleCellOK.Render(fmt.Sprintf("✓ %dd", days))
		}

		row := Row{domain, issuer, expiresStr, certStatus}
		if filter == "" || rowMatchesFilter(row, filter) {
			rows = append(rows, row)
		}
	}
	m.certTable.SetRows(rows)
}

func (m *Model) buildContainerRows(filter string) {
	now := time.Now()
	var rows []Row
	for _, c := range m.containers {
		uptime := "—"
		if !c.Started.IsZero() {
			uptime = formatDuration(now.Sub(c.Started))
		}
		state := styleContainerState(c.State)
		restarts := fmt.Sprintf("%d", c.Restarts)
		if c.Restarts > 5 {
			restarts = styleCellErr.Render(restarts)
		} else if c.Restarts > 0 {
			restarts = styleCellWarn.Render(restarts)
		}
		row := Row{c.Name, c.Image, state, uptime, restarts}
		if filter == "" || rowMatchesFilter(row, filter) {
			rows = append(rows, row)
		}
	}
	m.containerTable.SetRows(rows)
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.helpView()
	}

	var b strings.Builder

	// Tab bar
	b.WriteString(m.tabBarView())
	b.WriteString("\n")

	// Content
	b.WriteString(m.contentView())

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.statusBarView())

	return b.String()
}

func (m *Model) tabBarView() string {
	var tabs []string
	for i, name := range tabNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		if tab(i) == m.activeTab {
			tabs = append(tabs, styleTabActive.Render(label))
		} else {
			tabs = append(tabs, styleTabInactive.Render(label))
		}
	}

	title := styleCellBlue.Bold(true).Render("⚓ porthole")
	bar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	gap := strings.Repeat(" ", max(0, m.width-lipgloss.Width(bar)-lipgloss.Width(title)-2))
	return styleTabBar.Width(m.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top, bar, gap, title),
	)
}

func (m *Model) contentView() string {
	var content string

	if m.loading {
		content = fmt.Sprintf("\n  %s Fetching data…", m.spinner.View())
	} else if m.lastErr != nil {
		content = styleCellErr.Render(fmt.Sprintf("\n  Error: %s\n\n  Press r to retry.", m.lastErr))
	} else {
		switch m.activeTab {
		case tabRouters:
			content = m.routerTable.View()
		case tabServices:
			content = m.serviceTable.View()
		case tabCerts:
			content = m.certTable.View()
		case tabContainers:
			if m.dockerErr != nil {
				content = styleCellWarn.Render(fmt.Sprintf("  Docker unavailable: %s", m.dockerErr))
			} else {
				content = m.containerTable.View()
			}
		}
	}

	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Render(content)
}

func (m *Model) statusBarView() string {
	left := ""
	if m.filtering {
		left = fmt.Sprintf(" /%s█", m.filter)
	} else if m.filter != "" {
		left = fmt.Sprintf(" filter: %s", m.filter)
	} else {
		left = " ↑↓/jk move  tab next  / filter  r refresh  ? help  q quit"
	}

	right := ""
	if !m.lastUpdate.IsZero() {
		right = fmt.Sprintf("updated %s ", m.lastUpdate.Format("15:04:05"))
	}

	gap := strings.Repeat(" ", max(0, m.width-len(left)-len(right)-2))
	return styleStatusBar.Width(m.width).Render(left + gap + right)
}

func (m *Model) helpView() string {
	help := `
  ⚓ porthole — keybindings

  Navigation
    tab / shift+tab    Next / previous tab
    1-4                Jump to tab
    j / k / ↑↓         Move selection up/down

  Actions
    r                  Force refresh
    /                  Filter rows (type to search, Enter/Esc to exit)
    Esc                Clear filter / close overlays
    q / ctrl+c         Quit

  Tabs
    1  Routers     — HTTP routing rules, TLS, middlewares
    2  Services    — Backend services and load balancer info
    3  Certs       — TLS certificate expiry with color indicators
    4  Containers  — Docker container state, uptime, restart count

  Cert expiry colours
    ` + styleCellOK.Render("green") + `   > 90 days remaining
    ` + styleCellWarn.Render("yellow") + `  30–90 days remaining
    ` + styleCellErr.Render("red") + `     < 30 days or expired

  Press ? or Esc to close this help.
`
	return styleBorder.
		Width(m.width - 4).
		Margin(2, 2).
		Render(help)
}

func styleStatus(s string) string {
	switch strings.ToLower(s) {
	case "enabled":
		return styleCellOK.Render("enabled")
	case "disabled":
		return styleCellGray.Render("disabled")
	default:
		return styleCellErr.Render(s)
	}
}

func styleContainerState(s string) string {
	switch strings.ToLower(s) {
	case "running":
		return styleCellOK.Render("running")
	case "exited":
		return styleCellErr.Render("exited")
	case "paused":
		return styleCellWarn.Render("paused")
	case "restarting":
		return styleCellWarn.Render("restarting")
	default:
		return styleCellGray.Render(s)
	}
}

func rowMatchesFilter(row Row, filter string) bool {
	for _, cell := range row {
		if strings.Contains(strings.ToLower(cell), filter) {
			return true
		}
	}
	return false
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm%ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
