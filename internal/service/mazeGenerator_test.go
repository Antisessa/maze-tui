package service

import (
	"maze/internal/domain"
	"testing"
)

func TestGenerateMaze_InvalidSize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		height int
		width  int
	}{
		{name: "zero height", height: 0, width: 5},
		{name: "zero width", height: 5, width: 0},
		{name: "negative height", height: -1, width: 5},
		{name: "negative width", height: 5, width: -1},
		{name: "height too large", height: 51, width: 5},
		{name: "width too large", height: 5, width: 51},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := GenerateMaze(tc.height, tc.width)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestGenerateMaze_HasClosedOuterBorders(t *testing.T) {
	height, width := 10, 10
	board, err := GenerateMaze(height, width)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if height != board.Height {
		t.Fatalf("invalid maze height want: %d, got: %d", height, board.Height)
	} else if width != board.Width {
		t.Fatalf("invalid maze width want: %d, got: %d", width, board.Width)
	} else if len(board.Cells) != height {
		t.Fatalf("invalid maze rows len want: %d, got: %d", height, len(board.Cells))
	}

	for i := 0; i < height; i++ {
		if len(board.Cells[i]) != width {
			t.Fatalf("invalid maze width of row %d want: %d, got: %d", i, width, len(board.Cells[i]))
		}
		if board.Cells[i][width-1].RightWall != true {
			t.Fatalf("right border must be closed at row %d", i)
		}
	}

	for j := 0; j < width; j++ {
		if board.Cells[height-1][j].BottomWall != true {
			t.Fatalf("bottom border must be closed at col %d", j)
		}
	}
}

func TestGenerateMaze_IsConnectedAndAcyclic(t *testing.T) {
	height, width := 10, 10
	board, err := GenerateMaze(height, width)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitedCount := countReachableCells(board)
	if visitedCount != height*width {
		t.Fatal("maze must be fully connected")
	}

	edges := countPassages(board)
	if edges != height*width-1 {
		t.Fatal("perfect maze must have V-1 passages")
	}
}

func countReachableCells(board domain.Board) int {
	type point struct {
		r int
		c int
	}

	h, w := board.Height, board.Width
	visited := make([][]bool, h)
	for i := range visited {
		visited[i] = make([]bool, w)
	}

	queue := []point{{0, 0}}
	visited[0][0] = true
	count := 1

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		r, c := cur.r, cur.c

		// right
		if c+1 < w && !board.Cells[r][c].RightWall && !visited[r][c+1] {
			visited[r][c+1] = true
			queue = append(queue, point{r, c + 1})
			count++
		}

		// left
		if c-1 >= 0 && !board.Cells[r][c-1].RightWall && !visited[r][c-1] {
			visited[r][c-1] = true
			queue = append(queue, point{r, c - 1})
			count++
		}

		// down
		if r+1 < h && !board.Cells[r][c].BottomWall && !visited[r+1][c] {
			visited[r+1][c] = true
			queue = append(queue, point{r + 1, c})
			count++
		}

		// up
		if r-1 >= 0 && !board.Cells[r-1][c].BottomWall && !visited[r-1][c] {
			visited[r-1][c] = true
			queue = append(queue, point{r - 1, c})
			count++
		}
	}

	return count
}

func countPassages(board domain.Board) int {
	h, w := board.Height, board.Width
	edges := 0

	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			if c+1 < w && !board.Cells[r][c].RightWall {
				edges++
			}
			if r+1 < h && !board.Cells[r][c].BottomWall {
				edges++
			}
		}
	}

	return edges
}
