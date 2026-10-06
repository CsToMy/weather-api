package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CsToMy/weather-api/internal/api"
	"github.com/CsToMy/weather-api/internal/openweather"
	"github.com/CsToMy/weather-api/internal/server"
)

const (
	openWeatherBaseURL = "https://api.openweathermap.org"
	listenAddr         = ":8080"
	shutdownTimeout    = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENWEATHER_API_KEY environment variable is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{Timeout: 5 * time.Second}

	provider := openweather.NewClient(openWeatherBaseURL, apiKey, httpClient)

	srv := &http.Server{
		Handler:           api.NewRouter(provider),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	log.Println("listening on", listenAddr)
	if err := server.Run(ctx, srv, ln, shutdownTimeout); err != nil {
		return err
	}

	log.Fatal("server stopped")
	return nil
}
