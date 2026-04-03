package generator

import (
	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"fmt"
	"strconv"
	"strings"
)

type GenerateStep int

const (
	StepChooseDir GenerateStep = iota
	StepForm
	StepConfirm
)

type Model struct {
	Step GenerateStep

	Dir, Name     string
	Height, Width int

	Err error

	DirPicker   filepicker.Model
	NameInput   textinput.Model
	WidthInput  textinput.Model
	HeightInput textinput.Model

	Focused int

	ScreenWidth, ScreenHeight int
	StartDir                  string
}

func NewModel(startDir string) *Model {
	dirPicker := filepicker.New()
	dirPicker.CurrentDirectory = startDir
	dirPicker.FileAllowed = false
	dirPicker.DirAllowed = true

	nameInput := textinput.New()
	nameInput.Placeholder = "maze.txt"
	nameInput.SetValue("maze.txt")

	widthInput := textinput.New()
	widthInput.Placeholder = "width"

	heightInput := textinput.New()
	heightInput.Placeholder = "height"

	return &Model{
		Step:        StepChooseDir,
		StartDir:    startDir,
		Dir:         startDir,
		DirPicker:   dirPicker,
		NameInput:   nameInput,
		WidthInput:  widthInput,
		HeightInput: heightInput,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.DirPicker.Init()
}

func (m *Model) SetSize(width, height int) {
	m.ScreenWidth = width
	m.ScreenHeight = height

	h := max(height-6, 5)
	m.DirPicker.SetHeight(h)
	m.DirPicker.ShowPermissions = width >= 70
	m.DirPicker.ShowSize = width >= 45
}

func (m *Model) parseForm() error {
	name := strings.TrimSpace(m.NameInput.Value())
	if name == "" {
		return fmt.Errorf("имя файла не может быть пустым")
	}

	w, err := strconv.Atoi(strings.TrimSpace(m.WidthInput.Value()))
	if err != nil || w <= 0 {
		return fmt.Errorf("width должен быть положительным числом")
	}

	h, err := strconv.Atoi(strings.TrimSpace(m.HeightInput.Value()))
	if err != nil || h <= 0 {
		return fmt.Errorf("height должен быть положительным числом")
	}

	m.Name = name
	m.Width = w
	m.Height = h
	return nil
}

func (m *Model) focusInput(idx int) {
	m.Focused = idx

	m.NameInput.Blur()
	m.WidthInput.Blur()
	m.HeightInput.Blur()

	switch idx {
	case 0:
		m.NameInput.Focus()
	case 1:
		m.WidthInput.Focus()
	case 2:
		m.HeightInput.Focus()
	}
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch m.Step {
	case StepChooseDir:
		return m.updateChooseDir(msg)
	case StepForm:
		return m.updateForm(msg)
	case StepConfirm:
		return m.updateConfirm(msg)
	}
	return nil
}

func (m *Model) View() string {
	var b strings.Builder

	switch m.Step {
	case StepChooseDir:
		b.WriteString("Выберите директорию для сохранения\n")
		b.WriteString("Enter — подтвердить текущую папку | Esc — назад\n\n")

		if m.Err != nil {
			b.WriteString("Ошибка: " + m.Err.Error() + "\n\n")
		}

		b.WriteString("Текущая папка:\n")
		b.WriteString(m.DirPicker.CurrentDirectory + "\n\n")
		b.WriteString(m.DirPicker.View())

	case StepForm:
		b.WriteString("Параметры генерации\n")
		b.WriteString("Tab — следующее поле | Enter — далее | Esc — к выбору папки\n\n")

		b.WriteString("Папка: " + m.Dir + "\n\n")
		b.WriteString("Имя файла: " + m.NameInput.View() + "\n")
		b.WriteString("Width:     " + m.WidthInput.View() + "\n")
		b.WriteString("Height:    " + m.HeightInput.View() + "\n")

		if m.Err != nil {
			b.WriteString("\nОшибка: " + m.Err.Error())
		}

	case StepConfirm:
		b.WriteString("Подтвердите генерацию\n\n")
		b.WriteString("Папка:  " + m.Dir + "\n")
		b.WriteString("Файл:   " + m.Name + "\n")
		b.WriteString("Width:  " + strconv.Itoa(m.Width) + "\n")
		b.WriteString("Height: " + strconv.Itoa(m.Height) + "\n\n")
		b.WriteString("Enter — сгенерировать | Esc — назад")
	}

	return b.String()
}

func (m *Model) updateChooseDir(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			return func() tea.Msg { return GenerateCanceledMsg{} }
		}
	}

	var cmd tea.Cmd
	m.DirPicker, cmd = m.DirPicker.Update(msg)

	if ok, path := m.DirPicker.DidSelectFile(msg); ok {
		_ = path
	}

	if ok, path := m.DirPicker.DidSelectDisabledFile(msg); ok {
		m.Err = fmt.Errorf("нельзя выбрать: %s", path)
		return cmd
	}

	// Для directory picker обычно enter на папке просто выбирает её как current dir.
	// Поэтому удобнее сделать отдельную клавишу подтверждения.
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "enter":
			m.Dir = m.DirPicker.CurrentDirectory
			m.Step = StepForm
			m.Err = nil
			m.focusInput(0)
			return nil
		}
	}

	return cmd
}

func (m *Model) updateForm(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			m.Step = StepChooseDir
			m.Err = nil
			return nil

		case "tab", "shift+tab":
			if key.String() == "tab" {
				m.focusInput((m.Focused + 1) % 3)
			} else {
				m.focusInput((m.Focused + 2) % 3)
			}
			return nil

		case "enter":
			if err := m.parseForm(); err != nil {
				m.Err = err
				return nil
			}
			m.Err = nil
			m.Step = StepConfirm
			return nil
		}
	}

	var cmd tea.Cmd
	switch m.Focused {
	case 0:
		m.NameInput, cmd = m.NameInput.Update(msg)
	case 1:
		m.WidthInput, cmd = m.WidthInput.Update(msg)
	case 2:
		m.HeightInput, cmd = m.HeightInput.Update(msg)
	}

	return cmd
}

func (m *Model) updateConfirm(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			m.Step = StepForm
			return nil

		case "enter":
			return func() tea.Msg {
				return GenerateSubmitMsg{
					Dir:    m.Dir,
					Name:   m.Name,
					Width:  m.Width,
					Height: m.Height,
				}
			}
		}
	}
	return nil
}
