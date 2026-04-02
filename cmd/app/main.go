package main

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"log"
	"maze/internal/models"
)

func main() {
	m := models.InitModel()

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		log.Fatal(fmt.Printf("program error: %v", err))
	}
}
