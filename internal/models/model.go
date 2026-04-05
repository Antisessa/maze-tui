package models

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"maze/internal/domain"
	"maze/internal/models/generator"
	"maze/internal/models/opener"
	"path/filepath"
	"strings"
)

type State int

const (
	Start State = iota
	OpenMazeScreen
	GenerateMazeScreen
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

	Open *opener.Model
	Gen  *generator.Model
}

func InitModel(startDir string) *Model {
	return &Model{
		State:    Start,
		StartDir: startDir,
		Open:     opener.NewModel(startDir),
		Gen:      generator.NewModel(startDir),
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
		return m, nil

	// File loader cases START
	case opener.OpenMazeSelectedMsg:
		m.Err = nil
		m.SelectedFile = msg.Path
		m.State = FileLoading
		return m, opener.OpenMazeCmd(msg.Path)

	case opener.OpenMazeCanceledMsg: // отменили открытие лабиринта
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
		return m, generator.GenerateMazeCmd(fullPath, msg.Width, msg.Height)

	case generator.MazeGeneratedMsg:
		// при желании можно сразу открыть сгенерированный файл:
		m.SelectedFile = msg.Path
		m.State = FileLoading
		return m, opener.OpenMazeCmd(msg.Path)

	case generator.MazeGenerateErrMsg:
		m.Err = fmt.Errorf("ошибка генерации файла %s: %w", msg.Path, msg.Err)
		m.State = GenerateMazeScreen
		return m, nil
	} // Maze generator cases START

	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "q", "й", "ctrl+c":
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

		case "esc":
			if m.State == MazeLoaded {
				m.State = Start
				return m, nil
			}
		}
	}

	switch m.State {
	case OpenMazeScreen:
		return m, m.Open.Update(msg)
	case GenerateMazeScreen:
		return m, m.Gen.Update(msg)
	}

	return m, nil
}

func (m *Model) View() tea.View {
	var b strings.Builder

	b.WriteString("A1_Maze_Go\n\n")

	switch m.State {
	case Start:
		b.WriteString("o — открыть файл\n")
		b.WriteString("g — сгенерировать лабиринт\n")
		b.WriteString("q — выйти\n")

	case OpenMazeScreen:
		if m.Err != nil {
			b.WriteString("Ошибка: " + m.Err.Error() + "\n\n")
		}
		b.WriteString(m.Open.View().Content)

	case GenerateMazeScreen:
		if m.Err != nil {
			b.WriteString("Ошибка: " + m.Err.Error() + "\n\n")
		}
		b.WriteString(m.Gen.View())

	case FileLoading:
		b.WriteString("Обработка файла...\n\n")
		if m.SelectedFile != "" {
			b.WriteString(m.SelectedFile + "\n")
		}

	case MazeLoaded:
		b.WriteString("Лабиринт загружен.\n\n")
		if m.Maze != nil {
			availW := max(1, m.Width-2)
			availH := max(1, m.Height-6)
			b.WriteString(renderMaze(*m.Maze, availW, availH))
		}
		b.WriteString("\n\nEsc — назад | q — выйти\n")
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// renderMaze рисует лабиринт в доступную символьную область.
// В TUI нельзя буквально обеспечить "500x500 px" и "2 px wall":
// здесь это переводится в заполнение доступной области терминала символами. :contentReference[oaicite:1]{index=1}
func renderMaze(board domain.Board, areaW, areaH int) string {
	if board.Width <= 0 || board.Height <= 0 {
		return "empty maze"
	}

	// В TUI одна "толщина стены" = 1 символ.
	// Если терминал широкий и хочешь сделать стены толще,
	// можно увеличить до 2, но тогда резко вырастут требования к размеру окна.
	wall := 1

	// Квадратная область рендера, аналог "500x500".
	side := min(areaW, areaH)
	if side <= 0 {
		return "no space to render maze"
	}
	drawW, drawH := side, side

	// Минимально: каждая ячейка хотя бы 1x1 + стены
	minW := (board.Width+1)*wall + board.Width
	minH := (board.Height+1)*wall + board.Height
	if drawW < minW || drawH < minH {
		return "terminal too small to render maze"
	}

	// Аналог формулы из ТЗ:
	// всё пространство = стены + внутренности ячеек
	cellW := splitEven(drawW-(board.Width+1)*wall, board.Width)
	cellH := splitEven(drawH-(board.Height+1)*wall, board.Height)

	// Позиции вертикальных и горизонтальных границ
	xb := make([]int, board.Width+1)  // x начала каждой вертикальной границы
	yb := make([]int, board.Height+1) // y начала каждой горизонтальной границы

	for c := 0; c < board.Width; c++ {
		xb[c+1] = xb[c] + wall + cellW[c]
	}
	for r := 0; r < board.Height; r++ {
		yb[r+1] = yb[r] + wall + cellH[r]
	}

	totalW := xb[board.Width] + wall
	totalH := yb[board.Height] + wall

	canvas := make([][]rune, totalH)
	for y := range canvas {
		canvas[y] = make([]rune, totalW)
		for x := range canvas[y] {
			canvas[y][x] = ' '
		}
	}

	// Верхняя и левая внешние границы всегда есть
	fillRect(canvas, 0, 0, totalW, wall, '█') // top
	fillRect(canvas, 0, 0, wall, totalH, '█') // left

	for r := 0; r < board.Height; r++ {
		for c := 0; c < board.Width; c++ {
			cell := board.Cells[r][c]

			// Правая граница ячейки
			// Для надёжности последнюю колонку можно форсировать как внешнюю границу
			if cell.RightWall || c == board.Width-1 {
				fillRect(
					canvas,
					xb[c+1], // x начала правой границы
					yb[r],   // от верхней границы строки
					wall,
					cellH[r]+2*wall, // до нижней границы строки включительно
					'█',
				)
			}

			// Нижняя граница ячейки
			// Для надёжности последнюю строку можно форсировать как внешнюю границу
			if cell.BottomWall || r == board.Height-1 {
				fillRect(
					canvas,
					xb[c],           // от левой границы столбца
					yb[r+1],         // y начала нижней границы
					cellW[c]+2*wall, // до правой границы столбца включительно
					wall,
					'█',
				)
			}
		}
	}

	var out strings.Builder
	for y := 0; y < totalH; y++ {
		out.WriteString(string(canvas[y]))
		if y != totalH-1 {
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
