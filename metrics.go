package pgforecast

import (
	"fmt"
	"math"
)

const (
	// DegreesPerCompassPoint is the angular width of each of the 16 compass directions (360/16).
	DegreesPerCompassPoint = 22.5

	// CompassDirectionCount is the number of compass directions (N, NNE, NE, ...).
	CompassDirectionCount = 16

	// PressureLevelMinHPa is the lower bound (highest pressure) for flyable pressure levels (~sea level).
	PressureLevelMinHPa = 850

	// PressureLevelMaxHPa is the upper bound (lowest altitude) for flyable pressure levels.
	PressureLevelMaxHPa = 1000

	// LapseRatePressureHigh is the pressure level (hPa) used as the lower altitude reference for lapse rate.
	LapseRatePressureHigh = 925

	// LapseRatePressureLow is the pressure level (hPa) used as the upper altitude reference for lapse rate.
	LapseRatePressureLow = 700

	// DefaultLapseRate is the standard atmospheric lapse rate (°C/km) used when pressure level data is unavailable.
	DefaultLapseRate = 6.5

	// MetersPerKm converts meters to kilometers for lapse rate calculation.
	MetersPerKm = 1000.0

	// LapseRateStrongThreshold is the lapse rate (°C/km) above which an extra thermal score point is awarded.
	LapseRateStrongThreshold = 9.0

	// SpreadToCloudbaseDivisor is the temperature spread divisor in the cloudbase estimation formula.
	SpreadToCloudbaseDivisor = 2.5

	// SpreadToCloudbaseMultiplier converts the spread ratio to feet in the cloudbase estimation formula.
	SpreadToCloudbaseMultiplier = 1000

	// MetersToFeet is the conversion factor from meters to feet.
	MetersToFeet = 3.28084

	// DegreesHalfCircle is 180°, used in angle wrapping calculations.
	DegreesHalfCircle = 180

	// DegreesFullCircle is 360°, used in angle wrapping and modular arithmetic.
	DegreesFullCircle = 360

	// ScoreMin is the minimum clamped flyability score.
	ScoreMin = 1

	// ScoreMax is the maximum clamped flyability score.
	ScoreMax = 5

	// PrecipProbHighThreshold is the precipitation probability (%) above which a rain penalty applies.
	PrecipProbHighThreshold = 50

	// PrecipProbLowThreshold is the precipitation probability (%) above which a minor rain penalty applies.
	PrecipProbLowThreshold = 30

	// WindDirFarOffAngle is the angular distance (°) beyond which a full off-direction penalty applies.
	WindDirFarOffAngle = 90.0

	// WindDirModerateOffAngle is the angular distance (°) beyond which a moderate off-direction penalty applies.
	WindDirModerateOffAngle = 45.0

	// WindDirMarginalAngle is the angular distance (°) within which no off-direction penalty applies.
	WindDirMarginalAngle = 20.0
)

// DegreesToCompass converts degrees to compass direction string.
func DegreesToCompass(deg float64) string {
	dirs := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	idx := int(math.Round(deg/DegreesPerCompassPoint)) % CompassDirectionCount
	if idx < 0 {
		idx += CompassDirectionCount
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
		if l.Pressure >= PressureLevelMinHPa && l.Pressure <= PressureLevelMaxHPa && l.WindSpeed > maxUpper {
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
	if lapseRate > LapseRateStrongThreshold {
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
		if l.Pressure == LapseRatePressureHigh {
			t925 = l.Temperature
			h925 = l.GeopotentialHeight
			found925 = true
		}
		if l.Pressure == LapseRatePressureLow {
			t700 = l.Temperature
			h700 = l.GeopotentialHeight
			found700 = true
		}
	}
	if !found925 || !found700 || h700 <= h925 {
		return DefaultLapseRate
	}
	return (t925 - t700) / ((h700 - h925) / MetersPerKm)
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
	ft := int(spread / SpreadToCloudbaseDivisor * SpreadToCloudbaseMultiplier)
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
	if d > DegreesHalfCircle {
		d = DegreesFullCircle - d
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
		case distFromRange > WindDirFarOffAngle:
			score += s.DirOffPenalty // e.g. -2.0
		case distFromRange > WindDirModerateOffAngle:
			score -= 1.0
		case distFromRange > WindDirMarginalAngle:
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
	} else if h.PrecipitationProbability > PrecipProbHighThreshold {
		score += s.RainProbPenalty
	} else if h.PrecipitationProbability > PrecipProbLowThreshold {
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
	if result < ScoreMin {
		result = ScoreMin
	}
	if result > ScoreMax {
		result = ScoreMax
	}
	return result
}

func isInWindRange(dir float64, min, max int) bool {
	d := int(dir) % DegreesFullCircle
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
		FreezingLevel:    h.FreezingLevelHeight * MetersToFeet,
		IsDay:            h.IsDay == 1,
		PressureLevels:   h.PressureLevels,
	}
}
