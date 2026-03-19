// Package tui implements the Bubbletea terminal UI for Engram.
//
// Following the Gentleman Bubbletea patterns:
// - Screen constants as iota
// - Single Model struct holds ALL state
// - Update() with type switch
// - Per-screen key handlers returning (tea.Model, tea.Cmd)
// - Vim keys (j/k) for navigation
// - PrevScreen for back navigation
package tui

import (
	"time"

	"github.com/Gentleman-Programming/engram/internal/setup"
	"github.com/Gentleman-Programming/engram/internal/store"
	"github.com/Gentleman-Programming/engram/internal/version"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Screens ─────────────────────────────────────────────────────────────────

type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenSearch
	ScreenSearchResults
	ScreenRecent
	ScreenObservationDetail
	ScreenTimeline
	ScreenSessions
	ScreenSessionDetail
	ScreenProjects
	ScreenProjectDetail
	ScreenSetup
)

// ─── Confirmation Dialog Actions ─────────────────────────────────────────────

type ConfirmAction int

const (
	ConfirmNone ConfirmAction = iota
	ConfirmDeleteSession
	ConfirmClearProjectSessions
	ConfirmDeleteProject
)

// ─── Custom Messages ─────────────────────────────────────────────────────────

type updateAvailableMsg struct {
	msg string
}

type statsLoadedMsg struct {
	stats *store.Stats
	err   error
}

type searchResultsMsg struct {
	results []store.SearchResult
	query   string
	err     error
}

type recentObservationsMsg struct {
	observations []store.Observation
	err          error
}

type observationDetailMsg struct {
	observation *store.Observation
	err         error
}

type timelineMsg struct {
	timeline *store.TimelineResult
	err      error
}

type recentSessionsMsg struct {
	sessions []store.SessionSummary
	err      error
}

type sessionObservationsMsg struct {
	observations []store.Observation
	err          error
}

type setupInstallMsg struct {
	result *setup.Result
	err    error
}

type sessionDeletedMsg struct {
	err error
}

type projectSessionsClearedMsg struct {
	result *store.DeleteResult
	err    error
}

type projectDeletedMsg struct {
	result *store.DeleteResult
	err    error
}

type successClearMsg struct{}

type projectsLoadedMsg struct {
	projects []store.ProjectStats
	err      error
}

type projectDetailSessionsMsg struct {
	sessions []store.SessionSummary
	err      error
}

type filteredProjectsMsg struct {
	projects []store.ProjectStats
	err      error
}

type filteredSessionsMsg struct {
	sessions []store.SessionSummary
	err      error
}

// ─── Model ───────────────────────────────────────────────────────────────────

type Model struct {
	store      *store.Store
	Version    string
	Screen     Screen
	PrevScreen Screen
	Width      int
	Height     int
	Cursor     int
	Scroll     int

	UpdateMsg  string
	ErrorMsg   string
	SuccessMsg string

	ConfirmActive bool
	ConfirmAction ConfirmAction
	ConfirmMsg    string
	ConfirmDetail string
	ConfirmTarget string

	// Dashboard
	Stats *store.Stats

	// Search
	SearchInput   textinput.Model
	SearchQuery   string
	SearchResults []store.SearchResult

	// Recent observations
	RecentObservations []store.Observation

	// Observation detail
	SelectedObservation *store.Observation
	DetailScroll        int

	// Timeline
	Timeline *store.TimelineResult

	// Sessions
	Sessions            []store.SessionSummary
	SelectedSessionIdx  int
	SessionObservations []store.Observation
	SessionDetailScroll int

	// Projects
	Projects            []store.ProjectStats
	SelectedProjectIdx  int
	ProjectSessions     []store.SessionSummary
	ProjectDetailScroll int

	// Inline filter (shared across Projects, Sessions, ProjectDetail)
	FilterInput  textinput.Model
	FilterActive bool
	FilterQuery  string

	// Setup
	SetupAgents           []setup.Agent
	SetupResult           *setup.Result
	SetupError            string
	SetupDone             bool
	SetupInstalling       bool
	SetupInstallingName   string // agent name being installed (for display)
	SetupAllowlistPrompt  bool   // true = showing y/n prompt for allowlist
	SetupAllowlistApplied bool   // true = allowlist was added successfully
	SetupAllowlistError   string // error message if allowlist injection failed
	SetupSpinner          spinner.Model
}

// New creates a new TUI model connected to the given store.
func New(s *store.Store, version string) Model {
	ti := textinput.New()
	ti.Placeholder = "Search memories..."
	ti.CharLimit = 256
	ti.Width = 60

	fi := textinput.New()
	fi.Placeholder = "Filter..."
	fi.CharLimit = 128
	fi.Width = 40

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorLavender)

	return Model{
		store:        s,
		Version:      version,
		Screen:       ScreenDashboard,
		SearchInput:  ti,
		FilterInput:  fi,
		SetupSpinner: sp,
	}
}

// Init loads initial data (stats for the dashboard).
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadStats(m.store),
		checkForUpdate(m.Version),
		tea.EnterAltScreen,
	)
}

// ─── Commands (data loading) ─────────────────────────────────────────────────

func checkForUpdate(v string) tea.Cmd {
	return func() tea.Msg {
		return updateAvailableMsg{msg: version.CheckLatest(v)}
	}
}

func loadStats(s *store.Store) tea.Cmd {
	return func() tea.Msg {
		stats, err := s.Stats()
		return statsLoadedMsg{stats: stats, err: err}
	}
}

func searchMemories(s *store.Store, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := s.Search(query, store.SearchOptions{Limit: 50})
		return searchResultsMsg{results: results, query: query, err: err}
	}
}

func loadRecentObservations(s *store.Store) tea.Cmd {
	return func() tea.Msg {
		obs, err := s.AllObservations("", "", 50)
		return recentObservationsMsg{observations: obs, err: err}
	}
}

func loadObservationDetail(s *store.Store, id int64) tea.Cmd {
	return func() tea.Msg {
		obs, err := s.GetObservation(id)
		return observationDetailMsg{observation: obs, err: err}
	}
}

func loadTimeline(s *store.Store, obsID int64) tea.Cmd {
	return func() tea.Msg {
		tl, err := s.Timeline(obsID, 10, 10)
		return timelineMsg{timeline: tl, err: err}
	}
}

func loadRecentSessions(s *store.Store) tea.Cmd {
	return func() tea.Msg {
		sessions, err := s.AllSessions("", 50)
		return recentSessionsMsg{sessions: sessions, err: err}
	}
}

func loadSessionObservations(s *store.Store, sessionID string) tea.Cmd {
	return func() tea.Msg {
		obs, err := s.SessionObservations(sessionID, 200)
		return sessionObservationsMsg{observations: obs, err: err}
	}
}

func installAgent(agentName string) tea.Cmd {
	return func() tea.Msg {
		result, err := installAgentFn(agentName)
		return setupInstallMsg{result: result, err: err}
	}
}

func deleteSessionCmd(s *store.Store, id string) tea.Cmd {
	return func() tea.Msg {
		err := s.DeleteSession(id)
		return sessionDeletedMsg{err: err}
	}
}

func clearProjectSessionsCmd(s *store.Store, project string) tea.Cmd {
	return func() tea.Msg {
		result, err := s.ClearProjectSessions(project)
		return projectSessionsClearedMsg{result: result, err: err}
	}
}

func deleteProjectCmd(s *store.Store, project string) tea.Cmd {
	return func() tea.Msg {
		result, err := s.DeleteProject(project)
		return projectDeletedMsg{result: result, err: err}
	}
}

func clearSuccessAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return successClearMsg{}
	})
}

func loadProjects(s *store.Store) tea.Cmd {
	return func() tea.Msg {
		projects, err := s.ListProjectStats()
		return projectsLoadedMsg{projects: projects, err: err}
	}
}

func loadProjectSessions(s *store.Store, project string, limit int) tea.Cmd {
	return func() tea.Msg {
		sessions, err := s.ProjectSessions(project, limit)
		return projectDetailSessionsMsg{sessions: sessions, err: err}
	}
}

func filterProjects(s *store.Store, query string) tea.Cmd {
	return func() tea.Msg {
		projects, err := s.FilterProjects(query)
		return filteredProjectsMsg{projects: projects, err: err}
	}
}

func filterSessions(s *store.Store, query string, project string, limit int) tea.Cmd {
	return func() tea.Msg {
		sessions, err := s.FilterSessions(query, project, limit)
		return filteredSessionsMsg{sessions: sessions, err: err}
	}
}

var installAgentFn = setup.Install
var addClaudeCodeAllowlistFn = setup.AddClaudeCodeAllowlist
