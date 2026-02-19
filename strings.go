package pgforecast

// User-facing string constants, centralised for future localization.

// Wind gradient ratings.
const (
	// GradientLow indicates a safe wind gradient with minimal wind shear.
	GradientLow = "Low"
	// GradientMedium indicates a moderate wind gradient suitable for most pilots.
	GradientMedium = "Medium"
	// GradientHigh indicates a strong wind gradient with significant wind shear.
	GradientHigh = "High"
)

// Thermal ratings.
const (
	// ThermalNone indicates no usable thermal activity.
	ThermalNone = "None"
	// ThermalWeak indicates weak thermals with limited climb rates.
	ThermalWeak = "Weak"
	// ThermalModerate indicates reliable thermals with moderate climb rates.
	ThermalModerate = "Moderate"
	// ThermalStrong indicates strong thermals with robust climb rates.
	ThermalStrong = "Strong"
	// ThermalExtreme indicates very strong or turbulent thermals requiring high pilot skill.
	ThermalExtreme = "Extreme"
)

// CAPE ratings.
const (
	// CAPEOverdevelopment indicates CAPE levels that may lead to overdevelopment and shutdown of soaring.
	CAPEOverdevelopment = "Overdevelopment"
	// CAPEStrong indicates strong convective potential with high CAPE.
	CAPEStrong = "Strong"
	// CAPEModerate indicates moderate convective potential.
	CAPEModerate = "Moderate"
	// CAPEWeak indicates weak convective potential with limited storm or cloud growth.
	CAPEWeak = "Weak"
)

// Orographic lift ratings.
const (
	// OrographicNone indicates no significant orographic lift.
	OrographicNone = "None"
	// OrographicWeak indicates weak orographic lift from terrain.
	OrographicWeak = "Weak"
	// OrographicModerate indicates useful, consistent orographic lift.
	OrographicModerate = "Moderate"
	// OrographicStrong indicates strong orographic lift suitable for extended soaring.
	OrographicStrong = "Strong"
)

// XC potential ratings.
const (
	// XCEpic indicates exceptional conditions for long-distance cross-country flights.
	XCEpic = "Epic"
	// XCHigh indicates very good cross-country potential.
	XCHigh = "High"
	// XCMedium indicates moderate cross-country potential.
	XCMedium = "Medium"
	// XCLow indicates limited cross-country potential.
	XCLow = "Low"
)

// Cloudbase display strings.
const (
	// CloudbaseFog indicates conditions of fog or zero cloudbase.
	CloudbaseFog = "Fog"
)

// Format labels used in text output.
const (
	// LabelToday is the label used for the current day's forecast section.
	LabelToday = "TODAY"
	// LabelTomorrow is the label used for the next day's forecast section.
	LabelTomorrow = "TOMORROW"
)

// Site description labels.
const (
	// LabelFacing is the label for the site's facing direction.
	LabelFacing = "facing"
	// LabelIdeal is the label for the site's ideal wind conditions.
	LabelIdeal = "Ideal:"
	// LabelElev is the label for the site's elevation.
	LabelElev = "Elev:"
	// LabelGenerated is the label for the forecast generation timestamp.
	LabelGenerated = "Generated:"
)

// Column headers for detailed forecast.
const (
	// HeaderWind is the column header for mean wind speed.
	HeaderWind = "Wind"
	// HeaderDir is the column header for wind direction.
	HeaderDir = "Dir"
	// HeaderGust is the column header for gust speed.
	HeaderGust = "Gust"
	// HeaderGradient is the column header for wind gradient rating.
	HeaderGradient = "Gradient"
	// HeaderThermal is the column header for thermal strength rating.
	HeaderThermal = "Thermal"
	// HeaderCloud is the column header for cloud or cloudbase information.
	HeaderCloud = "Cloud"
	// HeaderRain is the column header for precipitation chance or intensity.
	HeaderRain = "Rain"
	// HeaderScore is the column header for the overall site or time-slot score.
	HeaderScore = "Score"
)

// Column headers for extended outlook.
const (
	// HeaderDay is the column header for the day label in the extended outlook.
	HeaderDay = "Day"
	// HeaderExtWind is the column header for wind in the extended outlook.
	HeaderExtWind = "Wind"
	// HeaderExtDir is the column header for wind direction in the extended outlook.
	HeaderExtDir = "Dir"
	// HeaderExtThermal is the column header for thermal rating in the extended outlook.
	HeaderExtThermal = "Thermal"
	// HeaderExtRain is the column header for rain in the extended outlook.
	HeaderExtRain = "Rain"
	// HeaderExtScore is the column header for overall score in the extended outlook.
	HeaderExtScore = "Score"
	// ExtendedOutlookTitle is the heading text for the extended outlook section.
	ExtendedOutlookTitle = "EXTENDED OUTLOOK"
)

// Format strings and labels.
const (
	// ForecastTitle is the formatted title line for the paragliding forecast.
	ForecastTitle = "🪂 PARAGLIDING FORECAST — %s"
	// BestWindowLabel is the label used to show the best flying window.
	BestWindowLabel = "🏆 Best Window: %s"
	// CloudbaseLabel is the format string for summarising cloudbase, CAPE, and freezing level.
	CloudbaseLabel = "Cloudbase: ~%s | CAPE: %.0f J/kg | Freezing: %.0fft"
	// OrographicLabel is the format string for describing orographic lift conditions.
	OrographicLabel = "Orographic: %s"
	// XCLabel is the format string for describing cross-country potential.
	XCLabel = "XC Potential: %s %s"
)
