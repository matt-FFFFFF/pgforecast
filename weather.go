package pgforecast

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var pressureLevels = []int{1000, 950, 925, 900, 850, 700}

var surfaceParams = []string{
	"temperature_2m", "relative_humidity_2m", "dew_point_2m",
	"wind_speed_10m", "wind_direction_10m", "wind_gusts_10m",
	"cloud_cover", "cloud_cover_low", "cloud_cover_mid", "cloud_cover_high",
	"cape", "shortwave_radiation", "precipitation", "precipitation_probability",
	"freezing_level_height", "is_day", "weather_code", "pressure_msl", "visibility",
}

func windSpeedUnit(units string) string {
	switch units {
	case "kph", "kmh":
		return "kmh"
	case "knots", "kn":
		return "kn"
	case "ms":
		return "ms"
	default:
		return "mph"
	}
}

// FetchWeather fetches weather data from Open-Meteo for a site.
func FetchWeather(site Site, opts ForecastOptions) ([]HourlyData, error) {
	u, _ := url.Parse("https://api.open-meteo.com/v1/forecast")
	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.4f", site.Lat))
	q.Set("longitude", fmt.Sprintf("%.4f", site.Lon))
	q.Set("hourly", buildHourlyParams())
	q.Set("wind_speed_unit", windSpeedUnit(opts.Units))
	q.Set("forecast_days", "16")
	q.Set("timezone", "UTC")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("fetching weather: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return parseHourlyData(raw)
}

// ParseOpenMeteoJSON parses a raw Open-Meteo JSON response into HourlyData.
// Exported for use by WASM and other consumers.
func ParseOpenMeteoJSON(rawJSON []byte) ([]HourlyData, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(rawJSON, &raw); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}
	return parseHourlyData(raw)
}

func buildHourlyParams() string {
	params := make([]string, len(surfaceParams))
	copy(params, surfaceParams)
	for _, p := range pressureLevels {
		params = append(params,
			fmt.Sprintf("wind_speed_%dhPa", p),
			fmt.Sprintf("wind_direction_%dhPa", p),
			fmt.Sprintf("temperature_%dhPa", p),
			fmt.Sprintf("geopotential_height_%dhPa", p),
		)
	}
	return strings.Join(params, ",")
}

func parseHourlyData(raw map[string]interface{}) ([]HourlyData, error) {
	hourly, ok := raw["hourly"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no hourly data in response")
	}

	times, ok := hourly["time"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no time array")
	}

	result := make([]HourlyData, len(times))
	for i, t := range times {
		ts, _ := time.Parse("2006-01-02T15:04", t.(string))
		result[i].Time = ts
	}

	getFloat := func(key string, i int) float64 {
		arr, ok := hourly[key].([]interface{})
		if !ok || i >= len(arr) { return 0 }
		if arr[i] == nil { return 0 }
		v, ok := arr[i].(float64)
		if !ok { return 0 }
		return v
	}

	for i := range result {
		d := &result[i]
		d.Temperature = getFloat("temperature_2m", i)
		d.RelativeHumidity = getFloat("relative_humidity_2m", i)
		d.DewPoint = getFloat("dew_point_2m", i)
		d.WindSpeed = getFloat("wind_speed_10m", i)
		d.WindDirection = getFloat("wind_direction_10m", i)
		d.WindGusts = getFloat("wind_gusts_10m", i)
		d.CloudCover = getFloat("cloud_cover", i)
		d.CloudCoverLow = getFloat("cloud_cover_low", i)
		d.CloudCoverMid = getFloat("cloud_cover_mid", i)
		d.CloudCoverHigh = getFloat("cloud_cover_high", i)
		d.CAPE = getFloat("cape", i)
		d.ShortwaveRadiation = getFloat("shortwave_radiation", i)
		d.Precipitation = getFloat("precipitation", i)
		d.PrecipitationProbability = getFloat("precipitation_probability", i)
		d.FreezingLevelHeight = getFloat("freezing_level_height", i)
		d.IsDay = int(getFloat("is_day", i))
		d.WeatherCode = int(getFloat("weather_code", i))
		d.PressureMSL = getFloat("pressure_msl", i)
		d.Visibility = getFloat("visibility", i)

		for _, p := range pressureLevels {
			pl := PressureLevel{
				Pressure:          p,
				WindSpeed:         getFloat(fmt.Sprintf("wind_speed_%dhPa", p), i),
				WindDirection:     getFloat(fmt.Sprintf("wind_direction_%dhPa", p), i),
				Temperature:       getFloat(fmt.Sprintf("temperature_%dhPa", p), i),
				GeopotentialHeight: getFloat(fmt.Sprintf("geopotential_height_%dhPa", p), i),
			}
			d.PressureLevels = append(d.PressureLevels, pl)
		}
	}

	return result, nil
}
