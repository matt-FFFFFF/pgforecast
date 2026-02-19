package pgforecast

import "time"

// Site represents a paragliding launch site.
type Site struct {
	Name      string  `yaml:"name" json:"name"`
	Lat       float64 `yaml:"lat" json:"lat"`
	Lon       float64 `yaml:"lon" json:"lon"`
	Elevation int     `yaml:"elevation" json:"elevation"`
	WindMin   int     `yaml:"wind_min" json:"wind_min"`
	WindMax   int     `yaml:"wind_max" json:"wind_max"`
	BestDir   int     `yaml:"best_dir" json:"best_dir"`
	Aspect    int     `yaml:"aspect" json:"aspect"`
}

// SitesConfig is the top-level YAML structure.
type SitesConfig struct {
	Sites []Site `yaml:"sites"`
}

// PressureLevel holds data for one pressure level at one hour.
type PressureLevel struct {
	Pressure           int     `json:"pressure_hpa"`
	WindSpeed          float64 `json:"wind_speed"`
	WindDirection      float64 `json:"wind_direction"`
	Temperature        float64 `json:"temperature"`
	GeopotentialHeight float64 `json:"geopotential_height"`
}

// HourlyData holds all weather data for one hour.
type HourlyData struct {
	Time                     time.Time       `json:"time"`
	Temperature              float64         `json:"temperature_2m"`
	RelativeHumidity         float64         `json:"relative_humidity_2m"`
	DewPoint                 float64         `json:"dew_point_2m"`
	WindSpeed                float64         `json:"wind_speed_10m"`
	WindDirection            float64         `json:"wind_direction_10m"`
	WindGusts                float64         `json:"wind_gusts_10m"`
	CloudCover               float64         `json:"cloud_cover"`
	CloudCoverLow            float64         `json:"cloud_cover_low"`
	CloudCoverMid            float64         `json:"cloud_cover_mid"`
	CloudCoverHigh           float64         `json:"cloud_cover_high"`
	CAPE                     float64         `json:"cape"`
	ShortwaveRadiation       float64         `json:"shortwave_radiation"`
	Precipitation            float64         `json:"precipitation"`
	PrecipitationProbability float64         `json:"precipitation_probability"`
	FreezingLevelHeight      float64         `json:"freezing_level_height"`
	IsDay                    int             `json:"is_day"`
	WeatherCode              int             `json:"weather_code"`
	PressureMSL              float64         `json:"pressure_msl"`
	Visibility               float64         `json:"visibility"`
	PressureLevels           []PressureLevel `json:"pressure_levels"`
}

// HourlyMetrics holds computed paragliding metrics for one hour.
type HourlyMetrics struct {
	Time             time.Time       `json:"time"`
	WindSpeed        float64         `json:"wind_speed"`
	WindDirection    float64         `json:"wind_direction"`
	WindDirStr       string          `json:"wind_dir_str"`
	WindGusts        float64         `json:"wind_gusts"`
	WindGradient     string          `json:"wind_gradient"` // Low/Medium/High
	WindGradientDiff float64         `json:"wind_gradient_diff"`
	ThermalRating    string          `json:"thermal_rating"` // None/Weak/Moderate/Strong/Extreme
	CAPE             float64         `json:"cape"`
	CAPERating       string          `json:"cape_rating"`
	CloudbaseFt      int             `json:"cloudbase_ft"`
	CloudCover       float64         `json:"cloud_cover"`
	Precipitation    float64         `json:"precipitation"`
	PrecipProb       float64         `json:"precip_probability"`
	OrographicLift   string          `json:"orographic_lift"`  // None/Weak/Moderate/Strong
	FlyabilityScore  int             `json:"flyability_score"` // 1-5
	XCPotential      string          `json:"xc_potential"`     // Low/Medium/High/Epic
	FreezingLevel    float64         `json:"freezing_level_ft"`
	IsDay            bool            `json:"is_day"`
	PressureLevels   []PressureLevel `json:"pressure_levels"`
}

// DaySummary holds aggregated metrics for extended outlook days.
type DaySummary struct {
	Date          time.Time `json:"date"`
	AvgWindSpeed  float64   `json:"avg_wind_speed"`
	AvgWindDir    float64   `json:"avg_wind_direction"`
	WindDirStr    string    `json:"wind_dir_str"`
	MaxGusts      float64   `json:"max_gusts"`
	ThermalRating string    `json:"thermal_rating"`
	MaxPrecipProb float64   `json:"max_precip_prob"`
	AvgCloudbase  int       `json:"avg_cloudbase_ft"`
	BestScore     int       `json:"best_score"`
	XCPotential   string    `json:"xc_potential"`
}

// SiteForecast holds the complete forecast for one site.
type SiteForecast struct {
	Site         Site          `json:"site"`
	Generated    time.Time     `json:"generated"`
	Units        string        `json:"units"`
	DetailedDays []DayForecast `json:"detailed_days"`
	ExtendedDays []DaySummary  `json:"extended_days"`
	BestWindow   string        `json:"best_window"`
}

// DayForecast holds hourly metrics for one day.
type DayForecast struct {
	Date    time.Time       `json:"date"`
	Hours   []HourlyMetrics `json:"hours"`
	Summary DaySummary      `json:"summary"`
}

// ForecastOptions holds runtime options.
type ForecastOptions struct {
	Units        string // mph, kph, knots, ms
	DetailedDays int
	Timezone     string
	Model        string
	OutputFormat string // text, json
	Tuning       *TuningConfig
}
