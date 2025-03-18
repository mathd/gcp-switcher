package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Configuration constants
const (
	listHeight     = 20
	defaultWidth   = 80
	commandTimeout = 5 * time.Second
	logFilePath    = "gcp-switcher.log"
	longTimeout    = 30 * time.Second
)

// UI States
const (
	stateLoading         = "loading"
	stateLoadingAccounts = "loading_accounts"
	stateLoadingProjects = "loading_projects"
	stateError           = "error"
	stateMain            = "main"
	stateAccounts        = "accounts"
	stateProjects        = "projects"
	stateManualProject   = "manual_project"
	stateConfirming      = "confirming"
	stateProcessing      = "processing"
	stateNewLogin        = "new_login"
)

// Styles defines the UI styling configuration
type Styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	subtitle      lipgloss.Style
	info          lipgloss.Style
	success       lipgloss.Style
	error         lipgloss.Style
	highlight     lipgloss.Style
	focusedButton lipgloss.Style
	blurredButton lipgloss.Style
	activeItem    lipgloss.Style
}

var (
	debugMode bool
	logger    *log.Logger
	styles    = NewStyles()
)

// NewStyles initializes and returns the UI styles
func NewStyles() Styles {
	return Styles{
		app: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true).
			MarginLeft(2),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("105")),
		info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("247")),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("84")),
		error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")),
		highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("159")).
			Bold(true),
		focusedButton: lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Background(lipgloss.Color("99")).
			Padding(0, 2).
			Bold(true),
		blurredButton: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2),
		activeItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("159")).
			Bold(true),
	}
}

// Init logger
func initLogger() {
	if !debugMode {
		logger = log.New(io.Discard, "", 0)
		return
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}

	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	logger.Println("=== GCP Switcher Started ===")
}

// Account represents a GCP account
type Account struct {
	Account string `json:"account"`
	Status  string `json:"status"`
}

// Project represents a GCP project
type Project struct {
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
}

// Item represents an item in the list
type Item struct {
	title       string
	description string
	isActive    bool
	id          string
}

func (i Item) Title() string {
	if i.isActive {
		return styles.activeItem.Render(i.title + " (ACTIVE)")
	}
	return i.title
}

func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title + i.description }

// AppModel represents the application state
type AppModel struct {
	state              string
	accounts           []Account
	projects           []Project
	activeAccount      string
	activeProject      string
	accountList        list.Model
	projectList        list.Model
	spinner            spinner.Model
	width              int
	height             int
	loaded             bool
	err                error
	searchInput        textinput.Model
	projectInput       textinput.Model
	confirmationChoice int
	confirmationText   string
	commandsComplete   int
	totalCommands      int
	commandErrors      []string
	previousState      string // Track previous state for confirmation
	selectedItemID     string // Track selected item ID for confirmation
	mainMenuChoice     int    // Track main menu selection
}

func initialModel() AppModel {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // Match the title color

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search projects..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

	// Initialize project input
	pi := textinput.New()
	pi.Placeholder = "Enter project ID..."
	pi.Focus()
	pi.CharLimit = 50
	pi.Width = 30

	// Initialize account list
	accountList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	accountList.Title = "GCP Accounts"
	accountList.SetShowTitle(true)
	accountList.SetShowStatusBar(true)
	accountList.SetFilteringEnabled(true)
	accountList.Styles.Title = styles.title
	accountList.Styles.PaginationStyle = styles.subtitle
	accountList.Styles.HelpStyle = styles.info

	// Initialize project list
	projectList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	projectList.Title = "GCP Projects"
	projectList.SetShowTitle(true)
	projectList.SetShowStatusBar(true)
	projectList.SetFilteringEnabled(true)
	projectList.Styles.Title = styles.title
	projectList.Styles.PaginationStyle = styles.subtitle
	projectList.Styles.HelpStyle = styles.info

	return AppModel{
		state:              stateLoading,
		accounts:           []Account{},
		projects:           []Project{},
		spinner:            s,
		searchInput:        ti,
		projectInput:       pi,
		accountList:        accountList,
		projectList:        projectList,
		confirmationChoice: 0,
		commandsComplete:   0,
		totalCommands:      4, // We'll load accounts and projects at startup
		commandErrors:      []string{},
		previousState:      "",
		selectedItemID:     "",
	}
}

func (m AppModel) Init() tea.Cmd {
	// Load all data at startup but with a reasonable fallback timer
	return tea.Batch(
		m.spinner.Tick,
		checkGcloudCmd,
		getActiveAccountCmd(),
		getActiveProjectCmd(),
		getAllAccountsCmd(),     // Load accounts at startup
		getSimpleProjectsCmd(),  // Load projects at startup
		createFallbackTimer(10), // 10 second fallback timer
	)
}

// Fallback timer to prevent infinite loading
func createFallbackTimer(seconds int) tea.Cmd {
	return func() tea.Msg {
		logger.Printf("Starting fallback timer for %d seconds", seconds)
		time.Sleep(time.Duration(seconds) * time.Second)
		logger.Println("Fallback timer triggered")
		return fallbackTimerMsg{timeoutSeconds: seconds}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Only log non-spinner messages to reduce noise
	if _, isSpinnerMsg := msg.(spinner.TickMsg); !isSpinnerMsg {
		logger.Printf("Update called with message type: %T", msg)
	}

	switch msg := msg.(type) {
	case fallbackTimerMsg:
		// Force transition if still loading
		if m.state == "loading" {
			logger.Printf("Fallback timer (%ds) forced transition to main", msg.timeoutSeconds)
			// Go to main menu even if not all commands are complete
			m.state = "main"
			m.loaded = true
			return m, nil
		}

	case tea.KeyMsg:
		logger.Printf("Key pressed: %s", msg.String())
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == "main" || m.state == "loading" || m.state == "error" {
				logger.Println("Quitting application")
				return m, tea.Quit
			}
			// Go back to main menu
			m.state = "main"
			return m, nil

		case "up", "k":
			if m.state == "main" {
				m.mainMenuChoice = (m.mainMenuChoice - 1 + 4) % 4
				return m, nil
			}

		case "down", "j":
			if m.state == "main" {
				m.mainMenuChoice = (m.mainMenuChoice + 1) % 4
				return m, nil
			}

		case "1", "a":
			if m.state == "main" {
				m.mainMenuChoice = 0
				// If we don't have accounts yet, load them now
				if len(m.accounts) == 0 {
					m.state = "loading_accounts"
					return m, getAllAccountsCmd()
				} else {
					m.state = "accounts"
				}
				return m, nil
			}

		case "2", "p":
			if m.state == "main" {
				m.mainMenuChoice = 1
				// If we don't have projects yet, load them now
				if len(m.projects) == 0 {
					m.state = "loading_projects"
					return m, getSimpleProjectsCmd()
				} else {
					m.state = "projects"
				}
				return m, nil
			}

		case "3", "l":
			if m.state == "main" {
				m.mainMenuChoice = 2
				m.previousState = "new_login" // Set previous state
				m.state = "confirming"
				m.confirmationText = "Would you like to login to a new GCP account?"
				return m, nil
			}

		case "4", "m":
			if m.state == "main" {
				m.mainMenuChoice = 3
				m.state = "manual_project"
				m.projectInput.Focus()
				return m, nil
			}

		case "enter":
			switch m.state {
			case "main":
				switch m.mainMenuChoice {
				case 0: // View/Switch Accounts
					if len(m.accounts) == 0 {
						m.state = "loading_accounts"
						return m, getAllAccountsCmd()
					} else {
						m.state = "accounts"
					}
				case 1: // View/Switch Projects
					if len(m.projects) == 0 {
						m.state = "loading_projects"
						return m, getSimpleProjectsCmd()
					} else {
						m.state = "projects"
					}
				case 2: // Login to a New Account
					m.previousState = "new_login"
					m.state = "confirming"
					m.confirmationText = "Would you like to login to a new GCP account?"
				case 3: // Enter Project ID Manually
					m.state = "manual_project"
					m.projectInput.Focus()
				}
			case "accounts":
				if len(m.accounts) > 0 {
					selectedItem := m.accountList.SelectedItem().(Item)
					if selectedItem.id != m.activeAccount {
						m.previousState = "accounts"       // Save previous state
						m.selectedItemID = selectedItem.id // Save selected item ID
						m.state = "confirming"
						m.confirmationText = fmt.Sprintf("Switch to account %s?", selectedItem.id)
						return m, nil
					}
				}
			case "projects":
				if len(m.projects) > 0 {
					selectedItem := m.projectList.SelectedItem().(Item)
					if selectedItem.id != m.activeProject {
						m.previousState = "projects"       // Save previous state
						m.selectedItemID = selectedItem.id // Save selected item ID
						m.state = "confirming"
						m.confirmationText = fmt.Sprintf("Switch to project %s?", selectedItem.id)
						return m, nil
					}
				}
			case "manual_project":
				projectID := m.projectInput.Value()
				if projectID != "" && projectID != m.activeProject {
					m.previousState = "manual_project" // Save previous state
					m.selectedItemID = projectID       // Save project ID
					m.state = "confirming"
					m.confirmationText = fmt.Sprintf("Switch to project %s?", projectID)
					return m, nil
				}
			case "confirming":
				if m.confirmationChoice == 0 { // Yes
					m.state = "processing"
					logger.Printf("Confirmation YES with previousState=%s, selectedItemID=%s", m.previousState, m.selectedItemID)

					if m.previousState == "accounts" {
						return m, switchAccountCmd(m.selectedItemID)
					} else if m.previousState == "projects" {
						return m, switchProjectCmd(m.selectedItemID)
					} else if m.previousState == "manual_project" {
						return m, switchProjectCmd(m.selectedItemID)
					} else if m.previousState == "new_login" {
						return m, loginNewAccountCmd()
					}
				} else {
					// Go back to previous state
					if m.previousState == "new_login" || m.previousState == "manual_project" {
						m.state = "main"
					} else {
						m.state = m.previousState
					}
				}
			}

		case "left", "h":
			if m.state == "confirming" {
				m.confirmationChoice = 0 // Yes
			}

		case "right":
			if m.state == "confirming" {
				m.confirmationChoice = 1 // No
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.accountList.SetSize(msg.Width-4, listHeight)
		m.projectList.SetSize(msg.Width-4, listHeight)

	case errMsg:
		logger.Printf("Error encountered: %v", msg.err)
		m.commandErrors = append(m.commandErrors, msg.err.Error())

		// Increment command counter to move on
		m.commandsComplete++

		if m.commandsComplete >= m.totalCommands {
			m.state = "main"
			m.loaded = true
		}
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case gcloudCheckMsg:
		logger.Printf("Received gcloudCheckMsg: %v", msg.available)
		if !msg.available {
			m.state = "error"
			m.err = fmt.Errorf("Google Cloud SDK (gcloud) is not installed or not in PATH\nPlease install it from: https://cloud.google.com/sdk/docs/install")
			m.loaded = true
		} else {
			m.commandsComplete++
			checkCompletion(&m)
		}

	case activeAccountMsg:
		logger.Printf("Received activeAccountMsg: %s", msg.account)
		m.activeAccount = msg.account
		m.commandsComplete++
		checkCompletion(&m)

	case activeProjectMsg:
		logger.Printf("Received activeProjectMsg: %s", msg.project)
		m.activeProject = msg.project
		m.commandsComplete++
		checkCompletion(&m)

	case accountListMsg:
		logger.Printf("Received accountListMsg with %d accounts", len(msg.accounts))
		m.accounts = msg.accounts
		accountItems := []list.Item{}
		for _, account := range m.accounts {
			accountItems = append(accountItems, Item{
				title:       account.Account,
				description: "",
				isActive:    account.Status == "ACTIVE",
				id:          account.Account,
			})
		}
		m.accountList.SetItems(accountItems)

		// If we're in the loading_accounts state, change to accounts
		if m.state == "loading_accounts" {
			m.state = "accounts"
		}

		m.commandsComplete++
		checkCompletion(&m)

	case projectListMsg:
		logger.Printf("Received projectListMsg with %d projects", len(msg.projects))
		m.projects = msg.projects
		projectItems := []list.Item{}
		for _, project := range m.projects {
			projectItems = append(projectItems, Item{
				title:       project.ProjectID,
				description: project.Name,
				isActive:    project.ProjectID == m.activeProject,
				id:          project.ProjectID,
			})
		}
		m.projectList.SetItems(projectItems)

		// If we're in the loading_projects state, change to projects
		if m.state == "loading_projects" {
			m.state = "projects"
		}

		m.commandsComplete++
		checkCompletion(&m)

	case operationResultMsg:
		logger.Printf("Operation result: success=%v", msg.success)
		if msg.success {
			// Check if we need to force project selection after account switch
			if msg.err != nil && msg.err.Error() == "SELECT_PROJECT" {
				// Load projects and transition to project selection
				if len(m.projects) == 0 {
					m.state = "loading_projects"
					return m, getSimpleProjectsCmd()
				} else {
					m.state = "projects"
					return m, nil
				}
			}

			// Refresh the data
			cmds = append(cmds,
				getActiveAccountCmd(),
				getActiveProjectCmd(),
				getAllAccountsCmd(),
				getSimpleProjectsCmd(),
			)
			m.state = "main"
		} else {
			m.state = "error"
			m.err = msg.err
		}
	}

	switch m.state {
	case "accounts":
		m.accountList, cmd = m.accountList.Update(msg)
		cmds = append(cmds, cmd)
	case "projects":
		m.projectList, cmd = m.projectList.Update(msg)
		cmds = append(cmds, cmd)
	case "manual_project":
		m.projectInput, cmd = m.projectInput.Update(msg)
		cmds = append(cmds, cmd)
	case "search":
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if !m.loaded {
		cmds = append(cmds, m.spinner.Tick)
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	var s string

	switch m.state {
	case "loading":
		s = fmt.Sprintf(
			"\n\n   %s Loading GCP configuration...\n\n",
			m.spinner.View(),
		)
		// Show loading progress
		s += styles.info.Render(fmt.Sprintf("  Commands completed: %d/%d\n", m.commandsComplete, m.totalCommands))
		return styles.app.Render(s)

	case "loading_accounts":
		s = fmt.Sprintf(
			"\n\n   %s Loading Accounts...\n\n",
			m.spinner.View(),
		)
		return styles.app.Render(s)

	case "loading_projects":
		s = fmt.Sprintf(
			"\n\n   %s Loading Projects...\n\n",
			m.spinner.View(),
		)
		return styles.app.Render(s)

	case "error":
		s = styles.title.Render("Error") + "\n\n"
		s += styles.error.Render(m.err.Error()) + "\n\n"
		s += styles.info.Render("Press q to quit")
		return styles.app.Render(s)

	case "main":
		// Header
		s = styles.title.Render("GCP Account Manager") + "\n\n"

		// Account and project info
		accountInfo := fmt.Sprintf("Active Account: %s", styles.highlight.Render(m.activeAccount))
		projectInfo := fmt.Sprintf("Active Project: %s", styles.highlight.Render(m.activeProject))
		s += accountInfo + "\n" + projectInfo + "\n\n"

		// Menu options
		s += styles.subtitle.Render("What would you like to do?") + "\n\n"
		menuItems := []string{
			" View/Switch Accounts ",
			" View/Switch Projects ",
			" Login to a New Account ",
			" Enter Project ID Manually ",
		}
		for i, item := range menuItems {
			buttonStyle := styles.blurredButton
			if i == m.mainMenuChoice {
				buttonStyle = styles.focusedButton
			}
			s += fmt.Sprintf("%d. %s\n", i+1, buttonStyle.Render(item))
		}
		s += "\n" + styles.info.Render("Press q to quit, ↑/↓ to navigate, Enter to select")

	case "accounts":
		s = m.accountList.View()
		s += "\n" + styles.info.Render("Press Enter to select, q to go back")

	case "projects":
		s = m.projectList.View()
		s += "\n" + styles.info.Render("Press Enter to select, q to go back")

	case "manual_project":
		s = styles.title.Render("Enter Project ID") + "\n\n"
		s += "Please enter the GCP project ID you want to switch to:\n\n"
		s += m.projectInput.View() + "\n\n"
		s += styles.info.Render("Press Enter to confirm, q to go back")

	case "confirming":
		s = styles.title.Render("Confirmation") + "\n\n"
		s += m.confirmationText + "\n\n"

		yesStyle := styles.blurredButton
		noStyle := styles.blurredButton

		if m.confirmationChoice == 0 {
			yesStyle = styles.focusedButton
		} else {
			noStyle = styles.focusedButton
		}

		s += yesStyle.Render(" Yes ") + "   " + noStyle.Render(" No ")
		s += "\n\n" + styles.info.Render("(Use arrow keys to select, Enter to confirm)")

	case "processing":
		s = fmt.Sprintf(
			"\n\n   %s Processing, please wait...\n\n",
			m.spinner.View(),
		)

	case "new_login":
		s = styles.title.Render("Login to a New Account") + "\n\n"
		s += m.confirmationText + "\n\n"

		yesStyle := styles.blurredButton
		noStyle := styles.blurredButton

		if m.confirmationChoice == 0 {
			yesStyle = styles.focusedButton
		} else {
			noStyle = styles.focusedButton
		}

		s += yesStyle.Render(" Yes ") + "   " + noStyle.Render(" No ")
		s += "\n\n" + styles.info.Render("(Use arrow keys to select, Enter to confirm)")
	}

	return styles.app.Render(s)
}

// Custom message types
type spinnerMsg tea.Msg
type errMsg struct{ err error }
type gcloudCheckMsg struct{ available bool }
type activeAccountMsg struct{ account string }
type activeProjectMsg struct{ project string }
type accountListMsg struct{ accounts []Account }
type projectListMsg struct{ projects []Project }
type operationResultMsg struct {
	success bool
	err     error
}
type fallbackTimerMsg struct{ timeoutSeconds int }

func (e errMsg) Error() string { return e.err.Error() }

func checkCompletion(m *AppModel) {
	logger.Printf("Checking completion: %d/%d commands complete", m.commandsComplete, m.totalCommands)
	if m.commandsComplete >= m.totalCommands {
		logger.Println("All commands complete! Transitioning to main state")
		m.loaded = true
		m.state = "main"
	}
}

// Commands
func checkGcloudCmd() tea.Msg {
	logger.Println("Checking if gcloud is installed")
	_, err := exec.LookPath("gcloud")
	result := gcloudCheckMsg{available: err == nil}
	logger.Printf("gcloud available: %v", result.available)
	return result
}

func getActiveAccountCmd() tea.Cmd {
	return func() tea.Msg {
		logger.Println("Getting active account")

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				logger.Println("Command timed out: gcloud auth list")
				return errMsg{fmt.Errorf("command timed out: gcloud auth list")}
			}
			logger.Printf("Error getting active account: %v", err)
			return errMsg{err}
		}

		account := strings.TrimSpace(string(output))
		logger.Printf("Active account: %s", account)
		return activeAccountMsg{account: account}
	}
}

func getActiveProjectCmd() tea.Cmd {
	return func() tea.Msg {
		logger.Println("Getting active project")

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "project")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				logger.Println("Command timed out: gcloud config get-value project")
				return errMsg{fmt.Errorf("command timed out: gcloud config get-value project")}
			}
			logger.Printf("Error getting active project: %v", err)
			return errMsg{err}
		}

		project := strings.TrimSpace(string(output))
		logger.Printf("Active project: %s", project)
		return activeProjectMsg{project: project}
	}
}

func getAllAccountsCmd() tea.Cmd {
	return func() tea.Msg {
		logger.Println("Getting all accounts")

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "auth", "list", "--format=json")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				logger.Println("Command timed out: gcloud auth list --format=json")
				return errMsg{fmt.Errorf("command timed out: gcloud auth list --format=json")}
			}
			logger.Printf("Error getting accounts: %v", err)
			return errMsg{err}
		}

		var accounts []Account
		if err := json.Unmarshal(output, &accounts); err != nil {
			logger.Printf("Error unmarshaling accounts: %v", err)
			return errMsg{err}
		}
		logger.Printf("Found %d accounts", len(accounts))
		return accountListMsg{accounts: accounts}
	}
}

// Use a more reliable method to get projects
func getSimpleProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		logger.Println("Getting projects (simple method)")

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "projects", "list", "--format=json")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				logger.Println("Command timed out: gcloud projects list")
				// Return empty project list instead of error
				return projectListMsg{projects: []Project{}}
			}
			logger.Printf("Error getting projects: %v", err)
			// Return empty list on error
			return projectListMsg{projects: []Project{}}
		}

		var projects []Project
		if err := json.Unmarshal(output, &projects); err != nil {
			logger.Printf("Error unmarshaling projects: %v", err)
			return projectListMsg{projects: []Project{}}
		}

		logger.Printf("Found %d projects", len(projects))
		return projectListMsg{projects: projects}
	}
}

func switchAccountCmd(account string) tea.Cmd {
	return func() tea.Msg {
		logger.Printf("Switching to account: %s", account)

		// Create context with timeout (longer for interactive operations)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Login to account - simplified version
		cmd := exec.CommandContext(ctx, "gcloud", "config", "set", "account", account)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			logger.Printf("Error switching account: %v", err)
			return operationResultMsg{success: false, err: err}
		}

		logger.Println("Account switch successful")
		// Return special success message that indicates we need to select a project
		return operationResultMsg{success: true, err: fmt.Errorf("SELECT_PROJECT")}
	}
}

func loginNewAccountCmd() tea.Cmd {
	return func() tea.Msg {
		logger.Println("Logging in to new account")

		// Create context with timeout (longer for interactive operations)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Login to new account
		loginCmd := exec.CommandContext(ctx, "gcloud", "auth", "login")
		loginCmd.Stdin = os.Stdin
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		err := loginCmd.Run()
		if err != nil {
			logger.Printf("Error logging in: %v", err)
			return operationResultMsg{success: false, err: err}
		}

		// Set up application default credentials
		adcCmd := exec.CommandContext(ctx, "gcloud", "auth", "application-default", "login")
		adcCmd.Stdin = os.Stdin
		adcCmd.Stdout = os.Stdout
		adcCmd.Stderr = os.Stderr
		err = adcCmd.Run()
		if err != nil {
			logger.Printf("Error setting up application default credentials: %v", err)
			return operationResultMsg{success: false, err: err}
		}

		logger.Println("Login successful")
		return operationResultMsg{success: true}
	}
}

func switchProjectCmd(projectID string) tea.Cmd {
	return func() tea.Msg {
		logger.Printf("Switching to project: %s", projectID)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "config", "set", "project", projectID)
		err := cmd.Run()
		if err != nil {
			logger.Printf("Error switching project: %v", err)
			return operationResultMsg{success: false, err: err}
		}

		logger.Println("Project switch successful")
		return operationResultMsg{success: true}
	}
}

func main() {
	// Parse command line flags
	flag.BoolVar(&debugMode, "debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	initLogger()

	logger.Println("Starting GCP Switcher application")
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	// Start the program
	if _, err := p.Run(); err != nil {
		logger.Printf("Error running program: %v", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
