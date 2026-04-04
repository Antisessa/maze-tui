package main

import (
	tea "charm.land/bubbletea/v2"
	"log"
	"maze/internal/models"
	"os"
)

func main() {
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("get current dir: %v", err)
	}
	m := models.InitModel(curDir)

	p := tea.NewProgram(m)

	if _, err = p.Run(); err != nil {
		log.Fatalf("program error: %v", err)
	}
}
