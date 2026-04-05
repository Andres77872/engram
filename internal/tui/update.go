package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Gentleman-Programming/engram/internal/setup"
	"github.com/Gentleman-Programming/engram/internal/store"
)

// ─── Update ──────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit — always works
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// If search input is focused, let it handle most keys
		if m.Screen == ScreenSearch && m.SearchInput.Focused() {
			return m.handleSearchInputKeys(msg)
		}
		// If filter input is focused, let it handle most keys
		if m.FilterActive && m.FilterInput.Focused() {
			return m.handleFilterInputKeys(msg)
		}
		return m.handleKeyPress(msg.String())

	// ─── Data loaded messages ────────────────────────────────────────────
	case updateCheckMsg:
		m.UpdateStatus = msg.result.Status
		m.UpdateMsg = msg.result.Message
		return m, nil

	case statsLoadedMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.Stats = msg.stats
		return m, nil

	case searchResultsMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.SearchResults = msg.results
		m.SearchQuery = msg.query
		m.Screen = ScreenSearchResults
		m.Cursor = 0
		m.Scroll = 0
		return m, nil

	case recentObservationsMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.RecentObservations = msg.observations
		return m, nil

	case observationDetailMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.SelectedObservation = msg.observation
		m.Screen = ScreenObservationDetail
		m.DetailScroll = 0
		return m, nil

	case timelineMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.Timeline = msg.timeline
		m.Screen = ScreenTimeline
		m.Scroll = 0
		return m, nil

	case recentSessionsMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.Sessions = msg.sessions
		return m, nil

	case sessionObservationsMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.SessionObservations = msg.observations
		m.Screen = ScreenSessionDetail
		m.Cursor = 0
		m.SessionDetailScroll = 0
		return m, nil

	case setupInstallMsg:
		m.SetupInstalling = false
		if msg.err != nil {
			m.SetupDone = true
			m.SetupError = msg.err.Error()
			return m, nil
		}
		m.SetupResult = msg.result
		m.SetupError = ""
		// For claude-code, show allowlist prompt before marking done
		if msg.result != nil && msg.result.Agent == "claude-code" {
			m.SetupAllowlistPrompt = true
			return m, nil
		}
		m.SetupDone = true
		return m, nil

	case spinner.TickMsg:
		if m.SetupInstalling {
			var cmd tea.Cmd
			m.SetupSpinner, cmd = m.SetupSpinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case sessionDeletedMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.ErrorMsg = ""
		m.SuccessMsg = "✓ Session deleted"
		m.ConfirmActive = false
		m.Screen = ScreenSessions
		m.Cursor = 0
		m.resetFilter()
		return m, tea.Batch(loadRecentSessions(m.store), clearSuccessAfterDelay())

	case projectSessionsClearedMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.ErrorMsg = ""
		if msg.result != nil {
			m.SuccessMsg = fmt.Sprintf("✓ %d sessions, %d observations cleared", msg.result.SessionsDeleted, msg.result.ObservationsDeleted)
		} else {
			m.SuccessMsg = "✓ Project sessions cleared"
		}
		m.ConfirmActive = false
		m.Screen = ScreenDashboard
		m.Cursor = 0
		m.resetFilter()
		return m, tea.Batch(loadStats(m.store), clearSuccessAfterDelay())

	case projectDeletedMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.ErrorMsg = ""
		if msg.result != nil {
			m.SuccessMsg = fmt.Sprintf("✓ Project deleted: %d sessions, %d observations removed", msg.result.SessionsDeleted, msg.result.ObservationsDeleted)
		} else {
			m.SuccessMsg = "✓ Project deleted"
		}
		m.ConfirmActive = false
		m.Screen = ScreenDashboard
		m.Cursor = 0
		m.resetFilter()
		return m, tea.Batch(loadStats(m.store), clearSuccessAfterDelay())

	case emptySessionsDeletedMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.ErrorMsg = ""
		if msg.result != nil {
			m.SuccessMsg = fmt.Sprintf("✓ %d empty sessions cleared", msg.result.SessionsDeleted)
		} else {
			m.SuccessMsg = "✓ Empty sessions cleared"
		}
		m.ConfirmActive = false
		m.Cursor = 0
		m.Scroll = 0
		m.ProjectDetailScroll = 0
		m.resetFilter()
		// Guard: if on ProjectDetail with stale SelectedProjectIdx, navigate back to Projects
		if m.Screen == ScreenProjectDetail && m.SelectedProjectIdx >= len(m.Projects) {
			m.Screen = ScreenProjects
			m.SelectedProjectIdx = 0
		}
		// Always reload projects and stats after deletion to keep counts aligned
		return m, tea.Batch(m.reloadCurrentScreen(), loadProjects(m.store), loadStats(m.store), clearSuccessAfterDelay())

	case projectsLoadedMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.Projects = msg.projects
		return m, nil

	case projectDetailSessionsMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.ProjectSessions = msg.sessions
		m.Screen = ScreenProjectDetail
		m.Cursor = 0
		m.ProjectDetailScroll = 0
		return m, nil

	case filteredProjectsMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		m.Projects = msg.projects
		m.Cursor = 0
		m.Scroll = 0
		return m, nil

	case filteredSessionsMsg:
		if msg.err != nil {
			m.ErrorMsg = msg.err.Error()
			return m, nil
		}
		if m.Screen == ScreenSessions {
			m.Sessions = msg.sessions
		} else if m.Screen == ScreenProjectDetail {
			m.ProjectSessions = msg.sessions
		}
		m.Cursor = 0
		m.Scroll = 0
		m.ProjectDetailScroll = 0
		return m, nil

	case successClearMsg:
		m.SuccessMsg = ""
		return m, nil
	}

	return m, nil
}

// ─── Key Press Router ────────────────────────────────────────────────────────

func (m Model) handleKeyPress(key string) (tea.Model, tea.Cmd) {
	m.ErrorMsg = ""

	if m.ConfirmActive {
		return m.handleConfirmKeys(key)
	}

	switch m.Screen {
	case ScreenDashboard:
		return m.handleDashboardKeys(key)
	case ScreenSearch:
		return m.handleSearchKeys(key)
	case ScreenSearchResults:
		return m.handleSearchResultsKeys(key)
	case ScreenRecent:
		return m.handleRecentKeys(key)
	case ScreenObservationDetail:
		return m.handleObservationDetailKeys(key)
	case ScreenTimeline:
		return m.handleTimelineKeys(key)
	case ScreenSessions:
		return m.handleSessionsKeys(key)
	case ScreenSessionDetail:
		return m.handleSessionDetailKeys(key)
	case ScreenProjects:
		return m.handleProjectsKeys(key)
	case ScreenProjectDetail:
		return m.handleProjectDetailKeys(key)
	case ScreenSetup:
		return m.handleSetupKeys(key)
	}
	return m, nil
}

func (m Model) handleConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y":
		return m.executeConfirmAction()
	case "n", "N", "esc":
		m.ConfirmActive = false
		m.ConfirmAction = ConfirmNone
		m.ConfirmMsg = ""
		m.ConfirmDetail = ""
		m.ConfirmTarget = ""
		return m, nil
	}
	return m, nil
}

func (m Model) executeConfirmAction() (tea.Model, tea.Cmd) {
	switch m.ConfirmAction {
	case ConfirmDeleteSession:
		return m, deleteSessionCmd(m.store, m.ConfirmTarget)
	case ConfirmClearProjectSessions:
		return m, clearProjectSessionsCmd(m.store, m.ConfirmTarget)
	case ConfirmDeleteProject:
		return m, deleteProjectCmd(m.store, m.ConfirmTarget)
	case ConfirmDeleteEmptySessions:
		return m, deleteEmptySessionsCmd(m.store, m.ConfirmTarget)
	}
	m.ConfirmActive = false
	return m, nil
}

// ─── Dashboard ───────────────────────────────────────────────────────────────

var dashboardMenuItems = []string{
	"Search memories",
	"Recent observations",
	"Browse sessions",
	"Browse projects",
	"Setup agent plugin",
	"Quit",
}

func (m Model) handleDashboardKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(dashboardMenuItems)-1 {
			m.Cursor++
		}
	case "enter", " ":
		return m.handleDashboardSelection()
	case "s", "/":
		m.PrevScreen = ScreenDashboard
		m.Screen = ScreenSearch
		m.Cursor = 0
		m.SearchInput.SetValue("")
		m.SearchInput.Focus()
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleDashboardSelection() (tea.Model, tea.Cmd) {
	switch m.Cursor {
	case 0: // Search
		m.PrevScreen = ScreenDashboard
		m.Screen = ScreenSearch
		m.Cursor = 0
		m.SearchInput.SetValue("")
		m.SearchInput.Focus()
		return m, nil
	case 1: // Recent observations
		m.PrevScreen = ScreenDashboard
		m.Screen = ScreenRecent
		m.Cursor = 0
		m.Scroll = 0
		return m, loadRecentObservations(m.store)
	case 2: // Sessions
		m.PrevScreen = ScreenDashboard
		m.Screen = ScreenSessions
		m.Cursor = 0
		m.Scroll = 0
		return m, loadRecentSessions(m.store)
	case 3: // Projects
		m.PrevScreen = ScreenDashboard
		m.Screen = ScreenProjects
		m.Cursor = 0
		m.Scroll = 0
		return m, loadProjects(m.store)
	case 4: // Setup
		m.PrevScreen = ScreenDashboard
		m.Screen = ScreenSetup
		m.Cursor = 0
		m.SetupAgents = setup.SupportedAgents()
		m.SetupResult = nil
		m.SetupError = ""
		m.SetupDone = false
		m.SetupInstalling = false
		m.SetupInstallingName = ""
		return m, nil
	case 5: // Quit
		return m, tea.Quit
	}
	return m, nil
}

// ─── Search Input ────────────────────────────────────────────────────────────

func (m Model) handleSearchInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := m.SearchInput.Value()
		if query != "" {
			m.SearchInput.Blur()
			return m, searchMemories(m.store, query)
		}
		return m, nil
	case "esc":
		m.SearchInput.Blur()
		m.Screen = ScreenDashboard
		m.Cursor = 0
		return m, nil
	}

	// Let the text input component handle everything else
	var cmd tea.Cmd
	m.SearchInput, cmd = m.SearchInput.Update(msg)
	return m, cmd
}

// ─── Filter Input ────────────────────────────────────────────────────────────

func (m Model) handleFilterInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Confirm filter — keep filtered results, blur input
		query := m.FilterInput.Value()
		m.FilterInput.Blur()
		m.FilterActive = false
		if query != "" {
			m.FilterQuery = query
		} else {
			m.FilterQuery = ""
		}
		return m, nil
	case "esc":
		// Cancel filter — clear query, reload full data
		m.FilterInput.Blur()
		m.FilterActive = false
		m.FilterQuery = ""
		m.FilterInput.SetValue("")
		m.Cursor = 0
		m.Scroll = 0
		m.ProjectDetailScroll = 0
		return m, m.reloadCurrentScreen()
	}

	// Let the text input component handle everything else (typing)
	var cmd tea.Cmd
	m.FilterInput, cmd = m.FilterInput.Update(msg)

	// Live filter — fire a filter command on every keystroke
	query := m.FilterInput.Value()
	if query != "" {
		return m, tea.Batch(cmd, m.fireFilterCmd(query))
	}
	// Empty input — reload full data
	return m, tea.Batch(cmd, m.reloadCurrentScreen())
}

// fireFilterCmd returns the appropriate filter command for the current screen.
func (m Model) fireFilterCmd(query string) tea.Cmd {
	switch m.Screen {
	case ScreenProjects:
		return filterProjects(m.store, query)
	case ScreenSessions:
		return filterSessions(m.store, query, "", 50)
	case ScreenProjectDetail:
		if m.SelectedProjectIdx < len(m.Projects) {
			project := m.Projects[m.SelectedProjectIdx].Name
			return filterSessions(m.store, query, project, 50)
		}
		return nil
	default:
		return nil
	}
}

// reloadCurrentScreen returns a command to reload the full (unfiltered) data for the current screen.
func (m Model) reloadCurrentScreen() tea.Cmd {
	switch m.Screen {
	case ScreenProjects:
		return loadProjects(m.store)
	case ScreenSessions:
		return loadRecentSessions(m.store)
	case ScreenProjectDetail:
		if m.SelectedProjectIdx < len(m.Projects) {
			project := m.Projects[m.SelectedProjectIdx].Name
			return loadProjectSessions(m.store, project, 50)
		}
		return nil
	default:
		return nil
	}
}

// resetFilter clears all filter state (call on screen transitions).
func (m *Model) resetFilter() {
	m.FilterActive = false
	m.FilterQuery = ""
	m.FilterInput.SetValue("")
	m.FilterInput.Blur()
}

// activateFilter activates the filter input for the current screen.
func (m *Model) activateFilter() {
	if m.ConfirmActive {
		return // Don't activate filter while confirmation dialog is active
	}
	m.FilterActive = true
	m.FilterQuery = ""
	m.FilterInput.SetValue("")
	m.FilterInput.Focus()
	m.Cursor = 0
	m.Scroll = 0
	m.ProjectDetailScroll = 0
}

func (m Model) handleSearchKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "q":
		m.Screen = ScreenDashboard
		m.Cursor = 0
		return m, nil
	case "i", "/":
		m.SearchInput.Focus()
		return m, nil
	}
	return m, nil
}

// ─── Search Results ──────────────────────────────────────────────────────────

func (m Model) handleSearchResultsKeys(key string) (tea.Model, tea.Cmd) {
	visibleItems := (m.Height - 10) / 2 // 2 lines per observation item
	if visibleItems < 3 {
		visibleItems = 3
	}

	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			// Scroll up if cursor goes above visible area
			if m.Cursor < m.Scroll {
				m.Scroll = m.Cursor
			}
		}
	case "down", "j":
		if m.Cursor < len(m.SearchResults)-1 {
			m.Cursor++
			// Scroll down if cursor goes below visible area
			if m.Cursor >= m.Scroll+visibleItems {
				m.Scroll = m.Cursor - visibleItems + 1
			}
		}
	case "enter":
		if len(m.SearchResults) > 0 && m.Cursor < len(m.SearchResults) {
			obsID := m.SearchResults[m.Cursor].ID
			m.PrevScreen = ScreenSearchResults
			return m, loadObservationDetail(m.store, obsID)
		}
	case "t":
		// Timeline for selected result
		if len(m.SearchResults) > 0 && m.Cursor < len(m.SearchResults) {
			obsID := m.SearchResults[m.Cursor].ID
			m.PrevScreen = ScreenSearchResults
			return m, loadTimeline(m.store, obsID)
		}
	case "/", "s":
		m.PrevScreen = ScreenSearchResults
		m.Screen = ScreenSearch
		m.SearchInput.Focus()
		return m, nil
	case "esc", "q":
		m.PrevScreen = ScreenDashboard
		m.Screen = ScreenSearch
		m.Cursor = 0
		m.Scroll = 0
		m.SearchInput.Focus()
		return m, nil
	}
	return m, nil
}

// ─── Recent Observations ─────────────────────────────────────────────────────

func (m Model) handleRecentKeys(key string) (tea.Model, tea.Cmd) {
	visibleItems := (m.Height - 8) / 2 // 2 lines per observation item
	if visibleItems < 3 {
		visibleItems = 3
	}

	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			if m.Cursor < m.Scroll {
				m.Scroll = m.Cursor
			}
		}
	case "down", "j":
		if m.Cursor < len(m.RecentObservations)-1 {
			m.Cursor++
			if m.Cursor >= m.Scroll+visibleItems {
				m.Scroll = m.Cursor - visibleItems + 1
			}
		}
	case "enter":
		if len(m.RecentObservations) > 0 && m.Cursor < len(m.RecentObservations) {
			obsID := m.RecentObservations[m.Cursor].ID
			m.PrevScreen = ScreenRecent
			return m, loadObservationDetail(m.store, obsID)
		}
	case "t":
		if len(m.RecentObservations) > 0 && m.Cursor < len(m.RecentObservations) {
			obsID := m.RecentObservations[m.Cursor].ID
			m.PrevScreen = ScreenRecent
			return m, loadTimeline(m.store, obsID)
		}
	case "esc", "q":
		m.Screen = ScreenDashboard
		m.Cursor = 0
		m.Scroll = 0
		return m, loadStats(m.store)
	}
	return m, nil
}

// ─── Observation Detail ──────────────────────────────────────────────────────

func (m Model) handleObservationDetailKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.DetailScroll > 0 {
			m.DetailScroll--
		}
	case "down", "j":
		m.DetailScroll++
	case "t":
		// View timeline for this observation
		if m.SelectedObservation != nil {
			return m, loadTimeline(m.store, m.SelectedObservation.ID)
		}
	case "esc", "q":
		m.Screen = m.PrevScreen
		m.Cursor = 0
		m.DetailScroll = 0
		return m, m.refreshScreen(m.PrevScreen)
	}
	return m, nil
}

// ─── Timeline ────────────────────────────────────────────────────────────────

func (m Model) handleTimelineKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.Scroll > 0 {
			m.Scroll--
		}
	case "down", "j":
		m.Scroll++
	case "esc", "q":
		m.Screen = m.PrevScreen
		m.Cursor = 0
		m.Scroll = 0
		return m, m.refreshScreen(m.PrevScreen)
	}
	return m, nil
}

// ─── Sessions ────────────────────────────────────────────────────────────────

func (m Model) handleSessionsKeys(key string) (tea.Model, tea.Cmd) {
	visibleItems := m.Height - 8
	if visibleItems < 5 {
		visibleItems = 5
	}

	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			if m.Cursor < m.Scroll {
				m.Scroll = m.Cursor
			}
		}
	case "down", "j":
		if m.Cursor < len(m.Sessions)-1 {
			m.Cursor++
			if m.Cursor >= m.Scroll+visibleItems {
				m.Scroll = m.Cursor - visibleItems + 1
			}
		}
	case "enter":
		if len(m.Sessions) > 0 && m.Cursor < len(m.Sessions) {
			m.SelectedSessionIdx = m.Cursor
			m.PrevScreen = ScreenSessions
			m.resetFilter()
			sessionID := m.Sessions[m.Cursor].ID
			return m, loadSessionObservations(m.store, sessionID)
		}
	case "/":
		m.activateFilter()
		return m, nil
	case "d":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		if len(m.Sessions) > 0 && m.Cursor < len(m.Sessions) {
			sess := m.Sessions[m.Cursor]
			sessID := sess.ID
			if len(sessID) > 8 {
				sessID = sessID[:8]
			}
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmDeleteSession
			m.ConfirmMsg = "Delete this session?"
			m.ConfirmDetail = fmt.Sprintf("Session: %s (%d observations)", sessID, sess.ObservationCount)
			m.ConfirmTarget = sess.ID
			return m, nil
		}
	case "e":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		stats, err := m.store.GetEmptySessionsStats("")
		if err != nil {
			m.ErrorMsg = fmt.Sprintf("Failed to count empty sessions: %v", err)
			return m, nil
		}
		if stats == nil {
			m.SuccessMsg = "No empty sessions (empty = no summary, observations, or prompts)"
			return m, clearSuccessAfterDelay()
		}
		m.ConfirmActive = true
		m.ConfirmAction = ConfirmDeleteEmptySessions
		m.ConfirmMsg = "Clear all empty sessions?"
		m.ConfirmDetail = buildEmptySessionsDetail(stats, "")
		m.ConfirmTarget = "" // all projects
		return m, nil
	case "esc":
		if m.FilterQuery != "" {
			m.resetFilter()
			m.Cursor = 0
			m.Scroll = 0
			return m, loadRecentSessions(m.store)
		}
		m.Screen = ScreenDashboard
		m.Cursor = 0
		m.Scroll = 0
		m.resetFilter()
		return m, loadStats(m.store)
	case "q":
		m.Screen = ScreenDashboard
		m.Cursor = 0
		m.Scroll = 0
		m.resetFilter()
		return m, loadStats(m.store)
	}
	return m, nil
}

// ─── Session Detail ──────────────────────────────────────────────────────────

func (m Model) handleSessionDetailKeys(key string) (tea.Model, tea.Cmd) {
	visibleItems := (m.Height - 12) / 2
	if visibleItems < 3 {
		visibleItems = 3
	}

	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			if m.Cursor < m.SessionDetailScroll {
				m.SessionDetailScroll = m.Cursor
			}
		}
	case "down", "j":
		if m.Cursor < len(m.SessionObservations)-1 {
			m.Cursor++
			if m.Cursor >= m.SessionDetailScroll+visibleItems {
				m.SessionDetailScroll = m.Cursor - visibleItems + 1
			}
		}
	case "enter":
		if len(m.SessionObservations) > 0 && m.Cursor < len(m.SessionObservations) {
			obsID := m.SessionObservations[m.Cursor].ID
			m.PrevScreen = ScreenSessionDetail
			return m, loadObservationDetail(m.store, obsID)
		}
	case "t":
		if len(m.SessionObservations) > 0 && m.Cursor < len(m.SessionObservations) {
			obsID := m.SessionObservations[m.Cursor].ID
			m.PrevScreen = ScreenSessionDetail
			return m, loadTimeline(m.store, obsID)
		}
	case "d":
		if m.SelectedSessionIdx < len(m.Sessions) {
			sess := m.Sessions[m.SelectedSessionIdx]
			sessID := sess.ID
			if len(sessID) > 8 {
				sessID = sessID[:8]
			}
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmDeleteSession
			m.ConfirmMsg = "Delete this session?"
			m.ConfirmDetail = fmt.Sprintf("Session: %s (%d observations)", sessID, sess.ObservationCount)
			m.ConfirmTarget = sess.ID
			return m, nil
		}
	case "c":
		if m.SelectedSessionIdx < len(m.Sessions) {
			sess := m.Sessions[m.SelectedSessionIdx]
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmClearProjectSessions
			m.ConfirmMsg = "Clear ALL sessions for this project?"
			m.ConfirmDetail = fmt.Sprintf("Project: %s", sess.Project)
			m.ConfirmTarget = sess.Project
			return m, nil
		}
	case "D":
		if m.SelectedSessionIdx < len(m.Sessions) {
			sess := m.Sessions[m.SelectedSessionIdx]
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmDeleteProject
			m.ConfirmMsg = "DELETE ENTIRE PROJECT?"
			m.ConfirmDetail = fmt.Sprintf("Project: %s — This cannot be undone!", sess.Project)
			m.ConfirmTarget = sess.Project
			return m, nil
		}
	case "esc", "q":
		m.Screen = ScreenSessions
		m.Cursor = m.SelectedSessionIdx
		m.SessionDetailScroll = 0
		return m, loadRecentSessions(m.store)
	}
	return m, nil
}

// ─── Projects ────────────────────────────────────────────────────────────────

func (m Model) handleProjectsKeys(key string) (tea.Model, tea.Cmd) {
	visibleItems := m.Height - 8
	if visibleItems < 5 {
		visibleItems = 5
	}

	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			if m.Cursor < m.Scroll {
				m.Scroll = m.Cursor
			}
		}
	case "down", "j":
		if m.Cursor < len(m.Projects)-1 {
			m.Cursor++
			if m.Cursor >= m.Scroll+visibleItems {
				m.Scroll = m.Cursor - visibleItems + 1
			}
		}
	case "enter":
		if len(m.Projects) > 0 && m.Cursor < len(m.Projects) {
			m.SelectedProjectIdx = m.Cursor
			m.PrevScreen = ScreenProjects
			m.resetFilter()
			project := m.Projects[m.Cursor].Name
			return m, loadProjectSessions(m.store, project, 50)
		}
	case "/":
		m.activateFilter()
		return m, nil
	case "d":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		if len(m.Projects) > 0 && m.Cursor < len(m.Projects) {
			proj := m.Projects[m.Cursor]
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmDeleteProject
			m.ConfirmMsg = "DELETE ENTIRE PROJECT?"
			m.ConfirmDetail = fmt.Sprintf("Project: %s (%d sessions, %d observations) — This cannot be undone!", proj.Name, proj.SessionCount, proj.ObservationCount)
			m.ConfirmTarget = proj.Name
			return m, nil
		}
	case "c":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		if len(m.Projects) > 0 && m.Cursor < len(m.Projects) {
			proj := m.Projects[m.Cursor]
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmClearProjectSessions
			m.ConfirmMsg = "Clear ALL sessions for this project?"
			m.ConfirmDetail = fmt.Sprintf("Project: %s (%d sessions)", proj.Name, proj.SessionCount)
			m.ConfirmTarget = proj.Name
			return m, nil
		}
	case "e":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		if len(m.Projects) > 0 && m.Cursor < len(m.Projects) {
			proj := m.Projects[m.Cursor]
			stats, err := m.store.GetEmptySessionsStats(proj.Name)
			if err != nil {
				m.ErrorMsg = fmt.Sprintf("Failed to count empty sessions: %v", err)
				return m, nil
			}
			if stats == nil {
				m.SuccessMsg = fmt.Sprintf("No empty sessions in %s (empty = no summary, observations, or prompts)", proj.Name)
				return m, clearSuccessAfterDelay()
			}
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmDeleteEmptySessions
			m.ConfirmMsg = fmt.Sprintf("Clear empty sessions for %s?", proj.Name)
			m.ConfirmDetail = buildEmptySessionsDetail(stats, proj.Name)
			m.ConfirmTarget = proj.Name
			return m, nil
		}
	case "esc":
		if m.FilterQuery != "" {
			m.resetFilter()
			m.Cursor = 0
			m.Scroll = 0
			return m, loadProjects(m.store)
		}
		m.Screen = ScreenDashboard
		m.Cursor = 0
		m.Scroll = 0
		m.resetFilter()
		return m, loadStats(m.store)
	case "q":
		m.Screen = ScreenDashboard
		m.Cursor = 0
		m.Scroll = 0
		m.resetFilter()
		return m, loadStats(m.store)
	}
	return m, nil
}

// ─── Project Detail ──────────────────────────────────────────────────────────

func (m Model) handleProjectDetailKeys(key string) (tea.Model, tea.Cmd) {
	visibleItems := m.Height - 10
	if visibleItems < 5 {
		visibleItems = 5
	}

	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			if m.Cursor < m.ProjectDetailScroll {
				m.ProjectDetailScroll = m.Cursor
			}
		}
	case "down", "j":
		if m.Cursor < len(m.ProjectSessions)-1 {
			m.Cursor++
			if m.Cursor >= m.ProjectDetailScroll+visibleItems {
				m.ProjectDetailScroll = m.Cursor - visibleItems + 1
			}
		}
	case "enter":
		if len(m.ProjectSessions) > 0 && m.Cursor < len(m.ProjectSessions) {
			m.SelectedSessionIdx = m.Cursor
			m.Sessions = m.ProjectSessions // Make session detail work
			m.PrevScreen = ScreenProjectDetail
			m.resetFilter()
			sessionID := m.ProjectSessions[m.Cursor].ID
			return m, loadSessionObservations(m.store, sessionID)
		}
	case "/":
		m.activateFilter()
		return m, nil
	case "c":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		if m.SelectedProjectIdx < len(m.Projects) {
			proj := m.Projects[m.SelectedProjectIdx]
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmClearProjectSessions
			m.ConfirmMsg = "Clear ALL sessions for this project?"
			m.ConfirmDetail = fmt.Sprintf("Project: %s (%d sessions)", proj.Name, proj.SessionCount)
			m.ConfirmTarget = proj.Name
			return m, nil
		}
	case "D":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		if m.SelectedProjectIdx < len(m.Projects) {
			proj := m.Projects[m.SelectedProjectIdx]
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmDeleteProject
			m.ConfirmMsg = "DELETE ENTIRE PROJECT?"
			m.ConfirmDetail = fmt.Sprintf("Project: %s — This cannot be undone!", proj.Name)
			m.ConfirmTarget = proj.Name
			return m, nil
		}
	case "e":
		if m.FilterActive || m.FilterQuery != "" {
			return m, nil // Block destructive keys when filter is active
		}
		if m.SelectedProjectIdx < len(m.Projects) {
			proj := m.Projects[m.SelectedProjectIdx]
			stats, err := m.store.GetEmptySessionsStats(proj.Name)
			if err != nil {
				m.ErrorMsg = fmt.Sprintf("Failed to count empty sessions: %v", err)
				return m, nil
			}
			if stats == nil {
				m.SuccessMsg = fmt.Sprintf("No empty sessions in %s (empty = no summary, observations, or prompts)", proj.Name)
				return m, clearSuccessAfterDelay()
			}
			m.ConfirmActive = true
			m.ConfirmAction = ConfirmDeleteEmptySessions
			m.ConfirmMsg = fmt.Sprintf("Clear empty sessions for %s?", proj.Name)
			m.ConfirmDetail = buildEmptySessionsDetail(stats, proj.Name)
			m.ConfirmTarget = proj.Name
			return m, nil
		}
	case "esc":
		if m.FilterQuery != "" {
			m.resetFilter()
			m.Cursor = 0
			m.ProjectDetailScroll = 0
			if m.SelectedProjectIdx < len(m.Projects) {
				project := m.Projects[m.SelectedProjectIdx].Name
				return m, loadProjectSessions(m.store, project, 50)
			}
			return m, nil
		}
		m.Screen = ScreenProjects
		m.Cursor = m.SelectedProjectIdx
		m.ProjectDetailScroll = 0
		m.resetFilter()
		return m, loadProjects(m.store)
	case "q":
		m.Screen = ScreenProjects
		m.Cursor = m.SelectedProjectIdx
		m.ProjectDetailScroll = 0
		m.resetFilter()
		return m, loadProjects(m.store)
	}
	return m, nil
}

// ─── Setup ───────────────────────────────────────────────────────────────────

func (m Model) handleSetupKeys(key string) (tea.Model, tea.Cmd) {
	// While installing, block all keys
	if m.SetupInstalling {
		return m, nil
	}

	// Allowlist prompt: y/n
	if m.SetupAllowlistPrompt {
		switch key {
		case "y", "Y":
			m.SetupAllowlistPrompt = false
			m.SetupDone = true
			if err := addClaudeCodeAllowlistFn(); err != nil {
				m.SetupAllowlistError = err.Error()
			} else {
				m.SetupAllowlistApplied = true
			}
			return m, nil
		case "n", "N", "esc":
			m.SetupAllowlistPrompt = false
			m.SetupDone = true
			return m, nil
		}
		return m, nil
	}

	// After install completed, any key goes back
	if m.SetupDone {
		switch key {
		case "esc", "q", "enter":
			m.Screen = ScreenDashboard
			m.Cursor = 0
			m.SetupDone = false
			m.SetupResult = nil
			m.SetupError = ""
			m.SetupAllowlistApplied = false
			m.SetupAllowlistError = ""
			return m, loadStats(m.store)
		}
		return m, nil
	}

	switch key {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.SetupAgents)-1 {
			m.Cursor++
		}
	case "enter":
		if len(m.SetupAgents) > 0 && m.Cursor < len(m.SetupAgents) {
			agent := m.SetupAgents[m.Cursor]
			m.SetupInstalling = true
			m.SetupInstallingName = agent.Name
			return m, tea.Batch(m.SetupSpinner.Tick, installAgent(agent.Name))
		}
	case "esc", "q":
		m.Screen = ScreenDashboard
		m.Cursor = 0
		return m, loadStats(m.store)
	}
	return m, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// refreshScreen returns the appropriate data-loading Cmd for a given screen.
// Used when navigating back so lists show fresh data from the DB.
func (m Model) refreshScreen(screen Screen) tea.Cmd {
	switch screen {
	case ScreenDashboard:
		return loadStats(m.store)
	case ScreenRecent:
		return loadRecentObservations(m.store)
	case ScreenSessions:
		return loadRecentSessions(m.store)
	case ScreenProjects:
		return loadProjects(m.store)
	default:
		return nil
	}
}

// buildEmptySessionsDetail builds the multi-line ConfirmDetail string from EmptySessionsStats.
// For all-projects scope (project == ""), includes project breakdown.
// For single-project scope, omits the breakdown line.
func buildEmptySessionsDetail(stats *store.EmptySessionsStats, project string) string {
	var lines []string

	// Line 1: "{empty} of {total} sessions ({pct}%) will be removed"
	pct := 0
	if stats.TotalCount > 0 {
		pct = stats.EmptyCount * 100 / stats.TotalCount
	}
	lines = append(lines, fmt.Sprintf("%d of %d sessions (%d%%) will be removed",
		stats.EmptyCount, stats.TotalCount, pct))

	// Line 2 (all-projects only): top 3 projects with "+N more"
	if project == "" && len(stats.ProjectBreakdown) > 0 {
		var parts []string
		shown := 3
		if len(stats.ProjectBreakdown) < shown {
			shown = len(stats.ProjectBreakdown)
		}
		for _, pc := range stats.ProjectBreakdown[:shown] {
			parts = append(parts, fmt.Sprintf("%s: %d", pc.Project, pc.Count))
		}
		line := strings.Join(parts, " · ")
		if remaining := len(stats.ProjectBreakdown) - shown; remaining > 0 {
			line += fmt.Sprintf(" · +%d more", remaining)
		}
		lines = append(lines, line)
	}

	// Line 3: date range
	if stats.OldestDate != "" {
		if stats.OldestDate == stats.NewestDate {
			lines = append(lines, fmt.Sprintf("From %s", stats.OldestDate))
		} else {
			lines = append(lines, fmt.Sprintf("From %s to %s", stats.OldestDate, stats.NewestDate))
		}
	}

	// Line 4: definition of "empty"
	lines = append(lines, "(Empty = no summary, no observations, no prompts)")

	return strings.Join(lines, "\n")
}
