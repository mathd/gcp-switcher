package internal

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mathd/gcp-switcher/cmd/gcp"
	"github.com/mathd/gcp-switcher/types"
)

// Update handles state updates based on incoming messages
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle state-specific updates first (for proper list navigation)
	switch m.State {
	case StateAccounts:
		m.AccountList, cmd = m.AccountList.Update(msg)
		cmds = append(cmds, cmd)
	case StateProjects:
		m.ProjectList, cmd = m.ProjectList.Update(msg)
		cmds = append(cmds, cmd)
	case StateManualProject:
		m.ProjectInput, cmd = m.ProjectInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case types.FallbackTimerMsg:
		return m.handleFallbackTimer(msg)

	case tea.KeyMsg:
		// Handle general key messages after state-specific ones
		newModel, newCmd := m.handleKeyMsg(msg)
		if newCmd != nil {
			cmds = append(cmds, newCmd)
		}
		m = newModel.(AppModel)

	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)

	case types.ErrMsg:
		return m.handleErrMsg(msg)

	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)

	case types.GcloudCheckMsg:
		return m.handleGcloudCheckMsg(msg)

	case types.ActiveAccountMsg:
		return m.handleActiveAccountMsg(msg)

	case types.ActiveProjectMsg:
		return m.handleActiveProjectMsg(msg)

	case types.AccountListMsg:
		return m.handleAccountListMsg(msg)

	case types.ProjectListMsg:
		return m.handleProjectListMsg(msg)

	case types.OperationResultMsg:
		return m.handleOperationResultMsg(msg)
	}

	if !m.Loaded {
		cmds = append(cmds, m.Spinner.Tick)
	}

	return m, tea.Batch(cmds...)
}

// handleFallbackTimer handles fallback timer messages
func (m AppModel) handleFallbackTimer(msg types.FallbackTimerMsg) (tea.Model, tea.Cmd) {
	if m.State == StateLoading {
		m.State = StateMain
		m.Loaded = true
	}
	return m, nil
}

// handleKeyMsg handles keyboard input messages
func (m AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m.handleQuitKey()
	case "up", "k":
		// Only handle up/down in main menu, let other states handle their own navigation
		if m.State == StateMain {
			return m.handleUpKey()
		}
	case "down", "j":
		// Only handle up/down in main menu, let other states handle their own navigation
		if m.State == StateMain {
			return m.handleDownKey()
		}
	case "1", "a":
		return m.handleMenuChoice(0)
	case "2", "p":
		return m.handleMenuChoice(1)
	case "3", "l":
		return m.handleMenuChoice(2)
	case "4", "m":
		return m.handleMenuChoice(3)
	case "enter":
		return handleEnterKey(m)
	}
	return m, nil
}

// handleQuitKey handles quit key presses
func (m AppModel) handleQuitKey() (tea.Model, tea.Cmd) {
	if m.State == StateMain || m.State == StateLoading || m.State == StateError {
		return m, tea.Quit
	}
	m.State = StateMain
	return m, nil
}

// handleUpKey handles up arrow key presses
func (m AppModel) handleUpKey() (tea.Model, tea.Cmd) {
	if m.State == StateMain {
		m.MainMenuChoice = (m.MainMenuChoice - 1 + 4) % 4
	}
	return m, nil
}

// handleDownKey handles down arrow key presses
func (m AppModel) handleDownKey() (tea.Model, tea.Cmd) {
	if m.State == StateMain {
		m.MainMenuChoice = (m.MainMenuChoice + 1) % 4
	}
	return m, nil
}

// handleMenuChoice handles menu selection shortcuts
func (m AppModel) handleMenuChoice(choice int) (tea.Model, tea.Cmd) {
	if m.State != StateMain {
		return m, nil
	}

	m.MainMenuChoice = choice
	switch choice {
	case 0: // Accounts
		if len(m.Accounts) == 0 {
			m.State = StateLoadingAccounts
			return m, gcp.GetAllAccounts()
		}
		m.State = StateAccounts
	case 1: // Projects
		if len(m.Projects) == 0 {
			m.State = StateLoadingProjects
			return m, gcp.GetSimpleProjects()
		}
		m.State = StateProjects
	case 2: // New Login
		m.PreviousState = StateNewLogin
		m.State = StateConfirming
		m.ConfirmationText = "Would you like to login to a new GCP account?"
	case 3: // Manual Project
		m.State = StateManualProject
		m.ProjectInput.Focus()
	}
	return m, nil
}

// handleWindowSizeMsg handles window resize messages
func (m AppModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.Width = msg.Width
	m.Height = msg.Height
	m.AccountList.SetSize(msg.Width-4, listHeight)
	m.ProjectList.SetSize(msg.Width-4, listHeight)
	return m, nil
}

// handleErrMsg handles error messages
func (m AppModel) handleErrMsg(msg types.ErrMsg) (tea.Model, tea.Cmd) {
	m.CommandErrors = append(m.CommandErrors, msg.Err.Error())
	m.CommandsComplete++
	if m.CommandsComplete >= m.TotalCommands {
		m.State = StateMain
		m.Loaded = true
	}
	return m, nil
}

// handleGcloudCheckMsg handles gcloud availability check messages
func (m AppModel) handleGcloudCheckMsg(msg types.GcloudCheckMsg) (tea.Model, tea.Cmd) {
	if !msg.Available {
		m.State = StateError
		m.Err = fmt.Errorf("Google Cloud SDK (gcloud) is not installed or not in PATH\nPlease install it from: https://cloud.google.com/sdk/docs/install")
		m.Loaded = true
	} else {
		m.CommandsComplete++
		CheckCompletion(&m)
	}
	return m, nil
}

// handleActiveAccountMsg handles active account update messages
func (m AppModel) handleActiveAccountMsg(msg types.ActiveAccountMsg) (tea.Model, tea.Cmd) {
	m.ActiveAccount = msg.Account
	m.CommandsComplete++
	CheckCompletion(&m)
	return m, nil
}

// handleActiveProjectMsg handles active project update messages
func (m AppModel) handleActiveProjectMsg(msg types.ActiveProjectMsg) (tea.Model, tea.Cmd) {
	m.ActiveProject = msg.Project
	m.CommandsComplete++
	CheckCompletion(&m)
	return m, nil
}

// handleAccountListMsg handles account list update messages
func (m AppModel) handleAccountListMsg(msg types.AccountListMsg) (tea.Model, tea.Cmd) {
	m.Accounts = msg.Accounts
	accountItems := make([]list.Item, len(m.Accounts))
	for i, account := range m.Accounts {
		accountItems[i] = types.NewItem(
			account.Account,
			"",
			account.Status == "ACTIVE",
			account.Account,
		)
	}
	m.AccountList.SetItems(accountItems)

	if m.State == StateLoadingAccounts {
		m.State = StateAccounts
	}

	m.CommandsComplete++
	CheckCompletion(&m)
	return m, nil
}

// handleProjectListMsg handles project list update messages
func (m AppModel) handleProjectListMsg(msg types.ProjectListMsg) (tea.Model, tea.Cmd) {
	m.Projects = msg.Projects
	projectItems := make([]list.Item, len(m.Projects))
	for i, project := range m.Projects {
		projectItems[i] = types.NewItem(
			project.ProjectID,
			project.Name,
			project.ProjectID == m.ActiveProject,
			project.ProjectID,
		)
	}
	m.ProjectList.SetItems(projectItems)

	if m.ActiveProject == "" {
		m.State = StateProjects
	} else if m.State == StateLoadingProjects {
		m.State = StateProjects
	}
	m.CommandsComplete++
	if m.ActiveProject != "" {
		CheckCompletion(&m)
	}
	return m, nil
}

// handleOperationResultMsg handles operation result messages
func (m AppModel) handleOperationResultMsg(msg types.OperationResultMsg) (tea.Model, tea.Cmd) {
	if msg.Success {
		if msg.Err != nil && msg.Err.Error() == "ACCOUNT_SWITCHED" {
			m.ActiveProject = ""
			m.Projects = nil
			m.ProjectList.SetItems([]list.Item{})
			m.State = StateLoadingProjects

			return m, tea.Batch(
				gcp.GetActiveAccount(),
				gcp.GetAllAccounts(),
				gcp.GetSimpleProjects(),
			)
		} else {
			cmds := []tea.Cmd{
				gcp.GetActiveAccount(),
				gcp.GetActiveProject(),
				gcp.GetAllAccounts(),
				gcp.GetSimpleProjects(),
			}
			m.State = StateMain
			return m, tea.Batch(cmds...)
		}
	} else {
		m.State = StateError
		m.Err = msg.Err
	}
	return m, nil
}

// handleEnterKey handles the Enter key press based on the current state
func handleEnterKey(m AppModel) (tea.Model, tea.Cmd) {
	switch m.State {
	case StateMain:
		switch m.MainMenuChoice {
		case 0: // View/Switch Accounts
			if len(m.Accounts) == 0 {
				m.State = StateLoadingAccounts
				return m, gcp.GetAllAccounts()
			}
			m.State = StateAccounts
		case 1: // View/Switch Projects
			if len(m.Projects) == 0 {
				m.State = StateLoadingProjects
				return m, gcp.GetSimpleProjects()
			}
			m.State = StateProjects
		case 2: // Login to a New Account
			m.PreviousState = StateNewLogin
			m.State = StateConfirming
			m.ConfirmationText = "Would you like to login to a new GCP account?"
		case 3: // Enter Project ID Manually
			m.State = StateManualProject
			m.ProjectInput.Focus()
		}

	case StateAccounts:
		if len(m.Accounts) > 0 {
			selectedItem := m.AccountList.SelectedItem().(types.Item)
			m.PreviousState = StateAccounts
			m.SelectedItemID = selectedItem.ID()
			m.State = StateConfirming
			m.ConfirmationText = fmt.Sprintf("Switch to account %s?", selectedItem.ID())
		}

	case StateProjects:
		if len(m.Projects) > 0 {
			selectedItem := m.ProjectList.SelectedItem().(types.Item)
			if selectedItem.ID() != m.ActiveProject {
				m.PreviousState = StateProjects
				m.SelectedItemID = selectedItem.ID()
				m.State = StateConfirming
				m.ConfirmationText = fmt.Sprintf("Switch to project %s?", selectedItem.ID())
			}
		}

	case StateManualProject:
		projectID := m.ProjectInput.Value()
		if projectID != "" && projectID != m.ActiveProject {
			m.PreviousState = StateManualProject
			m.SelectedItemID = projectID
			m.State = StateConfirming
			m.ConfirmationText = fmt.Sprintf("Switch to project %s?", projectID)
		}

	case StateConfirming:
		if m.ConfirmationChoice == 0 { // Yes
			m.State = StateProcessing

			switch m.PreviousState {
			case StateAccounts:
				return m, gcp.SwitchAccount(m.SelectedItemID)
			case StateProjects:
				return m, gcp.SwitchProject(m.SelectedItemID)
			case StateManualProject:
				return m, gcp.SwitchProject(m.SelectedItemID)
			case StateNewLogin:
				return m, gcp.LoginNewAccount()
			}
		} else {
			if m.PreviousState == StateNewLogin || m.PreviousState == StateManualProject {
				m.State = StateMain
			} else {
				m.State = m.PreviousState
			}
		}
	}

	return m, nil
}
