package generator

import (
	tea "charm.land/bubbletea/v2"
	"maze/internal/service"
)

func GenerateMazeCmd(path string, width, height int) tea.Cmd {
	return func() tea.Msg {
		board, err := service.GenerateMaze(height, width)
		if err != nil {
			return MazeGenerateErrMsg{
				Path: path,
				Err:  err,
			}
		}

		if err = service.SaveMaze(path, board); err != nil {
			return MazeGenerateErrMsg{
				Path: path,
				Err:  err,
			}
		}

		return MazeGeneratedMsg{
			Path: path,
		}
	}
}
