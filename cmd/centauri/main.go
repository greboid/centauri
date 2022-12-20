package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csmith/centauri/config"
	"github.com/csmith/centauri/proxy"
	"github.com/csmith/envflag"
)

var (
	configPath       = flag.String("config", "centauri.conf", "Path to config")
	selectedFrontend = flag.String("frontend", "tcp", "Frontend to listen on")
)

var proxyManager *proxy.Manager

func main() {
	envflag.Parse()

	providers, err := certProviders()
	if err != nil {
		log.Fatalf("Error creating certificate providers: %v", err)
	}

	defaultProvider := "lego"
	if _, ok := providers[defaultProvider]; !ok {
		defaultProvider = "selfsigned"
	}

	proxyManager = proxy.NewManager(providers, defaultProvider)
	rewriter := proxy.NewRewriter(proxyManager)
	updateRoutes()
	listenForHup()
	monitorCerts()

	f, ok := frontends[*selectedFrontend]
	if !ok {
		log.Fatalf("Invalid frontend specified: %s", *selectedFrontend)
	}

	err = f.Serve(proxyManager, rewriter)
	if err != nil {
		log.Fatalf("Failed to start frontend: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	// Wait for a signal
	log.Printf("Received signal %s, stopping frontend...", <-c)

	f.Stop(context.Background())

	log.Printf("Frontend stopped. Goodbye!")
}

func monitorCerts() {
	go func() {
		for {
			time.Sleep(12 * time.Hour)
			log.Printf("Checking for certificate validity...")
			if err := proxyManager.CheckCertificates(); err != nil {
				log.Fatalf("Error performing periodic check of certificates: %v", err)
			}
		}
	}()
}

func listenForHup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			<-c
			log.Printf("Received SIGHUP, updating routes")
			updateRoutes()
		}
	}()
}

func updateRoutes() {
	log.Printf("Reading config file %s", *configPath)

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	routes, err := config.Parse(configFile)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	log.Printf("Installing %d routes", len(routes))
	if err := proxyManager.SetRoutes(routes); err != nil {
		log.Fatalf("Route manager error: %v", err)
	}

	log.Printf("Finished installing %d routes", len(routes))
}
