package service

import (
	"bufio"
	"fmt"
	"maze/internal/domain"
	"os"
	"strconv"
	"strings"
)

const (
	HorizontalMatrix = iota
	VerticalMatrix
)

func Open(path string) (domain.Board, error) {
	file, err := os.Open(path)
	if err != nil {
		return domain.Board{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return domain.Board{}, fmt.Errorf("parse matrix dimension: %w", err)
		}
		return domain.Board{}, fmt.Errorf("empty file")
	}

	line := scanner.Text()
	dimensions := strings.Fields(line)
	if len(dimensions) != 2 {
		return domain.Board{}, fmt.Errorf("parse matrix dimension: %w", err)
	}

	height, err := strconv.Atoi(dimensions[0])
	if err != nil {
		return domain.Board{}, fmt.Errorf("parse matrix height: %w", err)
	} else if height <= 0 {
		return domain.Board{}, fmt.Errorf("matrix height should be greater than zero")
	}

	width, err := strconv.Atoi(dimensions[1])
	if err != nil {
		return domain.Board{}, fmt.Errorf("parse matrix width: %w", err)
	} else if width <= 0 {
		return domain.Board{}, fmt.Errorf("matrix width should be greater than zero")
	}

	rows := make([][]domain.Cell, height)
	for i := range rows {
		rows[i] = make([]domain.Cell, width)
	}

	err = scanMatrix(scanner, &rows, VerticalMatrix, height, width)
	if err != nil {
		return domain.Board{}, fmt.Errorf("scan vertical wall matrix: %w", err)
	}

	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return domain.Board{}, fmt.Errorf("parse \\n between matrixes: %w", err)
		} else {
			return domain.Board{}, fmt.Errorf("unexpected EOF while parse \\n between matrixes")
		}
	}
	_ = scanner.Text() // parse \n

	err = scanMatrix(scanner, &rows, HorizontalMatrix, height, width)
	if err != nil {
		return domain.Board{}, fmt.Errorf("scan horizontal wall matrix: %w", err)
	}

	if err = scanner.Err(); err != nil {
		return domain.Board{}, fmt.Errorf("scan ended with err: %w", err)
	}

	return domain.Board{
		Height: height,
		Width:  width,
		Cells:  rows,
	}, nil
}

func scanMatrix(scanner *bufio.Scanner, rows *[][]domain.Cell, direction int, height, width int) error {
	successRows := 0
	for i := 0; i < height && scanner.Scan(); i++ {
		rowStr := scanner.Text()
		cells := strings.Fields(rowStr)

		if len(cells) != width {
			return fmt.Errorf("%d row has %d cells but matrix's width is %d", i, len(cells), width)
		}

		for j := 0; j < len(cells); j++ {
			cell, err := strconv.Atoi(cells[j])
			if err != nil || (cell != 0 && cell != 1) {
				return fmt.Errorf("invalid cell '%s' with coords row: %d and col: %d", cells[j], i, j)
			} else if cell == 0 {
				continue
			}

			if direction == VerticalMatrix {
				(*rows)[i][j].RightWall = true
			} else if direction == HorizontalMatrix {
				(*rows)[i][j].BottomWall = true
			}
		}
		successRows++
	}
	if successRows != height {
		return fmt.Errorf("mismatch scanned rows and matrix height - got: %d want: %d", successRows+1, height)
	}
	return nil
}
