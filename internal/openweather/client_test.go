package openweather

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CsToMy/weather-api/internal/weather"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL, "test-key", srv.Client())
}

func TestCurrentSuccess(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("q") != "Stuttgart" || q.Get("appid") != "test-key" || q.Get("units") != "metric" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"weather":[{"description":"overcast clouds"}],"main":{"temp":12.5},"name":"Stuttgart"}`))
	})

	got, err := c.Current(context.Background(), "Stuttgart")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := weather.Weather{City: "Stuttgart", TempC: 17.4, Description: "overcast clouds"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestCurrentCityNotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"cod":"404","message":"city not found"}`))
	})

	_, err := c.Current(context.Background(), "Nowhere")
	if !errors.Is(err, weather.ErrCityNotFound) {
		t.Fatalf("error = %v, want ErrCityNotFound", err)
	}
}

func TestCurrentUpstreamError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := c.Current(context.Background(), "Berlin")
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, weather.ErrCityNotFound) {
		t.Fatal("a 500 must not be reported as city not found")
	}
}

func TestCurrentInvalidJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("this is not json"))
	})

	if _, err := c.Current(context.Background(), "Berlin"); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestCurrentTimeout(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Current(ctx, "Berlin")
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want context deadline exceeded", err)
	}
	if strings.Contains(err.Error(), "test-key") {
		t.Fatalf("error leaks the API key: %v", err)
	}
}
