package pgforecast

import (
	"testing"
)

func TestDegreesToCompass(t *testing.T) {
	tests := []struct {
		deg  float64
		want string
	}{
		{0, "N"},
		{45, "NE"},
		{90, "E"},
		{135, "SE"},
		{180, "S"},
		{225, "SW"},
		{270, "W"},
		{315, "NW"},
		{360, "N"},
		{350, "N"},
		{10, "N"},
		{22.5, "NNE"},
		{11.24, "N"},
		{11.26, "NNE"},
		{-10, "N"}, // negative wraps
	}
	for _, tt := range tests {
		got := DegreesToCompass(tt.deg)
		if got != tt.want {
			t.Errorf("DegreesToCompass(%v) = %q, want %q", tt.deg, got, tt.want)
		}
	}
}

func TestCalcWindGradient(t *testing.T) {
	tc := DefaultTuningConfig()

	tests := []struct {
		name       string
		surface    float64
		levels     []PressureLevel
		wantRating string
		wantLow    bool // diff should be low
	}{
		{
			name:    "no gradient",
			surface: 12,
			levels: []PressureLevel{
				{Pressure: 950, WindSpeed: 14},
				{Pressure: 900, WindSpeed: 13},
				{Pressure: 850, WindSpeed: 15},
			},
			wantRating: "Low",
		},
		{
			name:    "medium gradient",
			surface: 10,
			levels: []PressureLevel{
				{Pressure: 950, WindSpeed: 15},
				{Pressure: 900, WindSpeed: 22},
				{Pressure: 850, WindSpeed: 25},
			},
			wantRating: "Medium",
		},
		{
			name:    "high gradient",
			surface: 10,
			levels: []PressureLevel{
				{Pressure: 950, WindSpeed: 20},
				{Pressure: 900, WindSpeed: 30},
				{Pressure: 850, WindSpeed: 35},
			},
			wantRating: "High",
		},
		{
			name:    "700hPa excluded from gradient",
			surface: 10,
			levels: []PressureLevel{
				{Pressure: 950, WindSpeed: 12},
				{Pressure: 850, WindSpeed: 15},
				{Pressure: 700, WindSpeed: 80}, // jet stream — should be ignored
			},
			wantRating: "Low",
		},
		{
			name:    "1000hPa included",
			surface: 10,
			levels: []PressureLevel{
				{Pressure: 1000, WindSpeed: 30},
				{Pressure: 950, WindSpeed: 12},
			},
			wantRating: "High",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, rating := CalcWindGradient(tt.surface, tt.levels, tc)
			if rating != tt.wantRating {
				t.Errorf("got rating %q (diff=%.1f), want %q", rating, diff, tt.wantRating)
			}
			if diff < 0 {
				t.Error("gradient diff should never be negative")
			}
		})
	}
}

func TestCalcCloudbaseFt(t *testing.T) {
	tc := DefaultTuningConfig()

	tests := []struct {
		name    string
		temp    float64
		dew     float64
		wantMin int
		wantMax int
	}{
		{"large spread", 20, 10, 3500, 4500},
		{"small spread", 15, 14, 200, 600},
		{"zero spread (fog)", 10, 10, 200, 200},     // clamped to min
		{"negative spread", 10, 12, 200, 200},        // clamped to min
		{"moderate spread", 25, 15, 3500, 4500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcCloudbaseFt(tt.temp, tt.dew, tc)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalcCloudbaseFt(%v, %v) = %d, want %d-%d", tt.temp, tt.dew, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCloudbaseStr(t *testing.T) {
	tc := DefaultTuningConfig()

	if s := CloudbaseStr(200, tc); s != "Fog" {
		t.Errorf("CloudbaseStr(200) = %q, want Fog", s)
	}
	if s := CloudbaseStr(100, tc); s != "Fog" {
		t.Errorf("CloudbaseStr(100) = %q, want Fog", s)
	}
	if s := CloudbaseStr(1500, tc); s != "1500ft" {
		t.Errorf("CloudbaseStr(1500) = %q, want 1500ft", s)
	}
}

func TestCalcOrographicLift(t *testing.T) {
	tc := DefaultTuningConfig()

	tests := []struct {
		name       string
		windDir    float64
		windSpeed  float64
		siteAspect int
		want       string
	}{
		{"direct into face", 225, 15, 225, "Strong"},
		{"slight angle", 240, 15, 225, "Strong"},
		{"moderate angle", 260, 15, 225, "Weak"},      // 35° off = beyond moderate (30°)
		{"large angle", 280, 15, 225, "None"},          // 55° off = beyond weak (45°)
		{"off face", 45, 15, 225, "None"},
		{"too light", 225, 5, 225, "None"},
		{"wrap around north", 350, 12, 10, "Moderate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcOrographicLift(tt.windDir, tt.windSpeed, tt.siteAspect, tc)
			if got != tt.want {
				t.Errorf("CalcOrographicLift(dir=%v, spd=%v, asp=%v) = %q, want %q",
					tt.windDir, tt.windSpeed, tt.siteAspect, got, tt.want)
			}
		})
	}
}

func TestCalcThermalRating(t *testing.T) {
	tc := DefaultTuningConfig()

	tests := []struct {
		name string
		cape float64
		want string
	}{
		// CalcThermalRating uses cumulative scoring (CAPE bracket + lapse rate bonus)
		// With neutral lapse rate (~6.7°C/km), no lapse bonus applies
		{"zero cape", 0, "None"},
		{"weak cape", 150, "Weak"},         // CAPE score=1 → Weak
		{"moderate cape", 500, "Weak"},     // CAPE score=2 → still Weak (need 3 for Moderate)
		{"strong cape", 1500, "Moderate"},  // CAPE score=3 → Moderate
		{"extreme cape", 3000, "Strong"},   // CAPE score=4 → Strong
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use neutral lapse rate levels (won't add bonus)
			levels := []PressureLevel{
				{Pressure: 925, Temperature: 10, GeopotentialHeight: 750},
				{Pressure: 700, Temperature: -5, GeopotentialHeight: 3000},
			}
			got := CalcThermalRating(tt.cape, levels, tc)
			if got != tt.want {
				t.Errorf("CalcThermalRating(cape=%v) = %q, want %q", tt.cape, got, tt.want)
			}
		})
	}
}

func TestIsInWindRange(t *testing.T) {
	tests := []struct {
		dir  float64
		min  int
		max  int
		want bool
	}{
		{225, 210, 260, true},
		{200, 210, 260, false},
		{270, 210, 260, false},
		// Wrapping range (e.g. NNW-NNE through north)
		{350, 330, 30, true},
		{10, 330, 30, true},
		{180, 330, 30, false},
		{330, 330, 30, true},
		{30, 330, 30, true},
	}
	for _, tt := range tests {
		got := isInWindRange(tt.dir, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("isInWindRange(%v, %d, %d) = %v, want %v", tt.dir, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestDistanceFromWindRange(t *testing.T) {
	tests := []struct {
		dir     float64
		min     int
		max     int
		wantMin float64
		wantMax float64
	}{
		{225, 210, 260, 0, 0},           // inside
		{200, 210, 260, 9, 11},          // just outside min
		{300, 210, 260, 39, 41},         // outside max
		{90, 210, 260, 119, 121},        // way off
		{10, 330, 30, 0, 0},            // inside wrapping range
		{180, 330, 30, 149, 151},        // outside wrapping range
	}
	for _, tt := range tests {
		got := distanceFromWindRange(tt.dir, tt.min, tt.max)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("distanceFromWindRange(%v, %d, %d) = %v, want %v-%v",
				tt.dir, tt.min, tt.max, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestCalcFlyabilityScore(t *testing.T) {
	tc := DefaultTuningConfig()

	ringstead := Site{
		Name: "Ringstead", WindMin: 210, WindMax: 260,
		BestDir: 225, Aspect: 225, Elevation: 147,
	}

	tests := []struct {
		name     string
		hourly   HourlyData
		gradient string
		thermal  string
		wantMin  int
		wantMax  int
	}{
		{
			name: "perfect conditions",
			hourly: HourlyData{
				WindSpeed: 14, WindDirection: 225, WindGusts: 18,
				Precipitation: 0, PrecipitationProbability: 0, CAPE: 0,
			},
			gradient: "Low",
			thermal:  "None",
			wantMin:  4, wantMax: 5,
		},
		{
			name: "strong wind off direction with rain",
			hourly: HourlyData{
				WindSpeed: 30, WindDirection: 45, WindGusts: 45,
				Precipitation: 2.0, PrecipitationProbability: 90, CAPE: 0,
			},
			gradient: "High",
			thermal:  "None",
			wantMin:  1, wantMax: 1,
		},
		{
			name: "ideal wind but high gradient",
			hourly: HourlyData{
				WindSpeed: 14, WindDirection: 225, WindGusts: 18,
				Precipitation: 0, PrecipitationProbability: 0, CAPE: 0,
			},
			gradient: "High",
			thermal:  "None",
			wantMin:  3, wantMax: 4,
		},
		{
			name: "gusty conditions",
			hourly: HourlyData{
				WindSpeed: 10, WindDirection: 225, WindGusts: 25, // 2.5x gust factor
				Precipitation: 0, PrecipitationProbability: 0, CAPE: 0,
			},
			gradient: "Low",
			thermal:  "None",
			wantMin:  2, wantMax: 4,
		},
		{
			name: "thermic day",
			hourly: HourlyData{
				WindSpeed: 12, WindDirection: 230, WindGusts: 16,
				Precipitation: 0, PrecipitationProbability: 0, CAPE: 500,
			},
			gradient: "Low",
			thermal:  "Moderate",
			wantMin:  5, wantMax: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcFlyabilityScore(&tt.hourly, ringstead, tt.gradient, tt.thermal, tc)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalcFlyabilityScore(%s) = %d, want %d-%d", tt.name, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalcFlyabilityScoreClamped(t *testing.T) {
	tc := DefaultTuningConfig()
	site := Site{WindMin: 0, WindMax: 360, BestDir: 0, Aspect: 0}

	// Should never go below 1
	terrible := HourlyData{
		WindSpeed: 50, WindDirection: 180, WindGusts: 80,
		Precipitation: 10, PrecipitationProbability: 100,
	}
	got := CalcFlyabilityScore(&terrible, site, "High", "None", tc)
	if got != 1 {
		t.Errorf("worst case score = %d, want 1", got)
	}
}

func TestCalcXCPotential(t *testing.T) {
	tc := DefaultTuningConfig()

	tests := []struct {
		name    string
		cape    float64
		cloud   int
		wind    float64
		thermal string
		want    string
	}{
		{"poor conditions", 50, 500, 5, "None", "Low"},
		{"moderate conditions", 400, 3500, 12, "Moderate", "Medium"},
		{"strong conditions", 1200, 5000, 15, "Strong", "Epic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcXCPotential(tt.cape, tt.cloud, tt.wind, tt.thermal, tc)
			if got != tt.want {
				t.Errorf("CalcXCPotential = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeHourlyMetrics(t *testing.T) {
	tc := DefaultTuningConfig()
	site := Site{
		Name: "Test", WindMin: 210, WindMax: 260,
		BestDir: 225, Aspect: 225, Elevation: 150,
	}
	h := HourlyData{
		WindSpeed: 14, WindDirection: 225, WindGusts: 18,
		Temperature: 18, DewPoint: 10, CAPE: 200,
		CloudCover: 50, Precipitation: 0, PrecipitationProbability: 10,
		FreezingLevelHeight: 2000, IsDay: 1,
		PressureLevels: []PressureLevel{
			{Pressure: 950, WindSpeed: 16, Temperature: 12, GeopotentialHeight: 500},
			{Pressure: 850, WindSpeed: 18, Temperature: 5, GeopotentialHeight: 1500},
			{Pressure: 700, WindSpeed: 40, Temperature: -5, GeopotentialHeight: 3000},
		},
	}

	m := ComputeHourlyMetrics(&h, site, tc)

	if m.WindDirStr != "SW" {
		t.Errorf("WindDirStr = %q, want SW", m.WindDirStr)
	}
	if m.WindGradient != "Low" {
		t.Errorf("WindGradient = %q, want Low", m.WindGradient)
	}
	if m.OrographicLift != "Strong" {
		t.Errorf("OrographicLift = %q, want Strong", m.OrographicLift)
	}
	if m.CloudbaseFt < 2000 || m.CloudbaseFt > 4000 {
		t.Errorf("CloudbaseFt = %d, want 2000-4000", m.CloudbaseFt)
	}
	if m.FlyabilityScore < 4 {
		t.Errorf("FlyabilityScore = %d, want >= 4", m.FlyabilityScore)
	}
	if !m.IsDay {
		t.Error("IsDay should be true")
	}
	if len(m.PressureLevels) != 3 {
		t.Fatalf("PressureLevels length = %d, want 3", len(m.PressureLevels))
	}
	if m.PressureLevels[0].Pressure != 950 {
		t.Errorf("PressureLevels[0].Pressure = %d, want 950", m.PressureLevels[0].Pressure)
	}
}

func TestDefaultTuningConfig(t *testing.T) {
	tc := DefaultTuningConfig()
	if tc == nil {
		t.Fatal("DefaultTuningConfig returned nil")
	}
	if tc.Wind.IdealMin >= tc.Wind.IdealMax {
		t.Errorf("IdealMin (%v) should be < IdealMax (%v)", tc.Wind.IdealMin, tc.Wind.IdealMax)
	}
	if tc.Wind.AcceptableMax >= tc.Wind.DangerousMax {
		t.Errorf("AcceptableMax (%v) should be < DangerousMax (%v)", tc.Wind.AcceptableMax, tc.Wind.DangerousMax)
	}
	if tc.Scoring.BaseScore <= 0 {
		t.Errorf("BaseScore should be > 0, got %v", tc.Scoring.BaseScore)
	}
	if tc.Cloudbase.MinRealisticFt <= 0 {
		t.Errorf("MinRealisticFt should be > 0, got %d", tc.Cloudbase.MinRealisticFt)
	}
}
