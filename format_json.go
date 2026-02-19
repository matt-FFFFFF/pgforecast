package pgforecast

import (
	"encoding/json"
	"io"
)

// FormatJSON writes the forecast as JSON.
func FormatJSON(w io.Writer, f *SiteForecast) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(f)
}
