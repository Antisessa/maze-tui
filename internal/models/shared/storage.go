package shared

import "maze/internal/domain"

type CaveStorage interface {
	Open(path string) (domain.Cave, error)
	Save(path string, cave domain.Cave) error
}

type MazeStorage interface {
	Open(path string) (domain.Board, error)
	Save(path string, board domain.Board) error
}
