package cave

import "maze/internal/domain"

type LoadedMsg struct {
	Cave domain.Cave
	Path string
}

type GeneratedMsg struct {
	Cave domain.Cave
	Path string
}

type ErrorMsg struct {
	Err error
}

type stepTickMsg struct{}
type CancelMsg struct{}
