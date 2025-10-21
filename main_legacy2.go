// main.go — redirector entrypoint for DIS-Core
// Purpose: guide users toward the correct node launcher (cmd/dis-core)

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println(`🧭 DIS-CORE Repository Entrypoint
---------------------------------
You are at the root of the DIS-Core repo.

To start the node server:
	go run ./cmd/dis-core --dis_port=8080

To start a secondary node:
	go run ./cmd/dis-core --dis_port=6969

To use maintenance tools (receipts, freeze, etc.):
	go run ./cmd/main_legacy.go`)

	// Always start the server
	disPort := 8080
	fmt.Printf("🌐 Starting DIS node on port %d\n", disPort)
	cmd := exec.Command("go", "run", "./cmd/dis-core", fmt.Sprintf("--dis_port=%d", disPort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ server failed: %v\n", err)
		os.Exit(1)
	}

	// Print suggested commands for users
	fmt.Println("\nTo start the DIS node server:")
	fmt.Println("  go run ./cmd/dis-core --dis_port=8080")
	fmt.Println("  curl http://localhost:8080/api/status")
}
