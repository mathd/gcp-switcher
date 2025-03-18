package types

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

func NewItem(title, description string, isActive bool, id string) Item {
	return Item{
		title:       title,
		description: description,
		isActive:    isActive,
		id:          id,
	}
}

func (i Item) Title() string {
	if i.isActive {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("159")).Bold(true).Render(i.title + " (ACTIVE)")
	}
	return i.title
}

func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title + i.description }
func (i Item) ID() string          { return i.id }

// Message Types
type SpinnerMsg tea.Msg
type ErrMsg struct{ Err error }
type GcloudCheckMsg struct{ Available bool }
type ActiveAccountMsg struct{ Account string }
type ActiveProjectMsg struct{ Project string }
type AccountListMsg struct{ Accounts []Account }
type ProjectListMsg struct{ Projects []Project }
type OperationResultMsg struct {
	Success bool
	Err     error
}
type FallbackTimerMsg struct{ TimeoutSeconds int }

func (e ErrMsg) Error() string { return e.Err.Error() }
