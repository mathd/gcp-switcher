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

// Loading contexts for StateLoading
type LoadingContext int

const (
	LoadingInitial LoadingContext = iota
	LoadingAccounts
	LoadingProjects
)

// AppData holds application data state
type AppData struct {
	Accounts      []types.Account
	Projects      []types.Project
	ActiveAccount string
	ActiveProject string
}

// UIComponents holds all UI component state
type UIComponents struct {
	AccountList  list.Model
	ProjectList  list.Model
	Spinner      spinner.Model
	SearchInput  textinput.Model
	ProjectInput textinput.Model
}

// UIState holds UI-specific state
type UIState struct {
	Width              int
	Height             int
	Loaded             bool
	Err                error
	ConfirmationChoice int
	MainMenuChoice     int
	Styles             ui.Styles
	NeedProjectSelection bool // Flag to trigger project selection after account switch
}

// OperationState holds operation tracking state
type OperationState struct {
	CommandsComplete int
	TotalCommands    int
	CommandErrors    []string
}

// AppModel represents the application state
type AppModel struct {
	// State machine for formal state management
	StateMachine *AppStateMachine

	// Grouped state components
	Data       AppData
	Components UIComponents
	UI         UIState
	Operations OperationState
}

// Init initializes the application model
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.Components.Spinner.Tick,
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

	// Initialize state machine
	stateMachine := NewAppStateMachine()

	return AppModel{
		StateMachine: stateMachine,
		Data: AppData{
			Accounts:      []types.Account{},
			Projects:      []types.Project{},
			ActiveAccount: "",
			ActiveProject: "",
		},
		Components: UIComponents{
			Spinner:      s,
			SearchInput:  ti,
			ProjectInput: pi,
			AccountList:  accountList,
			ProjectList:  projectList,
		},
		UI: UIState{
			ConfirmationChoice: 0,
			Styles:             styles,
		},
		Operations: OperationState{
			CommandsComplete: 0,
			TotalCommands:    4,
			CommandErrors:    []string{},
		},
	}
}

// CheckCompletion checks if all commands are complete and updates the state
func CheckCompletion(m *AppModel) {
	if m.Operations.CommandsComplete >= m.Operations.TotalCommands {
		m.UI.Loaded = true
		m.StateMachine.Fire(TriggerDataLoaded)
	}
}

// createFallbackTimer creates a timer to prevent infinite loading
func createFallbackTimer(seconds int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Duration(seconds) * time.Second)
		return types.FallbackTimerMsg{TimeoutSeconds: seconds}
	}
}
