package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mathd/gcp-switcher/types"
)

const (
	commandTimeout = 5 * time.Second
	longTimeout    = 30 * time.Second
)

// CheckGcloud checks if gcloud CLI is installed
func CheckGcloud() tea.Msg {
	_, err := exec.LookPath("gcloud")
	return types.GcloudCheckMsg{Available: err == nil}
}

// GetActiveAccount retrieves the currently active GCP account
func GetActiveAccount() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return types.ErrMsg{Err: fmt.Errorf("command timed out: gcloud auth list")}
			}
			return types.ErrMsg{Err: err}
		}

		account := strings.TrimSpace(string(output))
		return types.ActiveAccountMsg{Account: account}
	}
}

// GetActiveProject retrieves the currently active GCP project
func GetActiveProject() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "config", "get-value", "project")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return types.ErrMsg{Err: fmt.Errorf("command timed out: gcloud config get-value project")}
			}
			return types.ErrMsg{Err: err}
		}

		project := strings.TrimSpace(string(output))
		return types.ActiveProjectMsg{Project: project}
	}
}

// GetAllAccounts retrieves all configured GCP accounts
func GetAllAccounts() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "auth", "list", "--format=json")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return types.ErrMsg{Err: fmt.Errorf("command timed out: gcloud auth list --format=json")}
			}
			return types.ErrMsg{Err: err}
		}

		var accounts []types.Account
		if err := json.Unmarshal(output, &accounts); err != nil {
			return types.ErrMsg{Err: err}
		}

		return types.AccountListMsg{Accounts: accounts}
	}
}

// GetSimpleProjects retrieves all accessible GCP projects
func GetSimpleProjects() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "projects", "list", "--format=json")
		output, err := cmd.CombinedOutput()

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return types.ProjectListMsg{Projects: []types.Project{}}
			}
			return types.ProjectListMsg{Projects: []types.Project{}}
		}

		var projects []types.Project
		if err := json.Unmarshal(output, &projects); err != nil {
			return types.ProjectListMsg{Projects: []types.Project{}}
		}

		return types.ProjectListMsg{Projects: projects}
	}
}

// SwitchAccount switches the active GCP account
func SwitchAccount(account string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "config", "set", "account", account)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()

		if err != nil {
			return types.OperationResultMsg{Success: false, Err: err}
		}

		return types.OperationResultMsg{Success: true, Err: fmt.Errorf("ACCOUNT_SWITCHED")}
	}
}

// LoginNewAccount initiates login for a new GCP account
func LoginNewAccount() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), longTimeout)
		defer cancel()

		loginCmd := exec.CommandContext(ctx, "gcloud", "auth", "login")
		loginCmd.Stdin = os.Stdin
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		err := loginCmd.Run()
		if err != nil {
			return types.OperationResultMsg{Success: false, Err: err}
		}

		adcCmd := exec.CommandContext(ctx, "gcloud", "auth", "application-default", "login")
		adcCmd.Stdin = os.Stdin
		adcCmd.Stdout = os.Stdout
		adcCmd.Stderr = os.Stderr
		err = adcCmd.Run()
		if err != nil {
			return types.OperationResultMsg{Success: false, Err: err}
		}

		return types.OperationResultMsg{Success: true}
	}
}

// SwitchProject switches the active GCP project
func SwitchProject(projectID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "gcloud", "config", "set", "project", projectID)
		err := cmd.Run()

		if err != nil {
			return types.OperationResultMsg{Success: false, Err: err}
		}

		return types.OperationResultMsg{Success: true}
	}
}
