package pgforecast

import (
	"fmt"
	"math"
)

// DegreesToCompass converts degrees to compass direction string.
func DegreesToCompass(deg float64) string {
	dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	idx := int(math.Round(deg/22.5)) % 16
	if idx < 0 {
		idx += 16
	}
	return dirs[idx]
}

// CalcWindGradient computes wind speed difference between surface and flyable
// pressure levels (1000-850 hPa, roughly 0-1500m). 700hPa (~3000m) is excluded
// as it's jet stream territory and always shows high winds, making the metric
// useless for paragliding decisions.
func CalcWindGradient(surface float64, levels []PressureLevel, tc *TuningConfig) (diff float64, rating string) {
	maxUpper := surface
	for _, l := range levels {
		// Only check levels a paraglider might actually reach: 1000-850 hPa
		// 1000 ~60-160m, 950 ~500m, 925 ~750m, 900 ~1000m, 850 ~1500m
		if l.Pressure >= 850 && l.Pressure <= 1000 && l.WindSpeed > maxUpper {
			maxUpper = l.WindSpeed
		}
	}
	diff = maxUpper - surface
	if diff < 0 {
		diff = 0
	}
	switch {
	case diff < tc.Gradient.LowThreshold:
		rating = "Low"
	case diff < tc.Gradient.HighThreshold:
		rating = "Medium"
	default:
		rating = "High"
	}
	return
}

// CalcThermalRating estimates thermal potential from CAPE and lapse rate.
func CalcThermalRating(cape float64, levels []PressureLevel, tc *TuningConfig) string {
	lapseRate := calcLapseRate(levels)

	score := 0.0
	switch {
	case cape > tc.Thermal.CAPEExtreme:
		score += 4
	case cape > tc.Thermal.CAPEStrong:
		score += 3
	case cape > tc.Thermal.CAPEModerate:
		score += 2
	case cape > tc.Thermal.CAPEWeak:
		score += 1
	}
	if lapseRate > tc.Thermal.LapseRateBonus {
		score += 1
	}
	if lapseRate > 9.0 {
		score += 1
	}

	switch {
	case score >= 5:
		return "Extreme"
	case score >= 4:
		return "Strong"
	case score >= 3:
		return "Moderate"
	case score >= 1:
		return "Weak"
	default:
		return "None"
	}
}

func calcLapseRate(levels []PressureLevel) float64 {
	var t925, h925, t700, h700 float64
	var found925, found700 bool
	for _, l := range levels {
		if l.Pressure == 925 {
			t925 = l.Temperature
			h925 = l.GeopotentialHeight
			found925 = true
		}
		if l.Pressure == 700 {
			t700 = l.Temperature
			h700 = l.GeopotentialHeight
			found700 = true
		}
	}
	if !found925 || !found700 || h700 <= h925 {
		return 6.5
	}
	return (t925 - t700) / ((h700 - h925) / 1000.0)
}

// CalcCAPERating rates CAPE value.
func CalcCAPERating(cape float64, tc *TuningConfig) string {
	switch {
	case cape >= tc.Thermal.CAPEExtreme:
		return "Overdevelopment"
	case cape >= tc.Thermal.CAPEStrong:
		return "Strong"
	case cape >= tc.Thermal.CAPEModerate:
		return "Moderate"
	default:
		return "Weak"
	}
}

// CalcCloudbaseFt estimates cloudbase in feet from temp and dewpoint.
func CalcCloudbaseFt(temp, dewpoint float64, tc *TuningConfig) int {
	spread := temp - dewpoint
	if spread < 0 {
		spread = 0
	}
	ft := int(spread / 2.5 * 1000)
	if ft < tc.Cloudbase.MinRealisticFt {
		ft = tc.Cloudbase.MinRealisticFt
	}
	return ft
}

// CloudbaseStr returns a display string for cloudbase, showing "Fog" for very low values.
func CloudbaseStr(ft int, tc *TuningConfig) string {
	if ft <= tc.Cloudbase.MinRealisticFt {
		return "Fog"
	}
	return fmt.Sprintf("%dft", ft)
}

// CalcOrographicLift rates orographic lift potential.
func CalcOrographicLift(windDir, windSpeed float64, siteAspect int, tc *TuningConfig) string {
	diff := angleDiff(windDir, float64(siteAspect))

	if windSpeed < tc.Orographic.MinWindSpeed {
		return "None"
	}

	switch {
	case diff <= tc.Orographic.StrongAngle:
		return "Strong"
	case diff <= tc.Orographic.ModerateAngle:
		return "Moderate"
	case diff <= tc.Orographic.WeakAngle:
		return "Weak"
	default:
		return "None"
	}
}

func angleDiff(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

// CalcFlyabilityScore calculates a 1-5 star rating using TuningConfig.
func CalcFlyabilityScore(h *HourlyData, site Site, gradientRating string, thermalRating string, tc *TuningConfig) int {
	s := tc.Scoring

	score := s.BaseScore

	// Wind speed
	ws := h.WindSpeed
	switch {
	case ws >= tc.Wind.IdealMin && ws <= tc.Wind.IdealMax:
		score += s.WindIdealBonus
	case ws >= tc.Wind.AcceptableMin && ws <= tc.Wind.AcceptableMax:
		score += s.WindAcceptableBonus
	case ws > tc.Wind.DangerousMax:
		score += s.WindDangerPenalty
	case ws > tc.Wind.AcceptableMax:
		score += s.WindHighPenalty
	}

	// Wind direction
	if isInWindRange(h.WindDirection, site.WindMin, site.WindMax) {
		score += s.DirOnBonus
	} else {
		// How far outside the acceptable range?
		distFromRange := distanceFromWindRange(h.WindDirection, site.WindMin, site.WindMax)
		switch {
		case distFromRange > 90:
			score += s.DirOffPenalty // e.g. -2.0
		case distFromRange > 45:
			score -= 1.0
		case distFromRange > 20:
			score -= 0.5
		// Within 20° of range edge — marginal, no penalty
		}
	}

	// Gusts
	if h.WindSpeed > 0 {
		gustFactor := h.WindGusts / h.WindSpeed
		if gustFactor > tc.Wind.DangerousGustFactor {
			score += s.GustHighPenalty
		} else if gustFactor > tc.Wind.MaxGustFactor {
			score += s.GustMedPenalty
		}
	}

	// Wind gradient (NEW - was missing from scoring)
	switch gradientRating {
	case "High":
		score += s.GradientHighPenalty
	case "Medium":
		score += s.GradientMedPenalty
	}

	// Rain
	if h.Precipitation > 0 {
		score += s.RainPenalty
	} else if h.PrecipitationProbability > 50 {
		score += s.RainProbPenalty
	} else if h.PrecipitationProbability > 30 {
		score -= 0.25
	}

	// Thermals bonus
	if h.CAPE >= tc.Thermal.CAPEModerate && h.CAPE < tc.Thermal.CAPEExtreme {
		score += s.CAPEBonus
	}
	if thermalRating == "Strong" || thermalRating == "Moderate" {
		score += s.ThermalStrongBonus
	}

	// Clamp 1-5
	result := int(math.Round(score))
	if result < 1 {
		result = 1
	}
	if result > 5 {
		result = 5
	}
	return result
}

func isInWindRange(dir float64, min, max int) bool {
	d := int(dir) % 360
	if min <= max {
		return d >= min && d <= max
	}
	return d >= min || d <= max
}

// distanceFromWindRange returns the minimum angular distance from dir to the
// nearest edge of the [min, max] wind range. Returns 0 if inside the range.
func distanceFromWindRange(dir float64, min, max int) float64 {
	if isInWindRange(dir, min, max) {
		return 0
	}
	dMin := angleDiff(dir, float64(min))
	dMax := angleDiff(dir, float64(max))
	if dMin < dMax {
		return dMin
	}
	return dMax
}

// CalcXCPotential rates cross-country potential.
func CalcXCPotential(cape float64, cloudbaseFt int, windSpeed float64, thermalRating string, tc *TuningConfig) string {
	score := 0
	if cape >= tc.Thermal.CAPEStrong {
		score += 2
	} else if cape >= tc.Thermal.CAPEModerate {
		score++
	}
	if cloudbaseFt >= tc.XC.GoodCloudbaseFt {
		score += 2
	} else if cloudbaseFt >= tc.XC.MinCloudbaseFt {
		score++
	}
	if windSpeed >= tc.XC.MinWindSpeed && windSpeed <= tc.XC.MaxWindSpeed {
		score++
	}
	switch thermalRating {
	case "Strong":
		score += 2
	case "Moderate":
		score++
	case "Extreme":
		score++
	}

	switch {
	case score >= tc.XC.EpicThreshold:
		return "Epic"
	case score >= tc.XC.HighThreshold:
		return "High"
	case score >= tc.XC.MediumThreshold:
		return "Medium"
	default:
		return "Low"
	}
}

// ComputeHourlyMetrics computes all paragliding metrics for one hour.
func ComputeHourlyMetrics(h *HourlyData, site Site, tc *TuningConfig) HourlyMetrics {
	gradientDiff, gradientRating := CalcWindGradient(h.WindSpeed, h.PressureLevels, tc)
	thermalRating := CalcThermalRating(h.CAPE, h.PressureLevels, tc)
	cloudbase := CalcCloudbaseFt(h.Temperature, h.DewPoint, tc)

	return HourlyMetrics{
		Time:             h.Time,
		WindSpeed:        h.WindSpeed,
		WindDirection:    h.WindDirection,
		WindDirStr:       DegreesToCompass(h.WindDirection),
		WindGusts:        h.WindGusts,
		WindGradient:     gradientRating,
		WindGradientDiff: gradientDiff,
		ThermalRating:    thermalRating,
		CAPE:             h.CAPE,
		CAPERating:       CalcCAPERating(h.CAPE, tc),
		CloudbaseFt:      cloudbase,
		CloudCover:       h.CloudCover,
		Precipitation:    h.Precipitation,
		PrecipProb:       h.PrecipitationProbability,
		OrographicLift:   CalcOrographicLift(h.WindDirection, h.WindSpeed, site.Aspect, tc),
		FlyabilityScore:  CalcFlyabilityScore(h, site, gradientRating, thermalRating, tc),
		XCPotential:      CalcXCPotential(h.CAPE, cloudbase, h.WindSpeed, thermalRating, tc),
		FreezingLevel:    h.FreezingLevelHeight * 3.28084,
		IsDay:            h.IsDay == 1,
	}
}
