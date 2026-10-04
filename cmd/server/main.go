package main

import (
	"context"
	"log"
	"net/http"

	"github.com/CsToMy/weather-api/internal/api"
	"github.com/CsToMy/weather-api/internal/weather"
)

type stubProvider struct{}

func (stubProvider) Current(ctx context.Context, city string) (weather.Weather, error) {
	return weather.Weather{City: city, TempC: 20, Description: "stub data"}, nil
}

func main() {
	router := api.NewRouter(stubProvider{})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
