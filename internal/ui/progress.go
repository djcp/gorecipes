package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// stepState tracks the display status of one pipeline step.
type stepState int

const (
	stepPending stepState = iota
	stepActive
	stepDone
	stepFailed
)

// progressStep is one row in the progress display.
type progressStep struct {
	label string
	state stepState
}

// StepUpdate is sent by the pipeline goroutine to advance the progress display.
type StepUpdate struct {
	Step  int
	Label string
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type tickMsg struct{}
type stepMsg StepUpdate

func tickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func waitForStep(ch <-chan StepUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-ch
		if !ok {
			return nil
		}
		return stepMsg(update)
	}
}

func renderStep(step progressStep, tick int) string {
	var icon, label string
	switch step.state {
	case stepPending:
		icon = MutedStyle.Render("○")
		label = MutedStyle.Render(step.label)
	case stepActive:
		frame := spinnerFrames[tick%len(spinnerFrames)]
		icon = lipgloss.NewStyle().Foreground(ColorPrimary).Render(frame)
		label = BoldStyle.Render(step.label)
	case stepDone:
		icon = SuccessStyle.Render("✓")
		label = step.label
	case stepFailed:
		icon = ErrorStyle.Render("✗")
		label = ErrorStyle.Render(step.label)
	}
	return fmt.Sprintf("%s  %s", icon, label)
}
