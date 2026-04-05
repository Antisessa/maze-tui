package domain

import (
	"reflect"
	"testing"
)

func TestInitCave_InvalidParams(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		chance float32
		birth  int
		death  int
		rows   int
		cols   int
	}{
		{name: "invalid init chance", chance: -0.5, birth: 1, death: 1, rows: 10, cols: 10},
		{name: "invalid birth rate", chance: 0.5, birth: 8, death: 1, rows: 10, cols: 10},
		{name: "invalid death rate", chance: 0.5, birth: 1, death: 8, rows: 10, cols: 10},
		{name: "invalid Rows", chance: 0.5, birth: 1, death: 1, rows: 51, cols: 10},
		{name: "invalid Cols", chance: 0.5, birth: 1, death: 1, rows: 10, cols: 51},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := InitCave(tc.chance, tc.birth, tc.death, tc.rows, tc.cols)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestInitCave_CheckBorder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		chance float32
		birth  int
		death  int
		rows   int
		cols   int
	}{
		{name: "square cave", chance: 0.5, birth: 1, death: 1, rows: 5, cols: 5},
		{name: "rectangle (Rows) cave", chance: 0.5, birth: 1, death: 1, rows: 8, cols: 5},
		{name: "rectangle (Cols) cave", chance: 0.5, birth: 1, death: 1, rows: 5, cols: 8},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cave, err := InitCave(tc.chance, tc.birth, tc.death, tc.rows, tc.cols)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			rowsLength := len(cave.CurrentCells)
			if rowsLength != tc.rows+BorderWidth*2 {
				t.Fatalf("cell's Rows length mismatch - want:%d, got:%d", tc.rows+BorderWidth*2, rowsLength)
			}

			for i := range cave.CurrentCells {
				colsLength := len(cave.CurrentCells[i])
				if colsLength != tc.cols+BorderWidth*2 {
					t.Fatalf("Cols:%d length mismatch - want:%d, got:%d", i, tc.cols+BorderWidth*2, colsLength)
				}
			}

			for i := range cave.CurrentCells {
				for j := range cave.CurrentCells[i] {
					isBorder := i < BorderWidth ||
						i >= len(cave.CurrentCells)-BorderWidth ||
						j < BorderWidth ||
						j >= len(cave.CurrentCells[i])-BorderWidth

					if isBorder && !cave.CurrentCells[i][j] {
						t.Fatalf("border cell i=%d, j=%d should be alive", i, j)
					}
				}
			}
		})
	}
}

func TestCellularAutomatonStep_LiveCellDies(t *testing.T) {
	t.Parallel()

	c := newTestCave(
		grid(
			"#######",
			"#.....#",
			"#.....#",
			"#..#..#",
			"#.....#",
			"#.....#",
			"#######",
		),
		3,
		1,
	)

	c.cellularAutomatonStep()

	if c.CurrentCells[3][3] {
		t.Fatalf("center cell should die")
	}
}

func TestCellularAutomatonStep_DeadCellBecomesAlive(t *testing.T) {
	t.Parallel()

	c := newTestCave(
		grid(
			"#####",
			"#...#",
			"#...#",
			"#...#",
			"#####",
		),
		2, // если соседей > 2, мертвая клетка оживает
		1,
	)

	c.cellularAutomatonStep()

	want := grid(
		"#####",
		"#####",
		"##.##",
		"#####",
		"#####",
	)

	if !reflect.DeepEqual(c.CurrentCells, want) {
		t.Fatalf("unexpected state after step\nwant: %#v\ngot:  %#v", want, c.CurrentCells)
	}
}

func TestCellularAutomatonStep_LiveCellStaysAlive(t *testing.T) {
	t.Parallel()

	c := newTestCave(
		grid(
			"#####",
			"##..#",
			"#.#.#",
			"#..##",
			"#####",
		),
		8, // чтобы новые клетки точно не рождались
		2, // живая клетка умирает только если соседей < 2
	)

	c.cellularAutomatonStep()

	if !c.CurrentCells[2][2] {
		t.Fatalf("center cell should stay alive")
	}
}

func TestCellularAutomatonStep_MatrixShouldChange(t *testing.T) {
	t.Parallel()

	c := newTestCave(
		grid(
			"############",
			"####..#..###",
			"#...###...##",
			"#.........##",
			"#.......#..#",
			"#..###..####",
			"#..###..####",
			"##..#...#..#",
			"##..##...#.#",
			"##.###.#####",
			"####.#.....#",
			"#.#......#.#",
			"############",
		),
		4, // чтобы новые клетки точно не рождались
		4, // живая клетка умирает только если соседей < 2
	)

	dWant := grid(
		"############",
		"####..#..###",
		"#...###...##",
		"#.........##",
		"#.......#..#",
		"#..###..####",
		"#..###..####",
		"##..#...#..#",
		"##..##...#.#",
		"##.###.#####",
		"####.#.....#",
		"#.#......#.#",
		"############",
	)

	c.cellularAutomatonStep()

	if reflect.DeepEqual(c.CurrentCells, dWant) {
		t.Fatalf("unexpected state after step\nwant: %#v\ngot:  %#v", dWant, c.CurrentCells)
	}
}

func grid(rows ...string) [][]bool {
	res := make([][]bool, len(rows))
	for i, row := range rows {
		res[i] = make([]bool, len(row))
		for j, ch := range row {
			res[i][j] = ch == '#'
		}
	}
	return res
}

func cloneGrid(src [][]bool) [][]bool {
	dst := make([][]bool, len(src))
	for i := range src {
		dst[i] = make([]bool, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

func newTestCave(cur [][]bool, birth, death int) *Cave {
	next := make([][]bool, len(cur))
	for i := range cur {
		next[i] = make([]bool, len(cur[i]))
	}
	return &Cave{
		CurrentCells:   cloneGrid(cur),
		nextCells:      next,
		birthThreshold: birth,
		deathThreshold: death,
	}
}
