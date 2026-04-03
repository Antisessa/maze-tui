package opener

import "maze/internal/domain"

type OpenMazeSelectedMsg struct {
	Path string
}

type OpenMazeCanceledMsg struct{}

type MazeLoadedMsg struct {
	Path  string
	Board domain.Board
}

type MazeLoadErrMsg struct {
	Path string
	Err  error
}
