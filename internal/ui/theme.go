package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// AppName отображается на всех экранах из одного места (см. LayoutWithBrand в model).
const AppName = "A1_Maze_Go"

// Палитра в духе Tokyo Night — хорошо читается в тёмных терминалах.
var (
	colorTitle   = lipgloss.Color("#7aa2f7")
	colorAccent  = lipgloss.Color("#bb9af7")
	colorText    = lipgloss.Color("#c0caf5")
	colorMuted   = lipgloss.Color("#565f89")
	colorErr     = lipgloss.Color("#f7768e")
	colorOk      = lipgloss.Color("#9ece6a")
	colorBorder  = lipgloss.Color("#414868")
	colorMaze    = lipgloss.Color("#e0af68")
	colorCaveOn  = lipgloss.Color("#7dcfff")
	colorCaveOff = lipgloss.Color("#1f2335")
)

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTitle).
			MarginBottom(1)

	styleHint = lipgloss.NewStyle().Foreground(colorMuted)

	styleKey = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)

	styleErr = lipgloss.NewStyle().Foreground(colorErr)

	styleLabel = lipgloss.NewStyle().Foreground(colorText)

	styleDim = lipgloss.NewStyle().Foreground(colorMuted)

	styleOk = lipgloss.NewStyle().Foreground(colorOk)
)

// Title заголовок экрана.
func Title(s string) string {
	return styleTitle.Render(s)
}

// Hint строка подсказок (клавиши).
func Hint(s string) string {
	return styleHint.Render(s)
}

// Key выделяет одну клавишу или шорткат.
func Key(s string) string {
	return styleKey.Render(s)
}

// ErrorLine форматирует текст ошибки.
func ErrorLine(s string) string {
	return styleErr.Render("Ошибка: " + s)
}

// Muted приглушённый текст.
func Muted(s string) string {
	return styleDim.Render(s)
}

// OkLine успех / статус.
func OkLine(s string) string {
	return styleOk.Render(s)
}

// Фиксированные ширины колонок, чтобы UI не "плыл".
const (
	menuKeyWidth   = 16
	labelKeyWidth  = 18
	focusMarkWidth = 2
)

// MenuLine одна строка меню: ключ — описание.
func MenuLine(key, desc string) string {
	k := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorAccent).
		Width(menuKeyWidth).
		Align(lipgloss.Left).
		Render(key)
	d := styleDim.Render(" — " + desc)
	return k + d
}

// BrandBanner верхняя строка с названием приложения (без рамки Page).
func BrandBanner() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(colorMuted).
		MarginBottom(1).
		Render(AppName)
}

// StackVert склеивает блоки по вертикали с выравниванием влево (для меню после Title).
func StackVert(lines ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// LayoutWithBrand добавляет баннер приложения над телом экрана.
func LayoutWithBrand(body string) string {
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return BrandBanner()
	}
	return StackVert(BrandBanner(), body)
}

// Label поле «имя: значение».
func Label(name, value string) string {
	key := styleLabel.Width(labelKeyWidth).Align(lipgloss.Left).Render(name)
	sep := styleDim.Render(" : ")
	val := lipgloss.NewStyle().Foreground(colorText).Render(value)
	return lipgloss.JoinHorizontal(lipgloss.Left, key, sep, val)
}

// FieldLine строка формы с фокусом (как renderField в cave).
func FieldLine(label, value string, focused bool) string {
	var mark string
	if focused {
		mark = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Width(focusMarkWidth).
			Align(lipgloss.Left).
			Render(">")
	} else {
		mark = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(focusMarkWidth).
			Align(lipgloss.Left).
			Render(" ")
	}
	key := styleLabel.Width(labelKeyWidth).Align(lipgloss.Left).Render(label)
	sep := styleDim.Render(" : ")
	val := lipgloss.NewStyle().Foreground(colorText).Render(value)
	return lipgloss.JoinHorizontal(lipgloss.Left, mark, key, sep, val)
}

// Page оборачивает контент в рамку с учётом ширины терминала.
func Page(termWidth int, content string) string {
	w := termWidth - 4
	if w < 24 {
		w = 24
	}
	if termWidth <= 0 {
		w = 80
	}
	return lipgloss.NewStyle().
		Width(w).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Foreground(colorText).
		Render(strings.TrimRight(content, "\n"))
}

// MazeBlock подсветка ASCII-лабиринта.
func MazeBlock(s string) string {
	return lipgloss.NewStyle().Foreground(colorMaze).Render(strings.TrimRight(s, "\n"))
}

// CaveBlock подсветка области пещеры (стены и фон).
func CaveBlock(s string) string {
	return lipgloss.NewStyle().
		Foreground(colorCaveOn).
		Background(colorCaveOff).
		Render(strings.TrimRight(s, "\n"))
}
