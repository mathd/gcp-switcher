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

	switch msg := msg.(type) {
	case types.FallbackTimerMsg:
		if m.State == StateLoading {
			m.State = StateMain
			m.Loaded = true
			return m, nil
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.State == StateMain || m.State == StateLoading || m.State == StateError {
				return m, tea.Quit
			}
			m.State = StateMain
			return m, nil

		case "up", "k":
			if m.State == StateMain {
				m.MainMenuChoice = (m.MainMenuChoice - 1 + 4) % 4
			}

		case "down", "j":
			if m.State == StateMain {
				m.MainMenuChoice = (m.MainMenuChoice + 1) % 4
			}

		case "1", "a":
			if m.State == StateMain {
				m.MainMenuChoice = 0
				if len(m.Accounts) == 0 {
					m.State = StateLoadingAccounts
					return m, gcp.GetAllAccounts()
				}
				m.State = StateAccounts
			}

		case "2", "p":
			if m.State == StateMain {
				m.MainMenuChoice = 1
				if len(m.Projects) == 0 {
					m.State = StateLoadingProjects
					return m, gcp.GetSimpleProjects()
				}
				m.State = StateProjects
			}

		case "3", "l":
			if m.State == StateMain {
				m.MainMenuChoice = 2
				m.PreviousState = StateNewLogin
				m.State = StateConfirming
				m.ConfirmationText = "Would you like to login to a new GCP account?"
			}

		case "4", "m":
			if m.State == StateMain {
				m.MainMenuChoice = 3
				m.State = StateManualProject
				m.ProjectInput.Focus()
			}

		case "enter":
			return handleEnterKey(m)
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.AccountList.SetSize(msg.Width-4, listHeight)
		m.ProjectList.SetSize(msg.Width-4, listHeight)

	case types.ErrMsg:
		m.CommandErrors = append(m.CommandErrors, msg.Err.Error())
		m.CommandsComplete++
		if m.CommandsComplete >= m.TotalCommands {
			m.State = StateMain
			m.Loaded = true
		}
		return m, nil

	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		cmds = append(cmds, cmd)

	case types.GcloudCheckMsg:
		if !msg.Available {
			m.State = StateError
			m.Err = fmt.Errorf("Google Cloud SDK (gcloud) is not installed or not in PATH\nPlease install it from: https://cloud.google.com/sdk/docs/install")
			m.Loaded = true
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

	case types.ProjectListMsg:
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

	case types.OperationResultMsg:
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
				cmds = append(cmds,
					gcp.GetActiveAccount(),
					gcp.GetActiveProject(),
					gcp.GetAllAccounts(),
					gcp.GetSimpleProjects(),
				)
				m.State = StateMain
			}
		} else {
			m.State = StateError
			m.Err = msg.Err
		}
	}

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

	if !m.Loaded {
		cmds = append(cmds, m.Spinner.Tick)
	}

	return m, tea.Batch(cmds...)
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
