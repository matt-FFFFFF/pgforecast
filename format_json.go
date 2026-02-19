package pgforecast

import (
	"encoding/json"
	"io"
)

// jsonOutput wraps the forecast with display configuration for the frontend.
type jsonOutput struct {
	*SiteForecast
	Display DisplayConfig `json:"display"`
}

// FormatJSON writes the forecast as JSON, including display configuration.
func FormatJSON(w io.Writer, f *SiteForecast, tc *TuningConfig) error {
	out := jsonOutput{
		SiteForecast: f,
		Display:      tc.Display,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
