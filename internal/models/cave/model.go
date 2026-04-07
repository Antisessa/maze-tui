package cave

import (
	"fmt"
	"maze/internal/domain"
	"maze/internal/models/shared"
	"maze/internal/ui"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type State int

const (
	StateMenu State = iota
	StateChooseDir
	StateForm
	StateConfirm
	StateLoad
	StateLoadRules
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

const (
	openFocusBirth = iota
	openFocusDeath
	openFocusCount
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
	OpenFocused     int

	// for loading
	CavePicker filepicker.Model

	ScreenWidth, ScreenHeight int
	StartDir                  string
	TickGen                   int
}

func New(startDir string, storage shared.CaveStorage) *Model {
	cavePicker := filepicker.New()
	cavePicker.CurrentDirectory = startDir
	cavePicker.FileAllowed = true
	cavePicker.DirAllowed = false
	cavePicker.SetHeight(10)

	dirPicker := filepicker.New()
	dirPicker.CurrentDirectory = startDir
	dirPicker.FileAllowed = false
	dirPicker.DirAllowed = true
	dirPicker.SetHeight(10)

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
		StepInterval:    1 * time.Second,
		CavePicker:      cavePicker,
		DirPicker:       dirPicker,
		NameInput:       name,
		RowsInput:       rows,
		ColsInput:       cols,
		InitChanceInput: chance,
		BirthRateInput:  birth,
		DeathRateInput:  death,
		Focused:         focusName,
		OpenFocused:     openFocusBirth,
	}

	m.setFocus(focusName)
	return m
}

func (m *Model) SetSize(width int, height int) {
	m.ScreenWidth = width
	m.ScreenHeight = height

	h := max(height-6, 5)
	m.DirPicker.SetHeight(h)
	m.DirPicker.ShowPermissions = width >= 70
	m.DirPicker.ShowSize = width >= 45

	m.CavePicker.SetHeight(h)
	m.CavePicker.ShowPermissions = width >= 70
	m.CavePicker.ShowSize = width >= 45
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func tickStepCmd(d time.Duration, gen int) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return stepTickMsg{Gen: gen}
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
		m.State = StateLoadRules
		m.setOpenFocus(openFocusBirth)
		return nil

	case GeneratedMsg:
		m.Cave = &msg.Cave
		m.Err = nil
		m.State = StateRun
		return tickStepCmd(m.StepInterval, m.TickGen)

	case ErrorMsg:
		m.Err = msg.Err
		return nil

	case stepTickMsg:
		if msg.Gen == m.TickGen && m.State == StateRun && m.Cave != nil {
			m.Cave.Step()
			return tickStepCmd(m.StepInterval, m.TickGen)
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
	case StateLoadRules:
		return m.updateLoadRules(msg)
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
		return m.CavePicker.Init()

	case "g":
		m.State = StateChooseDir
		m.Err = nil
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

func (m *Model) updateLoadRules(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc":
			m.State = StateLoad
			m.Err = nil
			m.Cave = nil
			return nil
		case "tab", "down":
			m.setOpenFocus((m.OpenFocused + 1) % openFocusCount)
			return nil
		case "shift+tab", "up":
			m.setOpenFocus((m.OpenFocused - 1 + openFocusCount) % openFocusCount)
			return nil
		case "enter":
			if m.OpenFocused == openFocusDeath {
				return m.submitLoadRules()
			}
			m.setOpenFocus((m.OpenFocused + 1) % openFocusCount)
			return nil
		case "ctrl+s":
			return m.submitLoadRules()
		}
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.BirthRateInput, cmd = m.BirthRateInput.Update(msg)
	cmds = append(cmds, cmd)
	m.DeathRateInput, cmd = m.DeathRateInput.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
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
		m.TickGen++
		return tickStepCmd(m.StepInterval, m.TickGen)

	case "-":
		m.StepInterval += 50 * time.Millisecond
		m.TickGen++
		return tickStepCmd(m.StepInterval, m.TickGen)
	}

	return nil
}

func (m *Model) submitLoadRules() tea.Cmd {
	if m.Cave == nil {
		m.Err = fmt.Errorf("cave is not loaded")
		return nil
	}

	birth, err := strconv.Atoi(strings.TrimSpace(m.BirthRateInput.Value()))
	if err != nil {
		m.Err = fmt.Errorf("birth threshold should be integer")
		return nil
	}

	death, err := strconv.Atoi(strings.TrimSpace(m.DeathRateInput.Value()))
	if err != nil {
		m.Err = fmt.Errorf("death threshold should be integer")
		return nil
	}

	if err = m.Cave.ConfigureRules(birth, death); err != nil {
		m.Err = err
		return nil
	}

	m.BirthRate = birth
	m.DeathRate = death
	m.Err = nil
	m.State = StateRun
	return tickStepCmd(m.StepInterval, m.TickGen)
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

func (m *Model) setOpenFocus(idx int) {
	m.OpenFocused = idx
	m.NameInput.Blur()
	m.RowsInput.Blur()
	m.ColsInput.Blur()
	m.InitChanceInput.Blur()
	m.BirthRateInput.Blur()
	m.DeathRateInput.Blur()

	switch idx {
	case openFocusBirth:
		m.BirthRateInput.Focus()
	case openFocusDeath:
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
		b.WriteString(ui.StackVert(
			ui.Title("Cave"),
			ui.MenuLine("o", "открыть файл (ctrl+o)"),
			ui.MenuLine("g", "сгенерировать"),
			ui.MenuLine("esc", "назад"),
		))

	case StateChooseDir:
		b.WriteString(ui.Title("Выберите директорию для сохранения"))
		b.WriteString(ui.Label("Текущая папка", m.DirPicker.CurrentDirectory) + "\n\n")
		b.WriteString(m.DirPicker.View())
		b.WriteString("\n\n")
		b.WriteString(ui.Hint("enter — перейти в папку | ctrl+s — выбрать текущую | r — стартовая | esc — назад"))

	case StateForm:
		b.WriteString(ui.Title("Параметры генерации"))
		b.WriteString(ui.Label("Папка", m.Dir) + "\n\n")
		b.WriteString(ui.FieldLine("Имя файла", m.NameInput.View(), m.Focused == focusName))
		b.WriteString("\n")
		b.WriteString(ui.FieldLine("Rows", m.RowsInput.View(), m.Focused == focusRows))
		b.WriteString("\n")
		b.WriteString(ui.FieldLine("Cols", m.ColsInput.View(), m.Focused == focusCols))
		b.WriteString("\n")
		b.WriteString(ui.FieldLine("Init chance", m.InitChanceInput.View(), m.Focused == focusChance))
		b.WriteString("\n")
		b.WriteString(ui.FieldLine("Birth threshold", m.BirthRateInput.View(), m.Focused == focusBirth))
		b.WriteString("\n")
		b.WriteString(ui.FieldLine("Death threshold", m.DeathRateInput.View(), m.Focused == focusDeath))
		b.WriteString("\n\n")
		b.WriteString(ui.Hint("tab/up/down — поле | enter — дальше | ctrl+s — подтвердить | esc — к папке"))

	case StateConfirm:
		b.WriteString(ui.Title("Подтвердите генерацию") + "\n")
		b.WriteString(ui.Label("Папка", m.Dir) + "\n")
		b.WriteString(ui.Label("Имя файла", m.Name) + "\n")
		b.WriteString(ui.Label("Rows", strconv.Itoa(m.Rows)) + "\n")
		b.WriteString(ui.Label("Cols", strconv.Itoa(m.Cols)) + "\n")
		b.WriteString(ui.Label("Init chance", fmt.Sprintf("%.2f", m.InitChance)) + "\n")
		b.WriteString(ui.Label("Birth threshold", strconv.Itoa(m.BirthRate)) + "\n")
		b.WriteString(ui.Label("Death threshold", strconv.Itoa(m.DeathRate)) + "\n\n")
		b.WriteString(ui.Hint("enter / ctrl+s — сгенерировать | esc — к форме"))

	case StateLoad:
		b.WriteString(ui.Title("Открытие файла пещеры"))
		b.WriteString(ui.Label("Папка", m.CavePicker.CurrentDirectory) + "\n\n")
		b.WriteString(m.CavePicker.View())
		if m.Err != nil {
			b.WriteString("\n" + ui.ErrorLine(m.Err.Error()) + "\n")
		}
		b.WriteString("\n" + ui.Hint("enter — открыть | r — стартовая папка | esc — назад"))

	case StateLoadRules:
		b.WriteString(ui.Title("Параметры эволюции"))
		b.WriteString(ui.FieldLine("Birth threshold", m.BirthRateInput.View(), m.OpenFocused == openFocusBirth))
		b.WriteString("\n")
		b.WriteString(ui.FieldLine("Death threshold", m.DeathRateInput.View(), m.OpenFocused == openFocusDeath))
		b.WriteString("\n\n")
		b.WriteString(ui.Hint("tab/up/down — поле | enter — дальше | ctrl+s — запустить | esc — к файлу"))

	case StateRun:
		b.WriteString(ui.Title("Пещера") + "\n")
		if m.Cave != nil {
			availW := max(1, m.ScreenWidth-2)
			availH := max(1, m.ScreenHeight-10)
			b.WriteString(ui.CaveBlock(renderCave(*m.Cave, availW, availH)))
		} else {
			b.WriteString(ui.Muted("empty cave"))
		}
		b.WriteString("\n\n" + ui.Label("Шаг", m.StepInterval.String()) + "\n")
		b.WriteString(ui.Hint("space — вручную | +/- — скорость | esc — назад"))
	}

	if m.Err != nil && m.State != StateLoad {
		b.WriteString("\n" + ui.ErrorLine(m.Err.Error()) + "\n")
	}

	return tea.NewView(b.String())
}

func renderCave(c domain.Cave, areaW, areaH int) string {
	cells := c.Cells()
	if len(cells) == 0 || len(cells[0]) == 0 {
		return "empty cave"
	}
	srcH := len(cells)
	srcW := len(cells[0])
	// 1 клетка = 2 символа по ширине
	viewportW := max(1, min(srcW, areaW/2))
	viewportH := max(1, min(srcH, areaH))
	// Центрированный crop для стабильного кадра
	startX := (srcW - viewportW) / 2
	startY := (srcH - viewportH) / 2
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	var b strings.Builder
	for y := 0; y < viewportH; y++ {
		row := cells[startY+y]
		for x := 0; x < viewportW; x++ {
			if row[startX+x] {
				b.WriteString("██")
			} else {
				b.WriteString("  ")
			}
		}
		if y != viewportH-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
