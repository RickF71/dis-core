package app

import (
	"dis-core/internal/canon"
	"log"
)

func BootstrapCanon(reg *canon.Registry) {
	log.Println("Seeding Terra + USA themes...")
	canon.BootstrapThemes(reg)
	log.Println("Seeding complete.")
}
