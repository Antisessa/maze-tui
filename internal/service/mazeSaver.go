package service

import (
	"bufio"
	"fmt"
	"maze/internal/domain"
	"os"
	"strings"
)

func SaveMaze(path string, board domain.Board) error {
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

	if _, err = fmt.Fprintf(wr, "%d %d\n", board.Height, board.Width); err != nil {
		return fmt.Errorf("write dimensions: %w", err)
	}

	builderVertMatrix := strings.Builder{}
	builderHorMatrix := strings.Builder{}
	for i := range board.Cells {
		for j := range board.Cells[i] {
			if board.Cells[i][j].RightWall {
				builderVertMatrix.Write([]byte("1"))
			} else {
				builderVertMatrix.Write([]byte("0"))
			}

			if board.Cells[i][j].BottomWall {
				builderHorMatrix.Write([]byte("1"))
			} else {
				builderHorMatrix.Write([]byte("0"))
			}

			if j != len(board.Cells[i])-1 {
				builderVertMatrix.Write([]byte(" "))
				builderHorMatrix.Write([]byte(" "))
			}
		}

		if i != len(board.Cells)-1 {
			builderVertMatrix.Write([]byte("\n"))
			builderHorMatrix.Write([]byte("\n"))
		}
	}

	if _, err = wr.WriteString(builderVertMatrix.String()); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	if _, err = wr.WriteString("\n\n"); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	if _, err = wr.WriteString(builderHorMatrix.String()); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}
	return nil
}
