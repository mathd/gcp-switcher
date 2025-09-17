package internal

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mathd/gcp-switcher/cmd/gcp"
	"github.com/qmuntal/stateless"
)

// AppState represents the application states
type AppState int

const (
	StateLoading AppState = iota
	StateError
	StateMain
	StateAccounts
	StateProjects
	StateManualProject
	StateConfirming
	StateProcessing
)

// AppTrigger represents the state transition triggers
type AppTrigger int

const (
	TriggerDataLoaded AppTrigger = iota
	TriggerError
	TriggerMenuChoice
	TriggerLoadAccounts
	TriggerLoadProjects
	TriggerAccountSelected
	TriggerProjectSelected
	TriggerManualProjectEntry
	TriggerConfirmYes
	TriggerConfirmNo
	TriggerOperationComplete
	TriggerOperationFailed
	TriggerGoBack
)

// StateMachineContext holds data for state transitions
type StateMachineContext struct {
	LoadingContext LoadingContext
	SelectedID     string
	MenuChoice     int
	HasAccounts    bool
	HasProjects    bool
	ConfirmText    string
	Error          error
}

// AppStateMachine wraps the stateless state machine
type AppStateMachine struct {
	machine *stateless.StateMachine
	context *StateMachineContext
}

// NewAppStateMachine creates and configures the application state machine
func NewAppStateMachine() *AppStateMachine {
	ctx := &StateMachineContext{}

	machine := stateless.NewStateMachine(StateLoading)

	// Configure Loading State
	machine.Configure(StateLoading).
		OnEntry(func(_ context.Context, args ...any) error {
			if len(args) > 0 {
				if loadingCtx, ok := args[0].(LoadingContext); ok {
					ctx.LoadingContext = loadingCtx
				}
			}
			return nil
		}).
		Permit(TriggerDataLoaded, StateMain).
		Permit(TriggerError, StateError)

	// Configure Main State
	machine.Configure(StateMain).
		Permit(TriggerLoadAccounts, StateLoading, func(_ context.Context, args ...any) bool {
			return !ctx.HasAccounts
		}).
		Permit(TriggerMenuChoice, StateAccounts, func(_ context.Context, args ...any) bool {
			return ctx.MenuChoice == 0 && ctx.HasAccounts
		}).
		Permit(TriggerLoadProjects, StateLoading, func(_ context.Context, args ...any) bool {
			return !ctx.HasProjects
		}).
		Permit(TriggerMenuChoice, StateProjects, func(_ context.Context, args ...any) bool {
			return ctx.MenuChoice == 1 && ctx.HasProjects
		}).
		Permit(TriggerMenuChoice, StateConfirming, func(_ context.Context, args ...any) bool {
			return ctx.MenuChoice == 2 // New Login
		}).
		Permit(TriggerMenuChoice, StateManualProject, func(_ context.Context, args ...any) bool {
			return ctx.MenuChoice == 3 // Manual Project
		})

	// Configure Accounts State
	machine.Configure(StateAccounts).
		Permit(TriggerAccountSelected, StateConfirming).
		Permit(TriggerGoBack, StateMain)

	// Configure Projects State
	machine.Configure(StateProjects).
		Permit(TriggerProjectSelected, StateConfirming).
		Permit(TriggerGoBack, StateMain)

	// Configure Manual Project State
	machine.Configure(StateManualProject).
		Permit(TriggerManualProjectEntry, StateConfirming).
		Permit(TriggerGoBack, StateMain)

	// Configure Confirming State
	machine.Configure(StateConfirming).
		OnEntry(func(_ context.Context, args ...any) error {
			if len(args) > 0 {
				if text, ok := args[0].(string); ok {
					ctx.ConfirmText = text
				}
			}
			return nil
		}).
		Permit(TriggerConfirmYes, StateProcessing).
		Permit(TriggerConfirmNo, StateMain)

	// Configure Processing State
	machine.Configure(StateProcessing).
		Permit(TriggerOperationComplete, StateMain).
		Permit(TriggerOperationFailed, StateError)

	// Configure Error State
	machine.Configure(StateError).
		OnEntry(func(_ context.Context, args ...any) error {
			if len(args) > 0 {
				if err, ok := args[0].(error); ok {
					ctx.Error = err
				}
			}
			return nil
		}).
		Permit(TriggerGoBack, StateMain)

	return &AppStateMachine{
		machine: machine,
		context: ctx,
	}
}

// GetState returns the current state
func (sm *AppStateMachine) GetState() AppState {
	return sm.machine.MustState().(AppState)
}

// GetContext returns the current context
func (sm *AppStateMachine) GetContext() *StateMachineContext {
	return sm.context
}

// Fire triggers a state transition
func (sm *AppStateMachine) Fire(trigger AppTrigger, args ...any) error {
	return sm.machine.Fire(trigger, args...)
}

// CanFire checks if a trigger can be fired
func (sm *AppStateMachine) CanFire(trigger AppTrigger) bool {
	canFire, _ := sm.machine.CanFire(trigger)
	return canFire
}

// SetMenuChoice sets the menu choice in context
func (sm *AppStateMachine) SetMenuChoice(choice int) {
	sm.context.MenuChoice = choice
}

// SetHasAccounts sets whether accounts are available
func (sm *AppStateMachine) SetHasAccounts(hasAccounts bool) {
	sm.context.HasAccounts = hasAccounts
}

// SetHasProjects sets whether projects are available
func (sm *AppStateMachine) SetHasProjects(hasProjects bool) {
	sm.context.HasProjects = hasProjects
}

// SetSelectedID sets the selected item ID
func (sm *AppStateMachine) SetSelectedID(id string) {
	sm.context.SelectedID = id
}

// GetLoadCommand returns the appropriate loading command based on context
func (sm *AppStateMachine) GetLoadCommand() tea.Cmd {
	switch sm.context.LoadingContext {
	case LoadingAccounts:
		return gcp.GetAllAccounts()
	case LoadingProjects:
		return gcp.GetSimpleProjects()
	default:
		return tea.Batch(
			gcp.CheckGcloud,
			gcp.GetActiveAccount(),
			gcp.GetActiveProject(),
			gcp.GetAllAccounts(),
			gcp.GetSimpleProjects(),
		)
	}
}

// GetActionCommand returns the appropriate action command based on context
func (sm *AppStateMachine) GetActionCommand() tea.Cmd {
	// Determine action based on previous state that led to confirmation
	state := sm.machine.MustState().(AppState)

	// Check what triggered the current processing state
	if state == StateProcessing {
		// For login action
		if sm.context.SelectedID == "" {
			return gcp.LoginNewAccount()
		}

		// For account switch (we know this if SelectedID looks like an email)
		if len(sm.context.SelectedID) > 0 && strings.Contains(sm.context.SelectedID, "@") {
			return gcp.SwitchAccount(sm.context.SelectedID)
		}

		// Otherwise it's a project switch
		return gcp.SwitchProject(sm.context.SelectedID)
	}
	return nil
}

// GetConfirmationText returns the confirmation text
func (sm *AppStateMachine) GetConfirmationText() string {
	return sm.context.ConfirmText
}

// GenerateDOTGraph generates a DOT graph representation of the state machine
func (sm *AppStateMachine) GenerateDOTGraph() string {
	return sm.machine.ToGraph()
}
