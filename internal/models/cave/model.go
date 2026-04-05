package cave

import (
	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"fmt"
	"maze/internal/domain"
	"maze/internal/models/shared"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type State int

const (
	StateMenu State = iota
	StateChooseDir
	StateForm
	StateConfirm
	StateLoad
	StateRun
)

const (
	focusName = iota
	focusRows
	focusCols
	focusChance
	focusBirth
	focusDeath
	focusCount
)

type Model struct {
	Cave *domain.Cave

	Dir, Name            string
	Rows, Cols           int
	InitChance           float32
	BirthRate, DeathRate int

	State        State
	Storage      shared.CaveStorage
	Err          error
	StepInterval time.Duration

	// generation
	DirPicker       filepicker.Model
	NameInput       textinput.Model
	RowsInput       textinput.Model
	ColsInput       textinput.Model
	InitChanceInput textinput.Model
	BirthRateInput  textinput.Model
	DeathRateInput  textinput.Model
	Focused         int

	// for loading
	CavePicker filepicker.Model

	ScreenWidth, ScreenHeight int
	StartDir                  string
}

func New(startDir string, storage shared.CaveStorage) *Model {
	loadPicker := filepicker.New()
	loadPicker.CurrentDirectory = startDir
	loadPicker.FileAllowed = true
	loadPicker.DirAllowed = false

	dirPicker := filepicker.New()
	dirPicker.CurrentDirectory = startDir
	dirPicker.FileAllowed = false
	dirPicker.DirAllowed = true

	name := textinput.New()
	name.Placeholder = "Имя файла"
	name.SetValue("Cave_5x5")
	name.CharLimit = 128

	rows := textinput.New()
	rows.Placeholder = "Rows"
	rows.SetValue("5")

	cols := textinput.New()
	cols.Placeholder = "Cols"
	cols.SetValue("5")

	chance := textinput.New()
	chance.Placeholder = "Init chance"
	chance.SetValue("0.45")

	birth := textinput.New()
	birth.Placeholder = "Birth threshold"
	birth.SetValue("4")

	death := textinput.New()
	death.Placeholder = "Death threshold"
	death.SetValue("3")

	m := &Model{
		StartDir:        startDir,
		Storage:         storage,
		State:           StateMenu,
		StepInterval:    300 * time.Millisecond,
		CavePicker:      loadPicker,
		DirPicker:       dirPicker,
		NameInput:       name,
		RowsInput:       rows,
		ColsInput:       cols,
		InitChanceInput: chance,
		BirthRateInput:  birth,
		DeathRateInput:  death,
		Focused:         focusName,
	}

	m.setFocus(focusName)
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func tickStepCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return stepTickMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ScreenWidth = msg.Width
		m.ScreenHeight = msg.Height

		h := max(msg.Height-10, 8)
		m.CavePicker.SetHeight(h)
		m.DirPicker.SetHeight(h / 2)
		return nil

	case LoadedMsg:
		m.Cave = &msg.Cave
		m.Err = nil
		m.State = StateRun
		return tickStepCmd(m.StepInterval)

	case GeneratedMsg:
		m.Cave = &msg.Cave
		m.Err = nil
		m.State = StateRun
		return tickStepCmd(m.StepInterval)

	case ErrorMsg:
		m.Err = msg.Err
		return nil

	case stepTickMsg:
		if m.State == StateRun && m.Cave != nil {
			m.Cave.Step()
			return tickStepCmd(m.StepInterval)
		}
		return nil
	}

	switch m.State {
	case StateMenu:
		return m.updateMenu(msg)
	case StateChooseDir:
		return m.updateChooseDir(msg)
	case StateForm:
		return m.updateForm(msg)
	case StateConfirm:
		return m.updateConfirm(msg)
	case StateLoad:
		return m.updateLoad(msg)
	case StateRun:
		return m.updateRun(msg)
	default:
		return nil
	}
}

func (m *Model) updateMenu(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch key.String() {
	case "o", "ctrl+o":
		m.State = StateLoad
		m.Err = nil
		m.CavePicker.CurrentDirectory = m.StartDir
		return m.CavePicker.Init()

	case "g":
		m.State = StateChooseDir
		m.Err = nil
		m.DirPicker.CurrentDirectory = m.StartDir
		return m.DirPicker.Init()

	case "esc":
		return func() tea.Msg { return CancelMsg{} }
	}

	return nil
}

func (m *Model) updateChooseDir(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			m.State = StateMenu
			m.Err = nil
			return nil
		case "r":
			m.DirPicker.CurrentDirectory = m.StartDir
			return m.DirPicker.Init()
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
			m.State = StateForm
			m.Err = nil
			m.setFocus(focusName)
			return nil
		}
	}

	return cmd
}

func (m *Model) updateForm(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			m.State = StateChooseDir
			m.Err = nil
			return nil

		case "tab", "down":
			m.setFocus((m.Focused + 1) % focusCount)
			return nil

		case "shift+tab", "up":
			m.setFocus((m.Focused - 1 + focusCount) % focusCount)
			return nil

		case "enter":
			if m.Focused == focusDeath {
				return m.submitForm()
			}
			m.setFocus((m.Focused + 1) % focusCount)
			return nil

		case "ctrl+s":
			return m.submitForm()
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.NameInput, cmd = m.NameInput.Update(msg)
	cmds = append(cmds, cmd)

	m.RowsInput, cmd = m.RowsInput.Update(msg)
	cmds = append(cmds, cmd)

	m.ColsInput, cmd = m.ColsInput.Update(msg)
	cmds = append(cmds, cmd)

	m.InitChanceInput, cmd = m.InitChanceInput.Update(msg)
	cmds = append(cmds, cmd)

	m.BirthRateInput, cmd = m.BirthRateInput.Update(msg)
	cmds = append(cmds, cmd)

	m.DeathRateInput, cmd = m.DeathRateInput.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *Model) submitForm() tea.Cmd {
	params, err := m.generateParams()
	if err != nil {
		m.Err = err
		return nil
	}

	m.Name = strings.TrimSpace(m.NameInput.Value())
	if m.Name == "" {
		m.Name = "generated_cave"
	}
	m.Rows = params.Rows
	m.Cols = params.Cols
	m.InitChance = params.Chance
	m.BirthRate = params.Birth
	m.DeathRate = params.Death

	m.Err = nil
	m.State = StateConfirm
	return nil
}

func (m *Model) updateConfirm(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch key.String() {
	case "esc":
		m.State = StateForm
		m.Err = nil
		return nil

	case "enter", "ctrl+s":
		params := GenerateParams{
			Chance: m.InitChance,
			Birth:  m.BirthRate,
			Death:  m.DeathRate,
			Rows:   m.Rows,
			Cols:   m.Cols,
		}
		m.Err = nil
		return GenerateAndSaveCmd(m.Storage, m.outputPath(), params)
	}

	return nil
}

func (m *Model) updateLoad(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			m.State = StateMenu
			m.Err = nil
			return nil
		case "r":
			m.CavePicker.CurrentDirectory = m.StartDir
			return m.CavePicker.Init()
		}
	}
	var cmd tea.Cmd
	m.CavePicker, cmd = m.CavePicker.Update(msg)

	if ok, path := m.CavePicker.DidSelectFile(msg); ok {
		m.Err = nil
		return OpenCaveCmd(m.Storage, path)
	}

	if ok, path := m.DirPicker.DidSelectDisabledFile(msg); ok {
		m.Err = fmt.Errorf("нельзя выбрать: %s", path)
		return cmd
	}

	return cmd
}

func (m *Model) updateRun(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch key.String() {
	case "esc":
		m.State = StateMenu
		m.Err = nil
		m.Cave = nil
		return nil

	case " ":
		if m.Cave != nil {
			m.Cave.Step()
		}
		return nil

	case "+":
		if m.StepInterval > 50*time.Millisecond {
			m.StepInterval -= 50 * time.Millisecond
		}
		return nil

	case "-":
		m.StepInterval += 50 * time.Millisecond
		return nil
	}

	return nil
}

func (m *Model) setFocus(idx int) {
	m.Focused = idx

	m.NameInput.Blur()
	m.RowsInput.Blur()
	m.ColsInput.Blur()
	m.InitChanceInput.Blur()
	m.BirthRateInput.Blur()
	m.DeathRateInput.Blur()

	switch idx {
	case focusName:
		m.NameInput.Focus()
	case focusRows:
		m.RowsInput.Focus()
	case focusCols:
		m.ColsInput.Focus()
	case focusChance:
		m.InitChanceInput.Focus()
	case focusBirth:
		m.BirthRateInput.Focus()
	case focusDeath:
		m.DeathRateInput.Focus()
	}
}

func (m *Model) generateParams() (GenerateParams, error) {
	rows, err := strconv.Atoi(strings.TrimSpace(m.RowsInput.Value()))
	if err != nil {
		return GenerateParams{}, fmt.Errorf("rows should be integer")
	}

	cols, err := strconv.Atoi(strings.TrimSpace(m.ColsInput.Value()))
	if err != nil {
		return GenerateParams{}, fmt.Errorf("cols should be integer")
	}

	chance64, err := strconv.ParseFloat(strings.TrimSpace(m.InitChanceInput.Value()), 32)
	if err != nil {
		return GenerateParams{}, fmt.Errorf("init chance should be float")
	}

	birth, err := strconv.Atoi(strings.TrimSpace(m.BirthRateInput.Value()))
	if err != nil {
		return GenerateParams{}, fmt.Errorf("birth threshold should be integer")
	}

	death, err := strconv.Atoi(strings.TrimSpace(m.DeathRateInput.Value()))
	if err != nil {
		return GenerateParams{}, fmt.Errorf("death threshold should be integer")
	}

	return GenerateParams{
		Chance: float32(chance64),
		Birth:  birth,
		Death:  death,
		Rows:   rows,
		Cols:   cols,
	}, nil
}

func (m *Model) outputPath() string {
	dir := m.Dir
	if strings.TrimSpace(dir) == "" {
		dir = m.DirPicker.CurrentDirectory
	}

	name := strings.TrimSpace(m.Name)
	if name == "" {
		name = strings.TrimSpace(m.NameInput.Value())
	}
	if name == "" {
		name = "generated_cave"
	}

	return filepath.Join(dir, name)
}

func (m *Model) View() tea.View {
	var b strings.Builder

	switch m.State {
	case StateMenu:
		b.WriteString("Cave\n\n")
		b.WriteString("o / ctrl+o — открыть файл\n")
		b.WriteString("g          — сгенерировать\n")
		b.WriteString("esc        — назад\n")

	case StateChooseDir:
		b.WriteString("Выберите директорию для сохранения\n\n")
		b.WriteString("Текущая папка:\n")
		b.WriteString(m.DirPicker.CurrentDirectory + "\n\n")
		b.WriteString(m.DirPicker.View())
		b.WriteString("\n\n")
		b.WriteString("enter  — перейти в папку\n")
		b.WriteString("ctrl+s — выбрать текущую папку\n")
		b.WriteString("r      — стартовая директория\n")
		b.WriteString("esc    — назад\n")

	case StateForm:
		b.WriteString("Параметры генерации\n\n")
		b.WriteString("Папка: " + m.Dir + "\n\n")
		b.WriteString(renderField("Имя файла", m.NameInput.View(), m.Focused == focusName))
		b.WriteString("\n")
		b.WriteString(renderField("Rows", m.RowsInput.View(), m.Focused == focusRows))
		b.WriteString("\n")
		b.WriteString(renderField("Cols", m.ColsInput.View(), m.Focused == focusCols))
		b.WriteString("\n")
		b.WriteString(renderField("Init chance", m.InitChanceInput.View(), m.Focused == focusChance))
		b.WriteString("\n")
		b.WriteString(renderField("Birth threshold", m.BirthRateInput.View(), m.Focused == focusBirth))
		b.WriteString("\n")
		b.WriteString(renderField("Death threshold", m.DeathRateInput.View(), m.Focused == focusDeath))
		b.WriteString("\n\n")
		b.WriteString("tab/up/down — сменить поле\n")
		b.WriteString("enter       — дальше\n")
		b.WriteString("ctrl+s      — к подтверждению\n")
		b.WriteString("esc         — назад к выбору директории\n")

	case StateConfirm:
		b.WriteString("Подтвердите генерацию\n\n")
		b.WriteString("Папка: " + m.Dir + "\n")
		b.WriteString("Имя файла: " + m.Name + "\n")
		b.WriteString("Rows: " + strconv.Itoa(m.Rows) + "\n")
		b.WriteString("Cols: " + strconv.Itoa(m.Cols) + "\n")
		b.WriteString("Init chance: " + fmt.Sprintf("%.2f", m.InitChance) + "\n")
		b.WriteString("Birth threshold: " + strconv.Itoa(m.BirthRate) + "\n")
		b.WriteString("Death threshold: " + strconv.Itoa(m.DeathRate) + "\n\n")
		b.WriteString("enter / ctrl+s — сгенерировать и сохранить\n")
		b.WriteString("esc            — назад к форме\n")

	case StateLoad:
		b.WriteString("Открытие файла пещеры\n\n")
		b.WriteString(m.CavePicker.CurrentDirectory + "\n\n")
		b.WriteString(m.CavePicker.View())
		b.WriteString("\n\n")
		b.WriteString("enter — открыть файл\n")
		b.WriteString("r     — вернуться в стартовую директорию\n")
		b.WriteString("esc   — назад\n")

	case StateRun:
		b.WriteString("Пещера\n\n")
		if m.Cave != nil {
			availW := max(1, m.ScreenWidth-2)
			availH := max(1, m.ScreenHeight-8)
			b.WriteString(renderCave(*m.Cave, availW, availH))
		} else {
			b.WriteString("empty cave")
		}
		b.WriteString("\n\nspace — шаг вручную | +/- — скорость | esc — назад\n")
	}

	if m.Err != nil {
		b.WriteString("\nОшибка: ")
		b.WriteString(m.Err.Error())
		b.WriteString("\n")
	}

	return tea.NewView(b.String())
}

func renderField(label, value string, focused bool) string {
	prefix := "  "
	if focused {
		prefix = "> "
	}
	return prefix + label + ": " + value
}

func renderCave(c domain.Cave, areaW, areaH int) string {
	cells := c.Cells()
	if len(cells) == 0 || len(cells[0]) == 0 {
		return "empty cave"
	}

	srcH := len(cells)
	srcW := len(cells[0])

	// В терминале символ примерно выше, чем шире,
	// поэтому одну логическую клетку рисуем двумя символами по ширине.
	dstH := max(1, min(areaH, srcH))
	dstWCells := max(1, min(areaW/2, srcW))

	var b strings.Builder

	for y := 0; y < dstH; y++ {
		srcY := y * srcH / dstH

		for x := 0; x < dstWCells; x++ {
			srcX := x * srcW / dstWCells

			if cells[srcY][srcX] {
				b.WriteString("██")
			} else {
				b.WriteString("  ")
			}
		}

		if y != dstH-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
