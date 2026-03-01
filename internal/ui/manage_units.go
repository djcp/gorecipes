package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/djcp/gorecipes/internal/db"
	"github.com/jmoiron/sqlx"
)

type manageUnitsPhase int

const (
	manageUnitsPhaseBrowse manageUnitsPhase = iota
	manageUnitsPhaseEdit
	manageUnitsPhaseMerge
	manageUnitsPhaseMergeConfirm
	manageUnitsPhaseResult
)

// manageUnitsModel is the TUI model for serving-unit management.
type manageUnitsModel struct {
	sqlDB *sqlx.DB

	phase manageUnitsPhase

	units  []db.UnitWithCount
	cursor int
	offset int

	// Edit.
	editInput textinput.Model

	// Merge.
	mergeList       []db.UnitWithCount
	mergeCursor     int
	mergeSourceName string
	mergeTargetName string

	// Result.
	resultMsg string
	resultErr bool

	width  int
	height int
}

func newManageUnitsModel(sqlDB *sqlx.DB) manageUnitsModel {
	return manageUnitsModel{sqlDB: sqlDB, width: 80, height: 24}
}

func (m *manageUnitsModel) loadUnits() error {
	units, err := db.ListUnitsWithCount(m.sqlDB)
	if err != nil {
		return err
	}
	m.units = units
	return nil
}

func (m manageUnitsModel) Init() tea.Cmd { return textinput.Blink }

func (m manageUnitsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	if m.phase == manageUnitsPhaseEdit {
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m manageUnitsModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case manageUnitsPhaseBrowse:
		return m.handleBrowseKey(msg)
	case manageUnitsPhaseEdit:
		return m.handleEditKey(msg)
	case manageUnitsPhaseMerge:
		return m.handleMergeKey(msg)
	case manageUnitsPhaseMergeConfirm:
		return m.handleMergeConfirmKey(msg)
	case manageUnitsPhaseResult:
		if err := m.loadUnits(); err != nil {
			m.resultMsg = "Error reloading: " + err.Error()
			m.resultErr = true
			return m, nil
		}
		m.cursor = 0
		m.offset = 0
		m.phase = manageUnitsPhaseBrowse
		return m, nil
	}
	return m, nil
}

func (m manageUnitsModel) visibleRows() int {
	// Banner(4) + header(1) + footer(2) = 7
	v := m.height - 7
	if v < 1 {
		v = 1
	}
	return v
}

func (m manageUnitsModel) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}
	case "down", "j":
		if m.cursor < len(m.units)-1 {
			m.cursor++
			visible := m.visibleRows()
			if m.cursor >= m.offset+visible {
				m.offset = m.cursor - visible + 1
			}
		}
	case "e":
		if len(m.units) == 0 {
			return m, nil
		}
		unit := m.units[m.cursor]
		ti := textinput.New()
		ti.SetValue(unit.Name)
		ti.Width = m.width - 16
		ti.Focus()
		m.editInput = ti
		m.phase = manageUnitsPhaseEdit
		return m, textinput.Blink
	case "m":
		if len(m.units) < 2 {
			return m, nil
		}
		source := m.units[m.cursor]
		mergeList := make([]db.UnitWithCount, 0, len(m.units)-1)
		for _, u := range m.units {
			if u.Name != source.Name {
				mergeList = append(mergeList, u)
			}
		}
		m.mergeList = mergeList
		m.mergeCursor = 0
		m.mergeSourceName = source.Name
		m.phase = manageUnitsPhaseMerge
	}
	return m, nil
}

func (m manageUnitsModel) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.phase = manageUnitsPhaseBrowse
		return m, nil
	case "enter":
		newName := strings.TrimSpace(m.editInput.Value())
		if newName == "" {
			m.phase = manageUnitsPhaseBrowse
			return m, nil
		}
		unit := m.units[m.cursor]
		if err := db.RenameUnit(m.sqlDB, unit.Name, newName); err != nil {
			m.resultMsg = "Error renaming: " + err.Error()
			m.resultErr = true
		} else {
			m.resultMsg = fmt.Sprintf("Renamed '%s' → '%s'", unit.Name, newName)
			m.resultErr = false
		}
		m.phase = manageUnitsPhaseResult
		return m, nil
	}
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m manageUnitsModel) handleMergeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.phase = manageUnitsPhaseBrowse
		return m, nil
	case "up", "k":
		if m.mergeCursor > 0 {
			m.mergeCursor--
		}
	case "down", "j":
		if m.mergeCursor < len(m.mergeList)-1 {
			m.mergeCursor++
		}
	case "enter", " ":
		target := m.mergeList[m.mergeCursor]
		m.mergeTargetName = target.Name
		m.phase = manageUnitsPhaseMergeConfirm
	}
	return m, nil
}

func (m manageUnitsModel) handleMergeConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "n":
		m.phase = manageUnitsPhaseMerge
		return m, nil
	case "y", "enter":
		if err := db.MergeUnit(m.sqlDB, m.mergeSourceName, m.mergeTargetName); err != nil {
			m.resultMsg = "Error merging: " + err.Error()
			m.resultErr = true
		} else {
			m.resultMsg = fmt.Sprintf("Merged '%s' into '%s'", m.mergeSourceName, m.mergeTargetName)
			m.resultErr = false
		}
		m.phase = manageUnitsPhaseResult
	}
	return m, nil
}

func (m manageUnitsModel) View() string {
	if m.width == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(renderManageUnitsBanner(m.width))
	sb.WriteString("\n")

	switch m.phase {
	case manageUnitsPhaseBrowse:
		sb.WriteString(m.viewBrowse())
	case manageUnitsPhaseEdit:
		sb.WriteString(m.viewEdit())
	case manageUnitsPhaseMerge:
		sb.WriteString(m.viewMerge())
	case manageUnitsPhaseMergeConfirm:
		sb.WriteString(m.viewMergeConfirm())
	case manageUnitsPhaseResult:
		sb.WriteString(m.viewResult())
	}

	return sb.String()
}

func (m manageUnitsModel) viewBrowse() string {
	var sb strings.Builder
	sb.WriteString("\n")

	visible := m.visibleRows()
	end := m.offset + visible
	if end > len(m.units) {
		end = len(m.units)
	}

	if len(m.units) == 0 {
		sb.WriteString(MutedStyle.Render("  No serving units found."))
		sb.WriteString("\n")
	} else {
		for i := m.offset; i < end; i++ {
			u := m.units[i]
			selected := i == m.cursor
			count := fmt.Sprintf("%d use", u.Count)
			if u.Count != 1 {
				count += "s"
			}
			row := fmt.Sprintf("  %-30s  %s", truncate(u.Name, 30), MutedStyle.Render(count))
			if selected {
				sb.WriteString(HighlightStyle.Width(m.width - 2).Render(row))
			} else {
				sb.WriteString(row)
			}
			sb.WriteString("\n")
		}
	}

	used := strings.Count(sb.String(), "\n")
	if fill := m.height - used - 3; fill > 0 {
		sb.WriteString(strings.Repeat("\n", fill))
	}
	sb.WriteString("\n")
	sb.WriteString(renderUnitsFooterBrowse(m.width))
	return sb.String()
}

func (m manageUnitsModel) viewEdit() string {
	var sb strings.Builder
	unit := m.units[m.cursor]
	sb.WriteString("\n")
	sb.WriteString(MutedStyle.Render(fmt.Sprintf("  Rename unit '%s':", unit.Name)))
	sb.WriteString("\n\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1).
		MarginLeft(4).
		Render(m.editInput.View())
	sb.WriteString(box)
	sb.WriteString("\n")
	used := strings.Count(sb.String(), "\n")
	if fill := m.height - used - 3; fill > 0 {
		sb.WriteString(strings.Repeat("\n", fill))
	}
	sb.WriteString("\n")
	sb.WriteString(renderUnitsFooterEdit(m.width))
	return sb.String()
}

func (m manageUnitsModel) viewMerge() string {
	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(MutedStyle.Render(fmt.Sprintf("  Merge '%s' into…", m.mergeSourceName)))
	sb.WriteString("\n\n")

	visible := m.height - 8
	if visible < 1 {
		visible = 1
	}
	offset := 0
	if m.mergeCursor >= visible {
		offset = m.mergeCursor - visible + 1
	}
	end := offset + visible
	if end > len(m.mergeList) {
		end = len(m.mergeList)
	}

	for i := offset; i < end; i++ {
		u := m.mergeList[i]
		selected := i == m.mergeCursor
		count := fmt.Sprintf("%d use", u.Count)
		if u.Count != 1 {
			count += "s"
		}
		row := fmt.Sprintf("  %-30s  %s", truncate(u.Name, 30), MutedStyle.Render(count))
		if selected {
			sb.WriteString(HighlightStyle.Width(m.width - 2).Render(row))
		} else {
			sb.WriteString(row)
		}
		sb.WriteString("\n")
	}

	used := strings.Count(sb.String(), "\n")
	if fill := m.height - used - 3; fill > 0 {
		sb.WriteString(strings.Repeat("\n", fill))
	}
	sb.WriteString("\n")
	sb.WriteString(renderUnitsFooterMerge(m.width))
	return sb.String()
}

func (m manageUnitsModel) viewMergeConfirm() string {
	var sb strings.Builder
	sb.WriteString("\n\n")

	inner := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("Merge units?"),
		"",
		MutedStyle.Render(fmt.Sprintf("Merge '%s' into '%s'?", m.mergeSourceName, m.mergeTargetName)),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorWarning).
		Padding(1, 3).
		Render(inner)

	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, box))
	sb.WriteString("\n")

	used := strings.Count(sb.String(), "\n")
	if fill := m.height - used - 3; fill > 0 {
		sb.WriteString(strings.Repeat("\n", fill))
	}
	sb.WriteString("\n")
	sb.WriteString(renderUnitsFooterConfirm(m.width))
	return sb.String()
}

func (m manageUnitsModel) viewResult() string {
	var sb strings.Builder
	sb.WriteString("\n\n")

	style := SuccessStyle
	if m.resultErr {
		style = ErrorStyle
	}

	inner := style.Render(m.resultMsg)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 3).
		Render(inner)

	sb.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, box))
	sb.WriteString("\n")

	used := strings.Count(sb.String(), "\n")
	if fill := m.height - used - 3; fill > 0 {
		sb.WriteString(strings.Repeat("\n", fill))
	}
	sb.WriteString("\n")
	sb.WriteString(renderUnitsFooterResult(m.width))
	return sb.String()
}

func renderManageUnitsBanner(width int) string {
	breadcrumb := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render(
			"🍳  gorecipes  " +
				MutedStyle.Render("/") +
				"  manage  " +
				MutedStyle.Render("/") +
				"  " +
				lipgloss.NewStyle().
					Bold(false).
					Foreground(lipgloss.Color("#5C4A3C")).
					Render("serving units"),
		)

	title := lipgloss.NewStyle().
		Padding(1, 2).
		Render(breadcrumb)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ColorBorder).
		Width(width - 2).
		Render(title)
}

func renderUnitsFooterBrowse(width int) string {
	keys := []string{"↑/↓ navigate", "e edit", "m merge", "esc back"}
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorBorder).
		Width(width - 2).
		Render(footerLine(keys, width-2))
}

func renderUnitsFooterEdit(width int) string {
	keys := []string{"enter save", "esc cancel"}
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorBorder).
		Width(width - 2).
		Render(footerLine(keys, width-2))
}

func renderUnitsFooterMerge(width int) string {
	keys := []string{"↑/↓ select target", "enter confirm", "esc cancel"}
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorBorder).
		Width(width - 2).
		Render(footerLine(keys, width-2))
}

func renderUnitsFooterConfirm(width int) string {
	yKey := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("y confirm")
	line := "  " + yKey + "   " + MutedStyle.Render("n / esc cancel")
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorWarning).
		Width(width - 2).
		Render(line)
}

func renderUnitsFooterResult(width int) string {
	keys := []string{"any key continue"}
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorBorder).
		Width(width - 2).
		Render(footerLine(keys, width-2))
}

// RunManageUnitsUI runs the serving-units management TUI.
func RunManageUnitsUI(sqlDB *sqlx.DB) error {
	m := newManageUnitsModel(sqlDB)
	if err := m.loadUnits(); err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
