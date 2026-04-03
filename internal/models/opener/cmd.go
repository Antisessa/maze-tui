package opener

import (
	tea "charm.land/bubbletea/v2"
	"maze/internal/service"
)

func OpenMazeCmd(path string) tea.Cmd {
	return func() tea.Msg {
		board, err := service.Open(path)
		if err != nil {
			return MazeLoadErrMsg{
				Path: path,
				Err:  err,
			}
		}

		return MazeLoadedMsg{
			Path:  path,
			Board: board,
		}
	}
}
