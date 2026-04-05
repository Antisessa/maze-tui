package opener

import (
	tea "charm.land/bubbletea/v2"
	"maze/internal/models/shared"
)

func OpenMazeCmd(path string, storage shared.MazeStorage) tea.Cmd {
	return func() tea.Msg {
		board, err := storage.Open(path)
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
