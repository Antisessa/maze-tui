package service

import (
	"fmt"
	"math/rand"
	"maze/internal/domain"
)

type ellerCell struct {
	cell domain.Cell
	set  int
}

type ellerSet struct {
	countCell       int
	countBottomWall int
}

func GenerateMaze(height, width int) (domain.Board, error) {
	if height <= 0 || height > 50 {
		return domain.Board{}, fmt.Errorf("invalid maze height")
	} else if width <= 0 || width > 50 {
		return domain.Board{}, fmt.Errorf("invalid maze width")
	}

	// инициализация матрицы
	cells := make([][]ellerCell, height)

	// Цикл по строкам
	for i := range cells {
		uSet := make(map[int]ellerSet, width)
		if i == 0 { // Инициализируем первую строку назначая каждой клетке свое уникальное множество
			cells[i] = make([]ellerCell, width)
		} else { // Для последующих - копируем предыдущую строку
			cells[i] = make([]ellerCell, width)
			copy(cells[i], cells[i-1])
			// Готовим строку по правилам
			for j := range cells[i] {
				eSet := uSet[cells[i][j].set]
				cells[i][j].cell.RightWall = false // Сносим все правые стены
				if cells[i][j].cell.BottomWall {
					cells[i][j].set = 0 // Убираем множество у клетки если у нее есть нижняя стена
				} else {
					eSet.countCell++
					uSet[cells[i][j].set] = eSet
				}
				cells[i][j].cell.BottomWall = false // Сносим все нижние стены
			}
		}

		row := cells[i]
		setUnique(row, uSet)

		// Цикл по клеткам
		for j := range row {
			// Строительство правых стен
			if j == len(row)-1 { // У самой правой клетки должна быть правая стена
				row[j].cell.RightWall = true
			} else if row[j].set == row[j+1].set { // У соседних клеток с одинаковым множеством должна быть стена
				row[j].cell.RightWall = true
			} else if rand.Intn(2) == 1 { // Либо ставим стену между клетками
				row[j].cell.RightWall = true
			} else { // Либо объединяем их A & B в одно множество A
				aSetIdx := row[j].set
				bSetIdx := row[j+1].set
				aSet := uSet[aSetIdx]
				bSet := uSet[bSetIdx]
				for k := 0; k < len(row); k++ {
					if row[k].set == bSetIdx {
						row[k].set = aSetIdx
						aSet.countCell++
						bSet.countCell--
					}
				}
				uSet[aSetIdx] = aSet
				uSet[bSetIdx] = bSet
			}

			// Строительство нижних стен
			if i == len(cells)-1 && j == len(row)-1 { // Для самой нижней правой просто ставим стену
				row[j].cell.BottomWall = true
				break
			} else if i == len(cells)-1 { // Для всех нижних клеток должна быть нижняя стена
				row[j].cell.BottomWall = true
				if row[j].set != row[j+1].set { // Если множества соседних клеток разные
					row[j].cell.RightWall = false // Убираем стену между ними
					bSetIdx := row[j+1].set
					for k := 0; k < len(row); k++ { // Объединяем множества
						if row[k].set == bSetIdx {
							row[k].set = row[j].set
						}
					}
				}
			} else { // Для всех других выбираем случайно, но нижних стен в множестве должно быть меньше кол-ва клеток
				eSet := uSet[row[j].set]
				if rand.Intn(2) == 1 && eSet.countCell > eSet.countBottomWall+1 {
					row[j].cell.BottomWall = true
					eSet.countBottomWall++
				}
				uSet[row[j].set] = eSet
			}
		}
		// Конец цикла по клеткам

	}
	// Конец цикла по строкам

	dCells := make([][]domain.Cell, height)
	for i := range dCells {
		dCells[i] = make([]domain.Cell, width)
		for j := range dCells[i] {
			dCells[i][j] = cells[i][j].cell
		}
	}

	return domain.Board{
		Height: height,
		Width:  width,
		Cells:  dCells,
	}, nil
}

// setUnique присваивает клеткам уникальное множество исходя из уже имеющихся
func setUnique(row []ellerCell, set map[int]ellerSet) {
	setIdx := 1
	for i := 0; i < len(row); i++ {
		if row[i].set != 0 {
			continue
		}

		// ищем первый свободный номер множества
		for {
			eSet := set[setIdx]
			if set[setIdx].countCell == 0 {
				eSet.countCell++
				set[setIdx] = eSet
				break
			}
			setIdx++
		}

		row[i].set = setIdx
		setIdx++
	}
}
