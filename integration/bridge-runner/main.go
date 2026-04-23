package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-i2p/go-sam-bridge/lib/embedding"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	listenAddr := envOrDefault("SAM_BRIDGE_ADDR", "127.0.0.1:7656")
	i2cpAddr := envOrDefault("I2CP_ADDR", "127.0.0.1:7654")

	bridge, err := embedding.New(
		embedding.WithListenAddr(listenAddr),
		embedding.WithI2CPAddr(i2cpAddr),
		embedding.WithDatagramPort(7655),
	)
	if err != nil {
		log.Fatalf("create bridge: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := bridge.Start(ctx); err != nil {
		log.Fatalf("start bridge: %v", err)
	}

	log.Printf("embedded SAM bridge started on %s", listenAddr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer stopCancel()
	if err := bridge.Stop(stopCtx); err != nil {
		log.Printf("stop bridge: %v", err)
	}
}
