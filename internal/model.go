package internal

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mathd/gcp-switcher/cmd/gcp"
	"github.com/mathd/gcp-switcher/types"
	"github.com/mathd/gcp-switcher/ui"
)

const (
	listHeight = 20
)

// UI States
const (
	StateLoading         = "loading"
	StateLoadingAccounts = "loading_accounts"
	StateLoadingProjects = "loading_projects"
	StateError           = "error"
	StateMain            = "main"
	StateAccounts        = "accounts"
	StateProjects        = "projects"
	StateManualProject   = "manual_project"
	StateConfirming      = "confirming"
	StateProcessing      = "processing"
	StateNewLogin        = "new_login"
)

// AppModel represents the application state
type AppModel struct {
	State              string
	Accounts           []types.Account
	Projects           []types.Project
	ActiveAccount      string
	ActiveProject      string
	AccountList        list.Model
	ProjectList        list.Model
	Spinner            spinner.Model
	Width              int
	Height             int
	Loaded             bool
	Err                error
	SearchInput        textinput.Model
	ProjectInput       textinput.Model
	ConfirmationChoice int
	ConfirmationText   string
	CommandsComplete   int
	TotalCommands      int
	CommandErrors      []string
	PreviousState      string
	SelectedItemID     string
	MainMenuChoice     int
	Styles             ui.Styles
}

// Init initializes the application model
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		gcp.CheckGcloud,
		gcp.GetActiveAccount(),
		gcp.GetActiveProject(),
		gcp.GetAllAccounts(),
		gcp.GetSimpleProjects(),
		createFallbackTimer(10),
	)
}

// InitialModel creates and returns the initial application model
func InitialModel(styles ui.Styles) AppModel {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Title

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
	accountList.Styles.Title = styles.Title
	accountList.Styles.PaginationStyle = styles.Subtitle
	accountList.Styles.HelpStyle = styles.Info

	// Initialize project list
	projectList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	projectList.Title = "GCP Projects"
	projectList.SetShowTitle(true)
	projectList.SetShowStatusBar(true)
	projectList.SetFilteringEnabled(true)
	projectList.Styles.Title = styles.Title
	projectList.Styles.PaginationStyle = styles.Subtitle
	projectList.Styles.HelpStyle = styles.Info

	return AppModel{
		State:              StateLoading,
		Accounts:           []types.Account{},
		Projects:           []types.Project{},
		Spinner:            s,
		SearchInput:        ti,
		ProjectInput:       pi,
		AccountList:        accountList,
		ProjectList:        projectList,
		ConfirmationChoice: 0,
		CommandsComplete:   0,
		TotalCommands:      4,
		CommandErrors:      []string{},
		PreviousState:      "",
		SelectedItemID:     "",
		Styles:             styles,
	}
}

// CheckCompletion checks if all commands are complete and updates the state
func CheckCompletion(m *AppModel) {
	if m.CommandsComplete >= m.TotalCommands {
		m.Loaded = true
		m.State = StateMain
	}
}

// createFallbackTimer creates a timer to prevent infinite loading
func createFallbackTimer(seconds int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Duration(seconds) * time.Second)
		return types.FallbackTimerMsg{TimeoutSeconds: seconds}
	}
}
