package internal

import (
	"fmt"
)

// View renders the current application state
func (m AppModel) View() string {
	var s string

	currentState := m.StateMachine.GetState()
	stateContext := m.StateMachine.GetContext()

	switch currentState {
	case StateLoading:
		var loadingText string
		switch stateContext.LoadingContext {
		case LoadingInitial:
			loadingText = "Loading GCP configuration..."
			s = fmt.Sprintf(
				"\n\n   %s %s\n\n",
				m.Components.Spinner.View(),
				loadingText,
			)
			s += m.UI.Styles.Info.Render(fmt.Sprintf("  Commands completed: %d/%d\n", m.Operations.CommandsComplete, m.Operations.TotalCommands))
		case LoadingAccounts:
			loadingText = "Loading Accounts..."
		case LoadingProjects:
			loadingText = "Loading Projects..."
		}

		if stateContext.LoadingContext != LoadingInitial {
			s = fmt.Sprintf(
				"\n\n   %s %s\n\n",
				m.Components.Spinner.View(),
				loadingText,
			)
		}

	case StateError:
		s = m.UI.Styles.Title.Render("Error") + "\n\n"
		s += m.UI.Styles.Error.Render(m.UI.Err.Error()) + "\n\n"
		s += m.UI.Styles.Info.Render("Press q to quit")

	case StateMain:
		s = m.UI.Styles.Title.Render("GCP Account Manager") + "\n\n"

		// Account and project info
		accountInfo := fmt.Sprintf("Active Account: %s", m.UI.Styles.Highlight.Render(m.Data.ActiveAccount))
		projectInfo := fmt.Sprintf("Active Project: %s", m.UI.Styles.Highlight.Render(m.Data.ActiveProject))
		s += accountInfo + "\n" + projectInfo + "\n\n"

		// Menu options
		s += m.UI.Styles.Subtitle.Render("What would you like to do?") + "\n\n"
		menuItems := []string{
			" View/Switch Accounts ",
			" View/Switch Projects ",
			" Login to a New Account ",
			" Enter Project ID Manually ",
		}
		for i, item := range menuItems {
			buttonStyle := m.UI.Styles.BlurredButton
			if i == m.UI.MainMenuChoice {
				buttonStyle = m.UI.Styles.FocusedButton
			}
			s += fmt.Sprintf("%d. %s\n", i+1, buttonStyle.Render(item))
		}
		s += "\n" + m.UI.Styles.Info.Render("Press q to quit, ↑/↓ to navigate, Enter to select")

	case StateAccounts:
		s = m.Components.AccountList.View()
		s += "\n" + m.UI.Styles.Info.Render("Press Enter to select, q to go back")

	case StateProjects:
		s = m.Components.ProjectList.View()
		s += "\n" + m.UI.Styles.Info.Render("Press Enter to select, q to go back")

	case StateManualProject:
		s = m.UI.Styles.Title.Render("Enter Project ID") + "\n\n"
		s += "Please enter the GCP project ID you want to switch to:\n\n"
		s += m.Components.ProjectInput.View() + "\n\n"
		s += m.UI.Styles.Info.Render("Press Enter to confirm, q to go back")

	case StateConfirming:
		s = m.UI.Styles.Title.Render("Confirmation") + "\n\n"
		s += m.StateMachine.GetConfirmationText() + "\n\n"

		yesStyle := m.UI.Styles.BlurredButton
		noStyle := m.UI.Styles.BlurredButton

		if m.UI.ConfirmationChoice == 0 {
			yesStyle = m.UI.Styles.FocusedButton
		} else {
			noStyle = m.UI.Styles.FocusedButton
		}

		s += yesStyle.Render(" Yes ") + "   " + noStyle.Render(" No ")
		s += "\n\n" + m.UI.Styles.Info.Render("(Use arrow keys to select, Enter to confirm)")

	case StateProcessing:
		s = fmt.Sprintf(
			"\n\n   %s Processing, please wait...\n\n",
			m.Components.Spinner.View(),
		)
	}

	return m.UI.Styles.App.Render(s)
}
