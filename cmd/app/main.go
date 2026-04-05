package main

import (
	tea "charm.land/bubbletea/v2"
	"log"
	"maze/internal/models"
	"maze/internal/storage"
	"os"
)

func main() {
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("get current dir: %v", err)
	}

	mazeStorage := &storage.MazeStorage{}
	caveStorage := &storage.CaveStorage{}
	m := models.InitModel(curDir, mazeStorage, caveStorage)

	p := tea.NewProgram(m)

	if _, err = p.Run(); err != nil {
		log.Fatalf("program error: %v", err)
	}
}
