package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CsToMy/weather-api/internal/weather"
)

type fakeProvider struct {
	result weather.Weather
	err    error
}

func (f fakeProvider) Current(ctx context.Context, city string) (weather.Weather, error) {
	return f.result, f.err
}

func TestWeatherHandlerStatus(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		provider   fakeProvider
		wantStatus int
	}{
		{"missing city", "/weather", fakeProvider{}, http.StatusBadRequest},
		{"unknown city", "/weather?city=Nowhere", fakeProvider{err: weather.ErrCityNotFound}, http.StatusNotFound},
		{"upstream failure", "/weather?city=Stuttgart", fakeProvider{err: errors.New("upstream error")}, http.StatusBadGateway},
		{"success", "/weather?city=Stuttgart", fakeProvider{result: weather.Weather{City: "Stuttgart"}}, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := NewRouter(tc.provider)
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestWeatherHandlerBody(t *testing.T) {
	want := weather.Weather{City: "Stuttgart", TempC: 17.4, Description: "cloudy"}
	router := NewRouter(fakeProvider{result: want})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/weather?city=Berlin", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var got weather.Weather

	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if got != want {
		t.Fatalf("body = %+v, want %+v", got, want)
	}
}
