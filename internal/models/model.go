package models

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"maze/internal/domain"
	"maze/internal/models/cave"
	"maze/internal/models/generator"
	"maze/internal/models/opener"
	"maze/internal/models/shared"
	"maze/internal/ui"
	"path/filepath"
	"strings"
)

type State int

const (
	Start State = iota
	OpenMazeScreen
	GenerateMazeScreen
	CaveScreen
	FileLoading
	MazeLoaded
)

type Model struct {
	State State

	Maze *domain.Board
	Err  error

	Width, Height int

	SelectedFile string
	StartDir     string

	MazeStorage shared.MazeStorage
	CaveStorage shared.CaveStorage

	Open *opener.Model
	Gen  *generator.Model
	Cave *cave.Model
}

func InitModel(startDir string, mazeStorage shared.MazeStorage, caveStorage shared.CaveStorage) *Model {
	return &Model{
		State:       Start,
		StartDir:    startDir,
		MazeStorage: mazeStorage,
		CaveStorage: caveStorage,
		Open:        opener.NewModel(startDir),
		Gen:         generator.NewModel(startDir),
		Cave:        cave.New(startDir, caveStorage),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		if m.Open != nil {
			m.Open.SetSize(msg.Width, msg.Height)
		}
		if m.Gen != nil {
			m.Gen.SetSize(msg.Width, msg.Height)
		}

		// resize cave-submodel нужен именно сообщением
		if m.State == CaveScreen && m.Cave != nil {
			return m, m.Cave.Update(msg)
		}
		return m, nil

	// File loader cases START
	case opener.OpenMazeSelectedMsg:
		m.Err = nil
		m.SelectedFile = msg.Path
		m.State = FileLoading
		return m, opener.OpenMazeCmd(msg.Path, m.MazeStorage)

	case opener.OpenMazeCanceledMsg:
		m.State = Start
		return m, nil

	case opener.MazeLoadedMsg:
		m.Maze = &msg.Board
		m.SelectedFile = msg.Path
		m.State = MazeLoaded
		m.Err = nil
		return m, nil

	case opener.MazeLoadErrMsg:
		m.Err = fmt.Errorf("ошибка чтения файла %s: %w", msg.Path, msg.Err)
		m.State = OpenMazeScreen
		return m, nil
	// File loader cases END

	// Maze generator cases START
	case generator.GenerateCanceledMsg:
		m.State = Start
		return m, nil

	case generator.GenerateSubmitMsg:
		fullPath := filepath.Join(msg.Dir, msg.Name)
		m.SelectedFile = fullPath
		m.State = FileLoading
		m.Err = nil
		return m, generator.GenerateMazeCmd(fullPath, msg.Width, msg.Height, m.MazeStorage)

	case generator.MazeGeneratedMsg:
		m.SelectedFile = msg.Path
		m.State = FileLoading
		return m, opener.OpenMazeCmd(msg.Path, m.MazeStorage)

	case generator.MazeGenerateErrMsg:
		m.Err = fmt.Errorf("ошибка генерации файла %s: %w", msg.Path, msg.Err)
		m.State = GenerateMazeScreen
		return m, nil
	// Maze generator cases END

	// Cave cases START
	case cave.CancelMsg:
		m.State = Start
		return m, nil
		// Cave cases END
	}

	// Глобально только аварийный выход
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Сначала отдаём сообщение активному дочернему экрану
	switch m.State {
	case OpenMazeScreen:
		return m, m.Open.Update(msg)

	case GenerateMazeScreen:
		return m, m.Gen.Update(msg)

	case CaveScreen:
		return m, m.Cave.Update(msg)
	}

	// Потом хоткеи parent
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch m.State {
		case Start:
			switch key.String() {
			case "q", "й":
				return m, tea.Quit

			case "o", "щ", "ctrl+o":
				m.Err = nil
				m.Open = opener.NewModel(m.StartDir)
				m.Open.SetSize(m.Width, m.Height)
				m.State = OpenMazeScreen
				return m, m.Open.Init()

			case "g", "п":
				m.Err = nil
				m.Gen = generator.NewModel(m.StartDir)
				m.Gen.SetSize(m.Width, m.Height)
				m.State = GenerateMazeScreen
				return m, m.Gen.Init()

			case "c", "с":
				m.Err = nil
				if m.Cave == nil {
					m.Cave = cave.New(m.StartDir, m.CaveStorage)
				}
				m.Cave.SetSize(m.Width, m.Height)
				m.State = CaveScreen
				return m, m.Cave.Init()
			}

		case MazeLoaded:
			switch key.String() {
			case "q", "й":
				return m, tea.Quit
			case "esc":
				m.State = Start
				return m, nil
			}
		}
	}

	return m, nil
}

func (m *Model) View() tea.View {
	var b strings.Builder

	switch m.State {
	case Start:
		b.WriteString(ui.StackVert(
			ui.Title("Меню"),
			ui.MenuLine("o", "открыть файл"),
			ui.MenuLine("g", "сгенерировать лабиринт"),
			ui.MenuLine("c", "пещеры"),
			ui.MenuLine("q", "выйти"),
		))

	case OpenMazeScreen:
		if m.Err != nil {
			b.WriteString(ui.ErrorLine(m.Err.Error()) + "\n\n")
		}
		b.WriteString(m.Open.View().Content)

	case GenerateMazeScreen:
		if m.Err != nil {
			b.WriteString(ui.ErrorLine(m.Err.Error()) + "\n\n")
		}
		b.WriteString(m.Gen.View())

	case FileLoading:
		b.WriteString(ui.Title("Обработка файла"))
		if m.SelectedFile != "" {
			b.WriteString("\n" + ui.Muted(m.SelectedFile) + "\n")
		}

	case MazeLoaded:
		b.WriteString(ui.OkLine("Лабиринт загружен.") + "\n\n")
		if m.Maze != nil {
			availW := max(1, m.Width-4)
			availH := max(1, m.Height-8)
			b.WriteString(ui.MazeBlock(renderMaze(*m.Maze, availW, availH)))
		}
		b.WriteString("\n\n" + ui.Hint("Esc — назад | q — выйти"))

	case CaveScreen:
		if m.Cave != nil {
			b.WriteString(m.Cave.View().Content)
		}
	}

	v := tea.NewView(ui.Page(m.Width, ui.LayoutWithBrand(b.String())))
	v.AltScreen = true
	return v
}

// renderMaze рисует лабиринт в доступную символьную область.
func renderMaze(board domain.Board, areaW, areaH int) string {
	if board.Width <= 0 || board.Height <= 0 {
		return "empty maze"
	}

	wall := 1

	// min размеры при cell=1
	minW := (board.Width+1)*wall + board.Width
	minH := (board.Height+1)*wall + board.Height
	if areaW < minW || areaH < minH {
		return "terminal too small to render maze"
	}

	// Единый размер клетки для всего лабиринта
	cellByW := (areaW - (board.Width+1)*wall) / board.Width
	cellByH := (areaH - (board.Height+1)*wall) / board.Height
	cell := min(cellByW, cellByH)
	if cell < 1 {
		cell = 1
	}

	totalW := (board.Width+1)*wall + board.Width*cell
	totalH := (board.Height+1)*wall + board.Height*cell

	padLeft := (areaW - totalW) / 2
	padTop := (areaH - totalH) / 2

	canvas := make([][]rune, areaH)
	for y := range canvas {
		canvas[y] = make([]rune, areaW)
		for x := range canvas[y] {
			canvas[y][x] = ' '
		}
	}

	// Внешние границы
	fillRect(canvas, padLeft, padTop, totalW, wall, '█')
	fillRect(canvas, padLeft, padTop, wall, totalH, '█')

	// Вспомогательные функции координат
	cellX := func(c int) int { return padLeft + wall + c*(cell+wall) }
	cellY := func(r int) int { return padTop + wall + r*(cell+wall) }
	rightWallX := func(c int) int { return cellX(c) + cell }
	bottomWallY := func(r int) int { return cellY(r) + cell }

	for r := 0; r < board.Height; r++ {
		for c := 0; c < board.Width; c++ {
			cellData := board.Cells[r][c]

			if cellData.RightWall || c == board.Width-1 {
				fillRect(canvas, rightWallX(c), cellY(r)-wall, wall, cell+2*wall, '█')
			}
			if cellData.BottomWall || r == board.Height-1 {
				fillRect(canvas, cellX(c)-wall, bottomWallY(r), cell+2*wall, wall, '█')
			}
		}
	}

	var out strings.Builder
	for y := 0; y < areaH; y++ {
		out.WriteString(string(canvas[y]))
		if y != areaH-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func splitEven(total, parts int) []int {
	out := make([]int, parts)
	base := total / parts
	extra := total % parts

	for i := 0; i < parts; i++ {
		out[i] = base
		if i < extra {
			out[i]++
		}
	}
	return out
}

func fillRect(canvas [][]rune, x, y, w, h int, ch rune) {
	for yy := y; yy < y+h && yy < len(canvas); yy++ {
		if yy < 0 {
			continue
		}
		for xx := x; xx < x+w && xx < len(canvas[yy]); xx++ {
			if xx < 0 {
				continue
			}
			canvas[yy][xx] = ch
		}
	}
}
