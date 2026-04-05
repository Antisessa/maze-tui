package service

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
	currentCells                   [][]bool
	nextCells                      [][]bool
	rows, cols                     int
}

//func LoadCave(path string) Cave {
//
//}

func InitCave(chance float32, birth, death, rows, cols int) (Cave, error) {
	if chance > 1.0 || chance <= 0 {
		return Cave{}, fmt.Errorf("cell initialization chance should be in range (0; 1]")
	} else if birth < 0 || birth > 7 {
		return Cave{}, fmt.Errorf("birth threshold should be in range [0; 7]")
	} else if death < 0 || death > 7 {
		return Cave{}, fmt.Errorf("death threshold should be in range [0; 7]")
	} else if rows < 0 || rows > 50 {
		return Cave{}, fmt.Errorf("cave's rows should be in range [1; 50]")
	} else if cols < 0 || cols > 50 {
		return Cave{}, fmt.Errorf("cave's cols should be in range [1; 50]")
	}

	newCave := Cave{
		initChance:     chance,
		birthThreshold: birth,
		deathThreshold: death,
		rows:           rows,
		cols:           cols,
	}

	newCave.generateCave()

	return newCave, nil
}

// generateCave инициализирует пещеру, обрамляя матрицу клеток рамкой "живой" границы
func (c *Cave) generateCave() {
	c.currentCells = make([][]bool, c.rows+BorderWidth*2)
	c.nextCells = make([][]bool, c.rows+BorderWidth*2)
	for i := range c.currentCells {
		c.currentCells[i] = make([]bool, c.cols+BorderWidth*2)
		c.nextCells[i] = make([]bool, c.cols+BorderWidth*2)
		for j := range c.currentCells[i] {
			isBorder := i < BorderWidth ||
				i >= len(c.currentCells)-BorderWidth ||
				j < BorderWidth ||
				j >= len(c.currentCells[i])-BorderWidth

			if isBorder {
				c.currentCells[i][j] = true
			} else {
				c.currentCells[i][j] = rand.Float32() < c.initChance
			}
		}
	}
}

func (c *Cave) cellularAutomatonStep() {
	for i := range c.currentCells {
		for j := range c.currentCells[i] {
			isBorder := i < BorderWidth ||
				i >= len(c.currentCells)-BorderWidth ||
				j < BorderWidth ||
				j >= len(c.currentCells[i])-BorderWidth

			if isBorder { // для клеток границы - отдельная проверка
				c.nextCells[i][j] = true
				continue
			}
			alive := c.checkNeighborhood(i, j)

			if c.currentCells[i][j] == true && alive < c.deathThreshold {
				c.nextCells[i][j] = false
			} else if c.currentCells[i][j] == false && alive > c.birthThreshold {
				c.nextCells[i][j] = true
			} else {
				c.nextCells[i][j] = c.currentCells[i][j]
			}
		}
	}

	c.currentCells, c.nextCells = c.nextCells, c.currentCells
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

			if c.currentCells[i][j] == true {
				alive++
			}
		}
	}
	return alive
}
