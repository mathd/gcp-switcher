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

	// Handle state-specific UI updates first (for proper component navigation)
	currentState := m.StateMachine.GetState()
	switch currentState {
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
		if currentState == StateLoading {
			m.StateMachine.Fire(TriggerDataLoaded)
		}

	case tea.KeyMsg:
		newModel, newCmd := m.handleKeyMsg(msg)
		if newCmd != nil {
			cmds = append(cmds, newCmd)
		}
		m = newModel.(AppModel)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.AccountList.SetSize(msg.Width-4, listHeight)
		m.ProjectList.SetSize(msg.Width-4, listHeight)

	case types.ErrMsg:
		m.CommandErrors = append(m.CommandErrors, msg.Err.Error())
		m.CommandsComplete++
		if m.CommandsComplete >= m.TotalCommands {
			m.StateMachine.Fire(TriggerDataLoaded)
		}

	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)

	case types.GcloudCheckMsg:
		if !msg.Available {
			m.Err = fmt.Errorf("Google Cloud SDK (gcloud) is not installed or not in PATH\nPlease install it from: https://cloud.google.com/sdk/docs/install")
			m.StateMachine.Fire(TriggerError, m.Err)
		} else {
			m.CommandsComplete++
			CheckCompletion(&m)
		}

	case types.ActiveAccountMsg:
		m.ActiveAccount = msg.Account
		m.CommandsComplete++
		CheckCompletion(&m)

	case types.ActiveProjectMsg:
		m.ActiveProject = msg.Project
		m.CommandsComplete++
		CheckCompletion(&m)

	case types.AccountListMsg:
		m.Accounts = msg.Accounts
		m.StateMachine.SetHasAccounts(len(m.Accounts) > 0)
		m.updateAccountList()
		if currentState == StateLoading && m.StateMachine.GetContext().LoadingContext == LoadingAccounts {
			m.StateMachine.Fire(TriggerDataLoaded)
		}
		m.CommandsComplete++
		CheckCompletion(&m)

	case types.ProjectListMsg:
		m.Projects = msg.Projects
		m.StateMachine.SetHasProjects(len(m.Projects) > 0)
		m.updateProjectList()
		if currentState == StateLoading && m.StateMachine.GetContext().LoadingContext == LoadingProjects {
			m.StateMachine.Fire(TriggerDataLoaded)
		}
		m.CommandsComplete++
		if m.ActiveProject != "" {
			CheckCompletion(&m)
		}

	case types.OperationResultMsg:
		if msg.Success {
			if msg.Err != nil && msg.Err.Error() == "ACCOUNT_SWITCHED" {
				m.ActiveProject = ""
				m.Projects = nil
				m.ProjectList.SetItems([]list.Item{})
				m.StateMachine.Fire(TriggerLoadProjects, LoadingProjects)
				cmd = m.StateMachine.GetLoadCommand()
				cmds = append(cmds, cmd)
			} else {
				m.StateMachine.Fire(TriggerOperationComplete)
				cmds = append(cmds, tea.Batch(
					gcp.GetActiveAccount(),
					gcp.GetActiveProject(),
					gcp.GetAllAccounts(),
					gcp.GetSimpleProjects(),
				))
			}
		} else {
			m.Err = msg.Err
			m.StateMachine.Fire(TriggerOperationFailed, msg.Err)
		}
	}

	if !m.Loaded {
		cmds = append(cmds, m.Spinner.Tick)
	}

	return m, tea.Batch(cmds...)
}

// updateAccountList updates the account list items
func (m *AppModel) updateAccountList() {
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
}

// updateProjectList updates the project list items
func (m *AppModel) updateProjectList() {
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
}

// handleFallbackTimer handles fallback timer messages
func (m AppModel) handleFallbackTimer(msg types.FallbackTimerMsg) (tea.Model, tea.Cmd) {
	if m.StateMachine.GetState() == StateLoading {
		m.StateMachine.Fire(TriggerDataLoaded)
		m.Loaded = true
	}
	return m, nil
}

// handleKeyMsg handles keyboard input messages
func (m AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	currentState := m.StateMachine.GetState()

	switch msg.String() {
	case "ctrl+c", "q":
		if currentState == StateMain || currentState == StateLoading || currentState == StateError {
			return m, tea.Quit
		}
		m.StateMachine.Fire(TriggerGoBack)

	case "up", "k":
		if currentState == StateMain {
			m.MainMenuChoice = (m.MainMenuChoice - 1 + 4) % 4
		} else if currentState == StateConfirming {
			m.ConfirmationChoice = 0
		}

	case "down", "j":
		if currentState == StateMain {
			m.MainMenuChoice = (m.MainMenuChoice + 1) % 4
		} else if currentState == StateConfirming {
			m.ConfirmationChoice = 1
		}

	case "left", "right":
		if currentState == StateConfirming {
			m.ConfirmationChoice = 1 - m.ConfirmationChoice
		}

	case "1", "a":
		if currentState == StateMain {
			return m.handleMenuChoice(0)
		}
	case "2", "p":
		if currentState == StateMain {
			return m.handleMenuChoice(1)
		}
	case "3", "l":
		if currentState == StateMain {
			return m.handleMenuChoice(2)
		}
	case "4", "m":
		if currentState == StateMain {
			return m.handleMenuChoice(3)
		}

	case "enter":
		return m.handleEnterKey()
	}
	return m, nil
}

// handleEnterKey handles the Enter key press based on the current state
func (m AppModel) handleEnterKey() (tea.Model, tea.Cmd) {
	currentState := m.StateMachine.GetState()

	switch currentState {
	case StateMain:
		return m.handleMenuChoice(m.MainMenuChoice)

	case StateAccounts:
		if len(m.Accounts) > 0 {
			selectedItem := m.AccountList.SelectedItem().(types.Item)
			m.StateMachine.SetSelectedID(selectedItem.ID())
			m.StateMachine.Fire(TriggerAccountSelected, fmt.Sprintf("Switch to account %s?", selectedItem.ID()))
		}

	case StateProjects:
		if len(m.Projects) > 0 {
			selectedItem := m.ProjectList.SelectedItem().(types.Item)
			if selectedItem.ID() != m.ActiveProject {
				m.StateMachine.SetSelectedID(selectedItem.ID())
				m.StateMachine.Fire(TriggerProjectSelected, fmt.Sprintf("Switch to project %s?", selectedItem.ID()))
			}
		}

	case StateManualProject:
		projectID := m.ProjectInput.Value()
		if projectID != "" && projectID != m.ActiveProject {
			m.StateMachine.SetSelectedID(projectID)
			m.StateMachine.Fire(TriggerManualProjectEntry, fmt.Sprintf("Switch to project %s?", projectID))
		}

	case StateConfirming:
		if m.ConfirmationChoice == 0 { // Yes
			m.StateMachine.Fire(TriggerConfirmYes)
			return m, m.StateMachine.GetActionCommand()
		} else {
			m.StateMachine.Fire(TriggerConfirmNo)
		}
	}

	return m, nil
}

// handleMenuChoice handles menu selection shortcuts
func (m AppModel) handleMenuChoice(choice int) (tea.Model, tea.Cmd) {
	currentState := m.StateMachine.GetState()
	if currentState != StateMain {
		return m, nil
	}

	m.MainMenuChoice = choice
	m.StateMachine.SetMenuChoice(choice)

	var cmd tea.Cmd
	switch choice {
	case 0: // Accounts
		if m.StateMachine.CanFire(TriggerLoadAccounts) {
			m.StateMachine.Fire(TriggerLoadAccounts, LoadingAccounts)
			cmd = m.StateMachine.GetLoadCommand()
		} else {
			m.StateMachine.Fire(TriggerMenuChoice)
		}
	case 1: // Projects
		if m.StateMachine.CanFire(TriggerLoadProjects) {
			m.StateMachine.Fire(TriggerLoadProjects, LoadingProjects)
			cmd = m.StateMachine.GetLoadCommand()
		} else {
			m.StateMachine.Fire(TriggerMenuChoice)
		}
	case 2: // New Login
		m.StateMachine.Fire(TriggerMenuChoice, "Would you like to login to a new GCP account?")
	case 3: // Manual Project
		m.StateMachine.Fire(TriggerMenuChoice)
		m.ProjectInput.Focus()
	}
	return m, cmd
}
