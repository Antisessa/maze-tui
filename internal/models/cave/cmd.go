package cave

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"maze/internal/domain"
	"maze/internal/models/shared"
)

type GenerateParams struct {
	Chance       float32
	Birth, Death int
	Rows, Cols   int
}

func OpenCaveCmd(storage shared.CaveStorage, path string) tea.Cmd {
	return func() tea.Msg {
		if storage.Open == nil {
			return ErrorMsg{Err: fmt.Errorf("storage.Open is nil")}
		}

		c, err := storage.Open(path)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("open cave %q: %w", path, err)}
		}

		return LoadedMsg{
			Cave: c,
			Path: path,
		}
	}
}

func GenerateAndSaveCmd(storage shared.CaveStorage, path string, p GenerateParams) tea.Cmd {
	return func() tea.Msg {
		if storage.Save == nil {
			return ErrorMsg{Err: fmt.Errorf("storage.Save is nil")}
		}

		c, err := domain.InitCave(p.Chance, p.Birth, p.Death, p.Rows, p.Cols)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("generate cave: %w", err)}
		}

		if err = storage.Save(path, c); err != nil {
			return ErrorMsg{Err: fmt.Errorf("save cave %q: %w", path, err)}
		}

		return GeneratedMsg{
			Cave: c,
			Path: path,
		}
	}
}
