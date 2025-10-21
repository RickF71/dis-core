package main

import (
	disnet "dis-core/internal/net"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var netPort = flag.Int("net_port", 9090, "DIS-Network peer port")
var configPath = flag.String("config", "network.yaml", "network config file")

func main() {
	flag.Parse()

	addr := fmt.Sprintf(":%d", *netPort)
	log.Printf("🌐 Starting DIS-Network node on %s", addr)

	// Create a new network manager
	manager := disnet.NewManager(nil)

	// Start listening for peers
	go func() {
		if err := manager.Listen(*netPort); err != nil {
			log.Fatalf("❌ network listen error: %v", err)
		}
	}()

	// Optional: log peer count periodically
	go func() {
		for {
			peers := manager.ListPeers()
			log.Printf("🧭 Connected peers: %d", len(peers))
			time.Sleep(10 * time.Second)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("🛑 Shutdown signal received — closing DIS-Network.")
	manager.Close()
}
