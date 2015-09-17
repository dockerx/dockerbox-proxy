package main

import (
	"os"
	"log"
	"syscall"
	"os/signal"
	"github.com/dockerx/dockerbox-proxy/backend"
	"github.com/dockerx/dockerbox-proxy/proxy"
)

func main() {
	backend.Initialize()
	proxy.StartProxy()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Println("Shutting down proxy")
}
