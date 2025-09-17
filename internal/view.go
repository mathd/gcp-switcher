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
				m.Spinner.View(),
				loadingText,
			)
			s += m.Styles.Info.Render(fmt.Sprintf("  Commands completed: %d/%d\n", m.CommandsComplete, m.TotalCommands))
		case LoadingAccounts:
			loadingText = "Loading Accounts..."
		case LoadingProjects:
			loadingText = "Loading Projects..."
		}

		if stateContext.LoadingContext != LoadingInitial {
			s = fmt.Sprintf(
				"\n\n   %s %s\n\n",
				m.Spinner.View(),
				loadingText,
			)
		}

	case StateError:
		s = m.Styles.Title.Render("Error") + "\n\n"
		s += m.Styles.Error.Render(m.Err.Error()) + "\n\n"
		s += m.Styles.Info.Render("Press q to quit")

	case StateMain:
		s = m.Styles.Title.Render("GCP Account Manager") + "\n\n"

		// Account and project info
		accountInfo := fmt.Sprintf("Active Account: %s", m.Styles.Highlight.Render(m.ActiveAccount))
		projectInfo := fmt.Sprintf("Active Project: %s", m.Styles.Highlight.Render(m.ActiveProject))
		s += accountInfo + "\n" + projectInfo + "\n\n"

		// Menu options
		s += m.Styles.Subtitle.Render("What would you like to do?") + "\n\n"
		menuItems := []string{
			" View/Switch Accounts ",
			" View/Switch Projects ",
			" Login to a New Account ",
			" Enter Project ID Manually ",
		}
		for i, item := range menuItems {
			buttonStyle := m.Styles.BlurredButton
			if i == m.MainMenuChoice {
				buttonStyle = m.Styles.FocusedButton
			}
			s += fmt.Sprintf("%d. %s\n", i+1, buttonStyle.Render(item))
		}
		s += "\n" + m.Styles.Info.Render("Press q to quit, ↑/↓ to navigate, Enter to select")

	case StateAccounts:
		s = m.AccountList.View()
		s += "\n" + m.Styles.Info.Render("Press Enter to select, q to go back")

	case StateProjects:
		s = m.ProjectList.View()
		s += "\n" + m.Styles.Info.Render("Press Enter to select, q to go back")

	case StateManualProject:
		s = m.Styles.Title.Render("Enter Project ID") + "\n\n"
		s += "Please enter the GCP project ID you want to switch to:\n\n"
		s += m.ProjectInput.View() + "\n\n"
		s += m.Styles.Info.Render("Press Enter to confirm, q to go back")

	case StateConfirming:
		s = m.Styles.Title.Render("Confirmation") + "\n\n"
		s += m.StateMachine.GetConfirmationText() + "\n\n"

		yesStyle := m.Styles.BlurredButton
		noStyle := m.Styles.BlurredButton

		if m.ConfirmationChoice == 0 {
			yesStyle = m.Styles.FocusedButton
		} else {
			noStyle = m.Styles.FocusedButton
		}

		s += yesStyle.Render(" Yes ") + "   " + noStyle.Render(" No ")
		s += "\n\n" + m.Styles.Info.Render("(Use arrow keys to select, Enter to confirm)")

	case StateProcessing:
		s = fmt.Sprintf(
			"\n\n   %s Processing, please wait...\n\n",
			m.Spinner.View(),
		)
	}

	return m.Styles.App.Render(s)
}
