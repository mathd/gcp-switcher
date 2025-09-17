package internal

import (
	"testing"
)

func TestStateMachine(t *testing.T) {
	// Create a new state machine
	sm := NewAppStateMachine()

	// Test initial state
	if sm.GetState() != StateLoading {
		t.Errorf("Expected initial state to be StateLoading, got %v", sm.GetState())
	}

	// Test transition to main state
	err := sm.Fire(TriggerDataLoaded)
	if err != nil {
		t.Errorf("Failed to transition to main state: %v", err)
	}

	if sm.GetState() != StateMain {
		t.Errorf("Expected state to be StateMain after data loaded, got %v", sm.GetState())
	}

	// Test menu choice with guard conditions
	sm.SetMenuChoice(0) // Accounts
	sm.SetHasAccounts(false) // No accounts available

	if !sm.CanFire(TriggerLoadAccounts) {
		t.Error("Should be able to fire TriggerLoadAccounts when no accounts available")
	}

	// Test with accounts available
	sm.SetHasAccounts(true)
	sm.SetMenuChoice(0)

	if !sm.CanFire(TriggerMenuChoice) {
		t.Error("Should be able to fire TriggerMenuChoice when accounts are available")
	}

	// Test confirmation flow
	err = sm.Fire(TriggerMenuChoice)
	if err != nil {
		t.Errorf("Failed to transition to accounts state: %v", err)
	}

	if sm.GetState() != StateAccounts {
		t.Errorf("Expected state to be StateAccounts, got %v", sm.GetState())
	}

	// Test selection
	sm.SetSelectedID("test@example.com")
	err = sm.Fire(TriggerAccountSelected, "Switch to test@example.com?")
	if err != nil {
		t.Errorf("Failed to transition to confirming state: %v", err)
	}

	if sm.GetState() != StateConfirming {
		t.Errorf("Expected state to be StateConfirming, got %v", sm.GetState())
	}

	// Test confirmation text
	expectedText := "Switch to test@example.com?"
	if sm.GetConfirmationText() != expectedText {
		t.Errorf("Expected confirmation text '%s', got '%s'", expectedText, sm.GetConfirmationText())
	}

	// Test processing
	err = sm.Fire(TriggerConfirmYes)
	if err != nil {
		t.Errorf("Failed to transition to processing state: %v", err)
	}

	if sm.GetState() != StateProcessing {
		t.Errorf("Expected state to be StateProcessing, got %v", sm.GetState())
	}
}

func TestStateMachineGuardConditions(t *testing.T) {
	sm := NewAppStateMachine()

	// Move to main state
	sm.Fire(TriggerDataLoaded)

	// Test guard condition for loading accounts when accounts are available
	sm.SetHasAccounts(true)
	sm.SetMenuChoice(0)

	if sm.CanFire(TriggerLoadAccounts) {
		t.Error("Should not be able to load accounts when accounts are already available")
	}

	// Test guard condition for loading accounts when no accounts available
	sm.SetHasAccounts(false)
	if !sm.CanFire(TriggerLoadAccounts) {
		t.Error("Should be able to load accounts when no accounts are available")
	}
}

func TestStateMachineErrorHandling(t *testing.T) {
	sm := NewAppStateMachine()

	// Test invalid transition
	err := sm.Fire(TriggerConfirmYes) // Can't confirm from loading state
	if err == nil {
		t.Error("Expected error when firing invalid trigger")
	}

	// State should remain unchanged
	if sm.GetState() != StateLoading {
		t.Errorf("State should remain StateLoading after invalid transition, got %v", sm.GetState())
	}
}