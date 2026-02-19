package pgforecast

import (
	"fmt"
	"io"
	"strings"
)

func starsStr(n int) string {
	return strings.Repeat("⭐", n)
}

func gradientIcon(g string) string {
	switch g {
	case "Low": return "✅"
	case "Medium": return "⚠️"
	default: return "🔴"
	}
}

func thermalIcon(t string) string {
	switch t {
	case "None": return "❄️"
	case "Weak": return "🌤"
	case "Moderate": return "☀️"
	case "Strong": return "🔥"
	case "Extreme": return "⚡"
	default: return "❓"
	}
}

func cloudIcon(cover float64) string {
	switch {
	case cover < 20: return "☀️"
	case cover < 50: return "⛅"
	case cover < 80: return "🌥"
	default: return "☁️"
	}
}

func rainStr(precip, prob float64) string {
	if precip > 0 { return fmt.Sprintf("🌧%.1f", precip) }
	if prob > 30 { return fmt.Sprintf("%0.f%%", prob) }
	return "-"
}

func xcIcon(xc string) string {
	switch xc {
	case "Epic": return "🚀"
	case "High": return "🦅"
	case "Medium": return "🪂"
	default: return ""
	}
}

// FormatText writes a pretty text forecast to the writer.
func FormatText(w io.Writer, f *SiteForecast, tc *TuningConfig) {
	fmt.Fprintf(w, "\n🪂 PARAGLIDING FORECAST — %s\n", f.Site.Name)
	fmt.Fprintf(w, "   %s facing | Ideal: %s | Elev: %dm\n",
		DegreesToCompass(float64(f.Site.Aspect)),
		windRangeStr(f.Site.WindMin, f.Site.WindMax, f.Site.BestDir),
		f.Site.Elevation)
	fmt.Fprintf(w, "   Generated: %s\n", f.Generated.Format("Mon 2 Jan 2006 15:04 MST"))

	for i, day := range f.DetailedDays {
		label := "TODAY"
		if i == 1 { label = "TOMORROW" }
		if i >= 2 { label = day.Date.Format("Mon 2 Jan") }
		
		fmt.Fprintf(w, "\n━━━ %s (%s) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", label, day.Date.Format("Mon 2 Jan"))
		fmt.Fprintf(w, "        %-8s %-5s %-6s %-9s %-10s %-5s %-6s %s\n",
			"Wind", "Dir", "Gust", "Gradient", "Thermal", "Cloud", "Rain", "Score")

		for _, h := range day.Hours {
			fmt.Fprintf(w, "%s  %-8s %-5s %-6s %s %-5s %s %-7s %-5s %-6s %s\n",
				h.Time.Format("15:04"),
				fmt.Sprintf("%.0f%s", h.WindSpeed, f.Units),
				h.WindDirStr,
				fmt.Sprintf("%.0f", h.WindGusts),
				gradientIcon(h.WindGradient),
				fmt.Sprintf("%s(+%.0f)", h.WindGradient, h.WindGradientDiff),
				thermalIcon(h.ThermalRating),
				h.ThermalRating,
				cloudIcon(h.CloudCover),
				rainStr(h.Precipitation, h.PrecipProb),
				starsStr(h.FlyabilityScore))
		}

		s := day.Summary
		if len(day.Hours) > 0 {
			h0 := day.Hours[len(day.Hours)/2] // mid-day representative
			fmt.Fprintf(w, "\nCloudbase: ~%s | CAPE: %.0f J/kg | Freezing: %.0fft\n",
				CloudbaseStr(h0.CloudbaseFt, tc), h0.CAPE, h0.FreezingLevel)
			fmt.Fprintf(w, "Orographic: %s\n", h0.OrographicLift)
			fmt.Fprintf(w, "XC Potential: %s %s\n", s.XCPotential, xcIcon(s.XCPotential))
		}
	}

	if len(f.ExtendedDays) > 0 {
		fmt.Fprintf(w, "\n━━━ EXTENDED OUTLOOK ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(w, "%-12s %-10s %-6s %-10s %-6s %s\n",
			"Day", "Wind", "Dir", "Thermal", "Rain", "Score")
		for _, d := range f.ExtendedDays {
			fmt.Fprintf(w, "%-12s %-10s %-6s %-10s %-6s %s\n",
				d.Date.Format("Mon 2 Jan"),
				fmt.Sprintf("%.0f%s", d.AvgWindSpeed, f.Units),
				d.WindDirStr,
				d.ThermalRating,
				fmt.Sprintf("%.0f%%", d.MaxPrecipProb),
				starsStr(d.BestScore))
		}
	}

	if f.BestWindow != "" {
		fmt.Fprintf(w, "\n🏆 Best Window: %s\n", f.BestWindow)
	}
	fmt.Fprintln(w)
}

func windRangeStr(min, max, best int) string {
	return fmt.Sprintf("%s-%s (%s)", DegreesToCompass(float64(min)), DegreesToCompass(float64(max)), DegreesToCompass(float64(best)))
}
