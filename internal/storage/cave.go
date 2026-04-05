package storage

import (
	"bufio"
	"fmt"
	"maze/internal/domain"
	"os"
	"strconv"
	"strings"
)

type CaveStorage struct{}

func (c *CaveStorage) Open(path string) (domain.Cave, error) {
	file, err := os.Open(path)
	if err != nil {
		return domain.Cave{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return domain.Cave{}, fmt.Errorf("parse matrix dimension: %w", err)
		}
		return domain.Cave{}, fmt.Errorf("empty file")
	}

	line := scanner.Text()
	dimensions := strings.Fields(line)
	if len(dimensions) != 2 {
		return domain.Cave{}, fmt.Errorf("parse matrix dimension: %w", err)
	}

	rowsNum, err := strconv.Atoi(dimensions[0])
	if err != nil {
		return domain.Cave{}, fmt.Errorf("parse matrix rows: %w", err)
	} else if rowsNum <= 0 {
		return domain.Cave{}, fmt.Errorf("matrix rows should be greater than zero")
	}

	colsNum, err := strconv.Atoi(dimensions[1])
	if err != nil {
		return domain.Cave{}, fmt.Errorf("parse matrix cols: %w", err)
	} else if colsNum <= 0 {
		return domain.Cave{}, fmt.Errorf("matrix cols should be greater than zero")
	}

	rows := make([][]bool, rowsNum)
	for i := range rows {
		rows[i] = make([]bool, colsNum)
	}

	err = scanCave(scanner, rows, rowsNum, colsNum)
	if err != nil {
		return domain.Cave{}, fmt.Errorf("scan matrix: %w", err)
	}

	if err = scanner.Err(); err != nil {
		return domain.Cave{}, fmt.Errorf("scan ended with err: %w", err)
	}

	newCave := domain.Cave{
		Rows: rowsNum,
		Cols: colsNum,
	}

	err = newCave.LoadCave(rows)
	if err != nil {
		return domain.Cave{}, fmt.Errorf("load cave: %w", err)
	}

	return newCave, nil
}

func scanCave(scanner *bufio.Scanner, cells [][]bool, rows, cols int) error {
	successRows := 0
	for i := 0; i < rows && scanner.Scan(); i++ {
		rowStr := scanner.Text()
		scannedCells := strings.Fields(rowStr)

		if len(scannedCells) != cols {
			return fmt.Errorf("%d row has %d cells but matrix's width is %d", i, len(scannedCells), cols)
		}

		for j := 0; j < len(scannedCells); j++ {
			cell, err := strconv.Atoi(scannedCells[j])
			if err != nil || (cell != 0 && cell != 1) {
				return fmt.Errorf("invalid cell '%s' with coords row: %d and col: %d", scannedCells[j], i, j)
			}

			cells[i][j] = cell == 1
		}
		successRows++
	}
	if successRows != rows {
		return fmt.Errorf("mismatch scanned rows and matrix height - got: %d want: %d", successRows+1, rows)
	}
	return nil
}

func (c *CaveStorage) Save(path string, cave domain.Cave) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file while save: %w", err)
	}
	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close file after save: %w", closeErr)
		}
	}()

	wr := bufio.NewWriter(file)
	defer func() {
		flushErr := wr.Flush()
		if err == nil && flushErr != nil {
			err = fmt.Errorf("flush writer after save: %w", flushErr)
		}
	}()

	if _, err = fmt.Fprintf(wr, "%d %d\n", cave.Rows, cave.Cols); err != nil {
		return fmt.Errorf("write dimensions: %w", err)
	}

	cells := cave.Cells()
	for i := range cells {
		for j := range cells[i] {
			if cells[i][j] {
				_, err = fmt.Fprintf(wr, "1")
				if err != nil {
					return fmt.Errorf("write cell i:%d, j:%d - %w", i, j, err)
				}
			} else {
				_, err = fmt.Fprintf(wr, "0")
				if err != nil {
					return fmt.Errorf("write cell i:%d, j:%d - %w", i, j, err)
				}
			}

			if j != len(cells[i])-1 {
				_, err = fmt.Fprintf(wr, " ")
				if err != nil {
					return fmt.Errorf("write whitespace after i:%d, j:%d - %w", i, j, err)
				}
			}
		}

		if i != len(cells)-1 {
			_, err = fmt.Fprintf(wr, "\n")
			if err != nil {
				return fmt.Errorf("write new line after line #:%d - %w", i, err)
			}
		}
	}
	return nil
}
