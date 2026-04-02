package models

import (
	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"fmt"
	"maze/internal/domain"
	"maze/internal/service"
	"os"
	"strings"
)

type State int

const (
	Start State = iota
	FilePicker
	FileLoading
	MazeLoaded
)

type mazeLoadedMsg struct {
	Path  string
	Board domain.Board
}

type mazeLoadErrMsg struct {
	Path string
	Err  error
}

func openMazeCmd(path string) tea.Cmd {
	return func() tea.Msg {
		board, err := service.Open(path)
		if err != nil {
			return mazeLoadErrMsg{
				Path: path,
				Err:  err,
			}
		}

		return mazeLoadedMsg{
			Path:  path,
			Board: board,
		}
	}
}

type Model struct {
	Maze   *domain.Board
	State  State
	Picker filepicker.Model
	Err    error

	Width, Height int

	StartDir     string
	SelectedFile string
}

func InitModel() *Model {
	model := &Model{
		Maze:  nil,
		State: Start,
	}

	wd, err := os.Getwd()
	if err == nil {
		model.StartDir = wd
	}

	fp := filepicker.New()
	fp.CurrentDirectory = model.StartDir
	fp.FileAllowed = true
	fp.DirAllowed = false
	fp.ShowPermissions = true
	fp.ShowSize = true

	model.Picker = fp

	return model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Update размера терминала
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Оставляем место под заголовок и подсказки
		h := max(msg.Height-6, 5)

		m.Picker.SetHeight(h)
		m.Picker.ShowPermissions = msg.Width >= 70
		m.Picker.ShowSize = msg.Width >= 45
		return m, nil

	case mazeLoadedMsg:
		m.Maze = &msg.Board
		m.SelectedFile = msg.Path
		m.State = MazeLoaded
		m.Err = nil
		return m, nil

	case mazeLoadErrMsg:
		m.Err = fmt.Errorf("ошибка чтения файла %s: %w", msg.Path, msg.Err)
		m.SelectedFile = msg.Path
		m.State = FilePicker
		return m, nil

	// Key pressed
	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			if m.State == FilePicker {
				m.Err = nil
				m.Picker.CurrentDirectory = m.StartDir
				return m, m.Picker.Init()
			}
		case "o", "щ", "ctrl+o":
			m.State = FilePicker
			m.Err = nil
			return m, m.Picker.Init()

		case "esc":
			if m.State == FilePicker {
				m.State = Start
				return m, nil
			} else if m.State == MazeLoaded {
				m.State = FilePicker
				m.SelectedFile = ""
				return m, nil
			}
		case "q", "й", "ctrl+c":
			return m, tea.Quit
		}
	}

	// Пока открыт file-picker, все сообщения нужно отдавать ему
	if m.State == FilePicker {
		var cmd tea.Cmd
		m.Picker, cmd = m.Picker.Update(msg)

		if ok, path := m.Picker.DidSelectDisabledFile(msg); ok {
			m.Err = fmt.Errorf("нельзя выбрать файл: %s", path)
			return m, cmd
		}

		if ok, path := m.Picker.DidSelectFile(msg); ok {
			m.SelectedFile = path
			m.State = FileLoading
			m.Err = nil
			return m, openMazeCmd(path)
		}

		return m, cmd
	}

	return m, nil
}

func (m *Model) View() tea.View {
	var b strings.Builder

	b.WriteString("A1_Maze_Go\n\n")

	switch m.State {
	case Start:
		b.WriteString("Press 'o' to open file or 'q' to quit.\n")

	case FilePicker:
		b.WriteString("Выберите файл лабиринта\n")
		b.WriteString("Используйте стрелки для навигации по директориям\n")
		b.WriteString("Enter — выбрать файл | Esc — главное меню | q — quit\n")

		if m.Err != nil {
			b.WriteString("Ошибка: " + m.Err.Error() + "\n\n")
		}

		b.WriteString(m.Picker.View())

	case FileLoading:
		b.WriteString("Загрузка лабиринта...\n\n")
		if m.SelectedFile != "" {
			b.WriteString(m.SelectedFile + "\n")
		}
		b.WriteString("\nPress 'q' to quit.\n")

	case MazeLoaded:
		b.WriteString("Лабиринт загружен.\n\n")

		if m.Maze != nil {
			// Выделяем область под рендер внутри окна.
			// 4 строки: заголовок + пустые строки + нижняя подсказка.
			availW := max(1, m.Width-2)
			availH := max(1, m.Height-6)
			b.WriteString(renderMaze(*m.Maze, availW, availH))
		}

		b.WriteString("\n\nPress 'o' to open another file | 'q' to quit.\n")
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
