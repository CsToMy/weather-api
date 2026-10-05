package openweather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/CsToMy/weather-api/internal/weather"
)

var _ weather.Provider = (*Client)(nil)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, httpClient: httpClient}
}

type apiResponse struct {
	Name string `json:"name"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

func (c *Client) Current(ctx context.Context, city string) (weather.Weather, error) {
	params := url.Values{}
	params.Set("q", city)
	params.Set("appid", c.apiKey)
	params.Set("units", "metric")

	endpoint := c.baseURL + "/data/2.5/weather?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return weather.Weather{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		return weather.Weather{}, fmt.Errorf("call openweather: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return weather.Weather{}, weather.ErrCityNotFound
	case resp.StatusCode != http.StatusOK:
		return weather.Weather{}, fmt.Errorf("openweather returned status %d", resp.StatusCode)
	}

	var body apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return weather.Weather{}, fmt.Errorf("decode response: %w", err)
	}

	description := ""
	if len(body.Weather) > 0 {
		description = body.Weather[0].Description
	}

	return weather.Weather{
		City:        body.Name,
		TempC:       body.Main.Temp,
		Description: description,
	}, nil
}
