package domain

import (
	"fmt"
	"math/rand"
)

const (
	BorderWidth = 1
)

type Cave struct {
	initChance                     float32
	birthThreshold, deathThreshold int
	CurrentCells                   [][]bool
	nextCells                      [][]bool
	Rows, Cols                     int
}

func InitCave(chance float32, birth, death, rows, cols int) (Cave, error) {
	if chance > 1.0 || chance <= 0 {
		return Cave{}, fmt.Errorf("cell initialization chance should be in range (0; 1]")
	} else if birth < 0 || birth > 7 {
		return Cave{}, fmt.Errorf("birth threshold should be in range [0; 7]")
	} else if death < 0 || death > 7 {
		return Cave{}, fmt.Errorf("death threshold should be in range [0; 7]")
	} else if rows <= 0 || rows > 50 {
		return Cave{}, fmt.Errorf("cave's Rows should be in range [1; 50]")
	} else if cols <= 0 || cols > 50 {
		return Cave{}, fmt.Errorf("cave's Cols should be in range [1; 50]")
	}

	newCave := Cave{
		initChance:     chance,
		birthThreshold: birth,
		deathThreshold: death,
		Rows:           rows,
		Cols:           cols,
	}

	newCave.generateCave()

	return newCave, nil
}

func (c *Cave) LoadCave(cells [][]bool) error {
	c.CurrentCells = make([][]bool, c.Rows+BorderWidth*2)
	c.nextCells = make([][]bool, c.Rows+BorderWidth*2)

	k, n := 0, 0
	for i := range c.CurrentCells {
		c.CurrentCells[i] = make([]bool, c.Cols+BorderWidth*2)
		c.nextCells[i] = make([]bool, c.Cols+BorderWidth*2)
		n = 0
		for j := range c.CurrentCells[i] {
			isBorder := i < BorderWidth ||
				i >= len(c.CurrentCells)-BorderWidth ||
				j < BorderWidth ||
				j >= len(c.CurrentCells[i])-BorderWidth

			if isBorder {
				c.CurrentCells[i][j] = true
			} else {
				c.CurrentCells[i][j] = cells[k][n]
				n++
			}
		}
		k++
	}
	if k != c.Rows-1 || n != c.Cols-1 {
		params := fmt.Sprintln("rows:", c.Rows, "cols:", c.Cols, "k:", k, "n:", n)
		return fmt.Errorf("dimension mismatch %s", params)
	}
	return nil
}

// generateCave инициализирует пещеру, обрамляя матрицу клеток рамкой "живой" границы
func (c *Cave) generateCave() {
	c.CurrentCells = make([][]bool, c.Rows+BorderWidth*2)
	c.nextCells = make([][]bool, c.Rows+BorderWidth*2)
	for i := range c.CurrentCells {
		c.CurrentCells[i] = make([]bool, c.Cols+BorderWidth*2)
		c.nextCells[i] = make([]bool, c.Cols+BorderWidth*2)
		for j := range c.CurrentCells[i] {
			isBorder := i < BorderWidth ||
				i >= len(c.CurrentCells)-BorderWidth ||
				j < BorderWidth ||
				j >= len(c.CurrentCells[i])-BorderWidth

			if isBorder {
				c.CurrentCells[i][j] = true
			} else {
				c.CurrentCells[i][j] = rand.Float32() < c.initChance
			}
		}
	}
}

func (c *Cave) cellularAutomatonStep() {
	for i := range c.CurrentCells {
		for j := range c.CurrentCells[i] {
			isBorder := i < BorderWidth ||
				i >= len(c.CurrentCells)-BorderWidth ||
				j < BorderWidth ||
				j >= len(c.CurrentCells[i])-BorderWidth

			if isBorder { // для клеток границы - отдельная проверка
				c.nextCells[i][j] = true
				continue
			}
			alive := c.checkNeighborhood(i, j)

			if c.CurrentCells[i][j] == true && alive < c.deathThreshold {
				c.nextCells[i][j] = false
			} else if c.CurrentCells[i][j] == false && alive > c.birthThreshold {
				c.nextCells[i][j] = true
			} else {
				c.nextCells[i][j] = c.CurrentCells[i][j]
			}
		}
	}

	c.CurrentCells, c.nextCells = c.nextCells, c.CurrentCells
}

// checkNeighborhood count alive Neighborhood around current cell
func (c *Cave) checkNeighborhood(row, col int) int {
	// row = 3 -> 2, 3, 4
	alive := 0
	for i := row - 1; i <= row+1; i++ {
		for j := col - 1; j <= col+1; j++ {
			if i == row && j == col {
				continue
			}

			if c.CurrentCells[i][j] == true {
				alive++
			}
		}
	}
	return alive
}

func (c *Cave) Cells() [][]bool {
	out := make([][]bool, c.Rows)
	for i := 0; i < c.Rows; i++ {
		out[i] = make([]bool, c.Cols)
		copy(out[i], c.CurrentCells[i+BorderWidth][BorderWidth:BorderWidth+c.Cols])
	}
	return out
}

func (c *Cave) Step() {
	c.cellularAutomatonStep()
}
