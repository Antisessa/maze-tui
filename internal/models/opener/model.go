package opener

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"maze/internal/ui"
)

type Model struct {
	Picker filepicker.Model
	Err    error

	Width, Height int
	StartDir      string
}

func NewModel(startDir string) *Model {
	fp := filepicker.New()
	fp.CurrentDirectory = startDir
	fp.FileAllowed = true
	fp.DirAllowed = false
	fp.ShowPermissions = true
	fp.ShowSize = true

	return &Model{
		Picker:   fp,
		StartDir: startDir,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.Picker.Init()
}

func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height

	h := max(height-6, 5)
	m.Picker.SetHeight(h)
	m.Picker.ShowPermissions = width >= 70
	m.Picker.ShowSize = width >= 45
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			return func() tea.Msg { return OpenMazeCanceledMsg{} }

		case "r":
			m.Err = nil
			m.Picker.CurrentDirectory = m.StartDir
			return m.Picker.Init()
		}
	}

	var cmd tea.Cmd
	m.Picker, cmd = m.Picker.Update(msg)

	if ok, path := m.Picker.DidSelectDisabledFile(msg); ok {
		m.Err = fmt.Errorf("нельзя выбрать файл: %s", path)
		return cmd
	}

	if ok, path := m.Picker.DidSelectFile(msg); ok {
		return func() tea.Msg {
			return OpenMazeSelectedMsg{Path: path}
		}
	}

	return cmd
}

func (m *Model) View() tea.View {
	var b strings.Builder
	b.WriteString(ui.Title("Выберите файл лабиринта"))
	b.WriteString(ui.Hint("Enter — выбрать | r — в начало | Esc — назад") + "\n\n")

	if m.Err != nil {
		b.WriteString(ui.ErrorLine(m.Err.Error()) + "\n\n")
	}

	b.WriteString(m.Picker.View())
	return tea.NewView(b.String())
}
