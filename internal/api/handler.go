package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/CsToMy/weather-api/internal/weather"
)

type Handler struct {
	provider weather.Provider
}

func NewRouter(provider weather.Provider) http.Handler {
	h := &Handler{provider: provider}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /weather", h.getWeather)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func (h *Handler) getWeather(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		writeError(w, http.StatusBadRequest, "city parameter is required")
		return
	}

	result, err := h.provider.Current(r.Context(), city)
	if err != nil {
		if errors.Is(err, weather.ErrCityNotFound) {
			writeError(w, http.StatusNotFound, "city not found")
			return
		}
		log.Printf("weather provider error: %v", err)
		writeError(w, http.StatusBadGateway, "weather service unavailable")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
