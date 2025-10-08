package main

import (
	"dis-core/internal/console"
	"log"
)

func main() {
	seats := []string{"uid-terracouncil-001", "uid-terracouncil-002"}
	ac := console.NewConsole("domain.terra", "DIS-CORE v1.0", seats)

	_, err := ac.LogAction("domain.freeze.v1", "policy.freeze.rego", "uid-terracouncil-001")
	if err != nil {
		log.Fatal("❌ Failed:", err)
	}

	if err := ac.SaveState(); err != nil {
		log.Fatal("❌ Could not save console state:", err)
	}

	log.Println("✅ Action & state saved with REAL Ed25519 signature. Keys in versions/v0.6/keys/")
}
