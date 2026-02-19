//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/matt-FFFFFF/pgforecast"
)

func main() {
	js.Global().Set("pgforecastWasm", js.ValueOf(map[string]interface{}{
		"computeMetrics":  js.FuncOf(computeMetrics),
		"defaultTuning":   js.FuncOf(defaultTuning),
		"degreesToCompass": js.FuncOf(degreesToCompass),
	}))

	// Keep alive
	select {}
}

// computeMetrics takes JSON weather data + site config, returns computed metrics JSON.
// Called from JS: pgforecastWasm.computeMetrics(weatherJSON, siteJSON, tuningJSON)
func computeMetrics(_ js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return jsError("need at least 2 args: weatherJSON, siteJSON")
	}

	weatherJSON := args[0].String()
	siteJSON := args[1].String()

	var tuningJSON string
	if len(args) >= 3 && !args[2].IsUndefined() && !args[2].IsNull() {
		tuningJSON = args[2].String()
	}

	// Parse site
	var site pgforecast.Site
	if err := json.Unmarshal([]byte(siteJSON), &site); err != nil {
		return jsError("parsing site: " + err.Error())
	}

	// Parse tuning or use defaults
	tc := pgforecast.DefaultTuningConfig()
	if tuningJSON != "" {
		if err := json.Unmarshal([]byte(tuningJSON), tc); err != nil {
			return jsError("parsing tuning: " + err.Error())
		}
	}

	// Parse weather data (raw Open-Meteo JSON response)
	weatherData, err := pgforecast.ParseOpenMeteoJSON([]byte(weatherJSON))
	if err != nil {
		return jsError("parsing weather: " + err.Error())
	}

	// Compute metrics for each hour
	results := make([]pgforecast.HourlyMetrics, len(weatherData))
	for i, h := range weatherData {
		results[i] = pgforecast.ComputeHourlyMetrics(&h, site, tc)
	}

	out, err := json.Marshal(results)
	if err != nil {
		return jsError("marshalling results: " + err.Error())
	}
	return string(out)
}

func defaultTuning(_ js.Value, _ []js.Value) interface{} {
	tc := pgforecast.DefaultTuningConfig()
	out, _ := json.Marshal(tc)
	return string(out)
}

func degreesToCompass(_ js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return ""
	}
	return pgforecast.DegreesToCompass(args[0].Float())
}

func jsError(msg string) interface{} {
	result := map[string]string{"error": msg}
	out, _ := json.Marshal(result)
	return string(out)
}
