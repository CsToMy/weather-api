package weather

import (
	"context"
	"errors"
)

var ErrCityNotFound = errors.New("City not found.")

type Weather struct {
	City        string  `json:"city"`
	TempC       float64 `json:"temp_c"`
	Description string  `json:"description"`
}

type Provider interface {
	Current(ctx context.Context, city string) (Weather, error)
}
