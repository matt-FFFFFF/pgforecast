package pgforecast

// User-facing string constants, centralised for future localization.

// Wind gradient ratings.
const (
	GradientLow    = "Low"
	GradientMedium = "Medium"
	GradientHigh   = "High"
)

// Thermal ratings.
const (
	ThermalNone     = "None"
	ThermalWeak     = "Weak"
	ThermalModerate = "Moderate"
	ThermalStrong   = "Strong"
	ThermalExtreme  = "Extreme"
)

// CAPE ratings.
const (
	CAPEOverdevelopment = "Overdevelopment"
	CAPEStrong          = "Strong"
	CAPEModerate        = "Moderate"
	CAPEWeak            = "Weak"
)

// Orographic lift ratings.
const (
	OrographicNone     = "None"
	OrographicWeak     = "Weak"
	OrographicModerate = "Moderate"
	OrographicStrong   = "Strong"
)

// XC potential ratings.
const (
	XCEpic   = "Epic"
	XCHigh   = "High"
	XCMedium = "Medium"
	XCLow    = "Low"
)

// Cloudbase display strings.
const (
	CloudbaseFog = "Fog"
)

// Format labels used in text output.
const (
	LabelToday    = "TODAY"
	LabelTomorrow = "TOMORROW"
)

// Column headers for detailed forecast.
const (
	HeaderWind     = "Wind"
	HeaderDir      = "Dir"
	HeaderGust     = "Gust"
	HeaderGradient = "Gradient"
	HeaderThermal  = "Thermal"
	HeaderCloud    = "Cloud"
	HeaderRain     = "Rain"
	HeaderScore    = "Score"
)

// Column headers for extended outlook.
const (
	HeaderDay          = "Day"
	HeaderExtWind      = "Wind"
	HeaderExtDir       = "Dir"
	HeaderExtThermal   = "Thermal"
	HeaderExtRain      = "Rain"
	HeaderExtScore     = "Score"
	ExtendedOutlookTitle = "EXTENDED OUTLOOK"
)

// Format strings and labels.
const (
	ForecastTitle   = "🪂 PARAGLIDING FORECAST — %s"
	BestWindowLabel = "🏆 Best Window: %s"
	CloudbaseLabel  = "Cloudbase: ~%s | CAPE: %.0f J/kg | Freezing: %.0fft"
	OrographicLabel = "Orographic: %s"
	XCLabel         = "XC Potential: %s %s"
)
