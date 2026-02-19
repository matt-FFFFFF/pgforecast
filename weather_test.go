package pgforecast

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestParseOpenMeteoJSON(t *testing.T) {
	// Minimal valid Open-Meteo response with 2 hours
	raw := `{
		"hourly": {
			"time": ["2026-02-19T08:00", "2026-02-19T09:00"],
			"temperature_2m": [10.0, 11.0],
			"relative_humidity_2m": [80.0, 75.0],
			"dew_point_2m": [7.0, 7.0],
			"wind_speed_10m": [12.0, 14.0],
			"wind_direction_10m": [225.0, 230.0],
			"wind_gusts_10m": [20.0, 22.0],
			"cloud_cover": [50.0, 60.0],
			"cloud_cover_low": [30.0, 40.0],
			"cloud_cover_mid": [10.0, 10.0],
			"cloud_cover_high": [10.0, 10.0],
			"cape": [100.0, 200.0],
			"shortwave_radiation": [300.0, 400.0],
			"precipitation": [0.0, 0.0],
			"precipitation_probability": [10.0, 20.0],
			"freezing_level_height": [2000.0, 2100.0],
			"is_day": [1, 1],
			"weather_code": [2, 3],
			"pressure_msl": [1013.0, 1012.0],
			"visibility": [10000.0, 10000.0],
			"wind_speed_1000hPa": [13.0, 15.0],
			"wind_direction_1000hPa": [225.0, 230.0],
			"temperature_1000hPa": [9.0, 10.0],
			"geopotential_height_1000hPa": [100.0, 100.0],
			"wind_speed_950hPa": [16.0, 18.0],
			"wind_direction_950hPa": [230.0, 235.0],
			"temperature_950hPa": [7.0, 8.0],
			"geopotential_height_950hPa": [500.0, 500.0],
			"wind_speed_925hPa": [18.0, 20.0],
			"wind_direction_925hPa": [235.0, 240.0],
			"temperature_925hPa": [5.0, 6.0],
			"geopotential_height_925hPa": [750.0, 750.0],
			"wind_speed_900hPa": [20.0, 22.0],
			"wind_direction_900hPa": [240.0, 245.0],
			"temperature_900hPa": [3.0, 4.0],
			"geopotential_height_900hPa": [1000.0, 1000.0],
			"wind_speed_850hPa": [22.0, 25.0],
			"wind_direction_850hPa": [245.0, 250.0],
			"temperature_850hPa": [0.0, 1.0],
			"geopotential_height_850hPa": [1500.0, 1500.0],
			"wind_speed_700hPa": [40.0, 45.0],
			"wind_direction_700hPa": [260.0, 265.0],
			"temperature_700hPa": [-10.0, -9.0],
			"geopotential_height_700hPa": [3000.0, 3000.0]
		}
	}`

	data, err := ParseOpenMeteoJSON([]byte(raw))
	if err != nil {
		t.Fatalf("ParseOpenMeteoJSON: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("got %d hours, want 2", len(data))
	}

	h := data[0]
	if h.Temperature != 10.0 {
		t.Errorf("temp = %v, want 10.0", h.Temperature)
	}
	if h.WindSpeed != 12.0 {
		t.Errorf("wind = %v, want 12.0", h.WindSpeed)
	}
	if h.CAPE != 100.0 {
		t.Errorf("CAPE = %v, want 100.0", h.CAPE)
	}
	if len(h.PressureLevels) != 6 {
		t.Errorf("pressure levels = %d, want 6", len(h.PressureLevels))
	}

	// Check pressure level ordering
	found850 := false
	for _, pl := range h.PressureLevels {
		if pl.Pressure == 850 {
			found850 = true
			if pl.WindSpeed != 22.0 {
				t.Errorf("850hPa wind = %v, want 22.0", pl.WindSpeed)
			}
		}
	}
	if !found850 {
		t.Error("850hPa level not found")
	}
}

func TestParseOpenMeteoJSON_Invalid(t *testing.T) {
	_, err := ParseOpenMeteoJSON([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHourlyMetricsJSON(t *testing.T) {
	// Ensure HourlyMetrics serialises correctly for WASM consumers
	m := HourlyMetrics{
		WindSpeed:       14,
		WindDirStr:      "SW",
		FlyabilityScore: 4,
		XCPotential:     "Low",
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]interface{}
	json.Unmarshal(b, &decoded)

	if decoded["wind_dir_str"] != "SW" {
		t.Errorf("wind_dir_str = %v", decoded["wind_dir_str"])
	}
	if decoded["flyability_score"].(float64) != 4 {
		t.Errorf("flyability_score = %v", decoded["flyability_score"])
	}
}

func TestHTTPClientHasTimeout(t *testing.T) {
	if httpClient.Timeout != 30*time.Second {
		t.Errorf("httpClient.Timeout = %v, want 30s", httpClient.Timeout)
	}
}

func TestFetchWeatherWithContext_Cancellation(t *testing.T) {
	// Use a custom RoundTripper that blocks until context is cancelled
	origClient := httpClient
	httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		}),
	}
	defer func() { httpClient = origClient }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := FetchWeatherWithContext(ctx, Site{Lat: 0, Lon: 0}, ForecastOptions{})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchWeatherWithContext_NilContext(t *testing.T) {
	// Ensure nil context doesn't panic - just verify it doesn't crash
	origClient := httpClient
	httpClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("expected")
		}),
	}
	defer func() { httpClient = origClient }()

	_, err := FetchWeatherWithContext(nil, Site{}, ForecastOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
}
