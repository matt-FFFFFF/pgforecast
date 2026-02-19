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
	case GradientLow: return "✅"
	case GradientMedium: return "⚠️"
	default: return "🔴"
	}
}

func thermalIcon(t string) string {
	switch t {
	case ThermalNone: return "❄️"
	case ThermalWeak: return "🌤"
	case ThermalModerate: return "☀️"
	case ThermalStrong: return "🔥"
	case ThermalExtreme: return "⚡"
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
	case XCEpic: return "🚀"
	case XCHigh: return "🦅"
	case XCMedium: return "🪂"
	default: return ""
	}
}

// FormatText writes a pretty text forecast to the writer.
func FormatText(w io.Writer, f *SiteForecast, tc *TuningConfig) {
	fmt.Fprintf(w, "\n"+ForecastTitle+"\n", f.Site.Name)
	fmt.Fprintf(w, "   %s %s | %s %s | %s %dm\n",
		DegreesToCompass(float64(f.Site.Aspect)),
		LabelFacing,
		LabelIdeal,
		windRangeStr(f.Site.WindMin, f.Site.WindMax, f.Site.BestDir),
		LabelElev,
		f.Site.Elevation)
	fmt.Fprintf(w, "   %s %s\n", LabelGenerated, f.Generated.Format("Mon 2 Jan 2006 15:04 MST"))

	for i, day := range f.DetailedDays {
		label := LabelToday
		if i == 1 { label = LabelTomorrow }
		if i >= 2 { label = day.Date.Format("Mon 2 Jan") }
		
		fmt.Fprintf(w, "\n━━━ %s (%s) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", label, day.Date.Format("Mon 2 Jan"))
		fmt.Fprintf(w, "        %-8s %-5s %-6s %-9s %-10s %-5s %-6s %s\n",
			HeaderWind, HeaderDir, HeaderGust, HeaderGradient, HeaderThermal, HeaderCloud, HeaderRain, HeaderScore)

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
			fmt.Fprintf(w, "\n"+CloudbaseLabel+"\n",
				CloudbaseStr(h0.CloudbaseFt, tc), h0.CAPE, h0.FreezingLevel)
			fmt.Fprintf(w, OrographicLabel+"\n", h0.OrographicLift)
			fmt.Fprintf(w, XCLabel+"\n", s.XCPotential, xcIcon(s.XCPotential))
		}
	}

	if len(f.ExtendedDays) > 0 {
		fmt.Fprintf(w, "\n━━━ "+ExtendedOutlookTitle+" ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(w, "%-12s %-10s %-6s %-10s %-6s %s\n",
			HeaderDay, HeaderExtWind, HeaderExtDir, HeaderExtThermal, HeaderExtRain, HeaderExtScore)
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
		fmt.Fprintf(w, "\n"+BestWindowLabel+"\n", f.BestWindow)
	}
	fmt.Fprintln(w)
}

func windRangeStr(min, max, best int) string {
	return fmt.Sprintf("%s-%s (%s)", DegreesToCompass(float64(min)), DegreesToCompass(float64(max)), DegreesToCompass(float64(best)))
}
