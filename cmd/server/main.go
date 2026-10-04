package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CsToMy/weather-api/internal/api"
	"github.com/CsToMy/weather-api/internal/openweather"
)

const openWeatherBaseURL = "https://api.openweathermap.org"

func main() {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENWEATHER_API_KEY environment variable is required")
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}

	provider := openweather.NewClient(openWeatherBaseURL, apiKey, httpClient)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           api.NewRouter(provider),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on :8080")
	log.Fatal(srv.ListenAndServe())
}
