package pgforecast

// TuningConfig holds all tunable parameters for the forecast engine.
type TuningConfig struct {
	Wind struct {
		IdealMin            float64 `mapstructure:"ideal_min" yaml:"ideal_min" json:"ideal_min"`
		IdealMax            float64 `mapstructure:"ideal_max" yaml:"ideal_max" json:"ideal_max"`
		AcceptableMin       float64 `mapstructure:"acceptable_min" yaml:"acceptable_min" json:"acceptable_min"`
		AcceptableMax       float64 `mapstructure:"acceptable_max" yaml:"acceptable_max" json:"acceptable_max"`
		DangerousMax        float64 `mapstructure:"dangerous_max" yaml:"dangerous_max" json:"dangerous_max"`
		MaxGustFactor       float64 `mapstructure:"max_gust_factor" yaml:"max_gust_factor" json:"max_gust_factor"`
		DangerousGustFactor float64 `mapstructure:"dangerous_gust_factor" yaml:"dangerous_gust_factor" json:"dangerous_gust_factor"`
	} `mapstructure:"wind" yaml:"wind" json:"wind"`

	Gradient struct {
		LowThreshold  float64 `mapstructure:"low_threshold" yaml:"low_threshold" json:"low_threshold"`
		HighThreshold float64 `mapstructure:"high_threshold" yaml:"high_threshold" json:"high_threshold"`
		HighPenalty   float64 `mapstructure:"high_penalty" yaml:"high_penalty" json:"high_penalty"`
		MediumPenalty float64 `mapstructure:"medium_penalty" yaml:"medium_penalty" json:"medium_penalty"`
	} `mapstructure:"gradient" yaml:"gradient" json:"gradient"`

	Thermal struct {
		CAPEWeak       float64 `mapstructure:"cape_weak" yaml:"cape_weak" json:"cape_weak"`
		CAPEModerate   float64 `mapstructure:"cape_moderate" yaml:"cape_moderate" json:"cape_moderate"`
		CAPEStrong     float64 `mapstructure:"cape_strong" yaml:"cape_strong" json:"cape_strong"`
		CAPEExtreme    float64 `mapstructure:"cape_extreme" yaml:"cape_extreme" json:"cape_extreme"`
		LapseRateBonus float64 `mapstructure:"lapse_rate_bonus" yaml:"lapse_rate_bonus" json:"lapse_rate_bonus"`
	} `mapstructure:"thermal" yaml:"thermal" json:"thermal"`

	Orographic struct {
		MinWindSpeed  float64 `mapstructure:"min_wind_speed" yaml:"min_wind_speed" json:"min_wind_speed"`
		StrongAngle   float64 `mapstructure:"strong_angle" yaml:"strong_angle" json:"strong_angle"`
		ModerateAngle float64 `mapstructure:"moderate_angle" yaml:"moderate_angle" json:"moderate_angle"`
		WeakAngle     float64 `mapstructure:"weak_angle" yaml:"weak_angle" json:"weak_angle"`
	} `mapstructure:"orographic" yaml:"orographic" json:"orographic"`

	Cloudbase struct {
		MinRealisticFt int `mapstructure:"min_realistic_ft" yaml:"min_realistic_ft" json:"min_realistic_ft"`
	} `mapstructure:"cloudbase" yaml:"cloudbase" json:"cloudbase"`

	Scoring struct {
		BaseScore           float64 `mapstructure:"base_score" yaml:"base_score" json:"base_score"`
		WindIdealBonus      float64 `mapstructure:"wind_ideal_bonus" yaml:"wind_ideal_bonus" json:"wind_ideal_bonus"`
		WindAcceptableBonus float64 `mapstructure:"wind_acceptable_bonus" yaml:"wind_acceptable_bonus" json:"wind_acceptable_bonus"`
		WindDangerPenalty   float64 `mapstructure:"wind_danger_penalty" yaml:"wind_danger_penalty" json:"wind_danger_penalty"`
		WindHighPenalty     float64 `mapstructure:"wind_high_penalty" yaml:"wind_high_penalty" json:"wind_high_penalty"`
		DirOnBonus          float64 `mapstructure:"dir_on_bonus" yaml:"dir_on_bonus" json:"dir_on_bonus"`
		DirOffPenalty       float64 `mapstructure:"dir_off_penalty" yaml:"dir_off_penalty" json:"dir_off_penalty"`
		GustHighPenalty     float64 `mapstructure:"gust_high_penalty" yaml:"gust_high_penalty" json:"gust_high_penalty"`
		GustMedPenalty      float64 `mapstructure:"gust_med_penalty" yaml:"gust_med_penalty" json:"gust_med_penalty"`
		RainPenalty         float64 `mapstructure:"rain_penalty" yaml:"rain_penalty" json:"rain_penalty"`
		RainProbPenalty     float64 `mapstructure:"rain_prob_penalty" yaml:"rain_prob_penalty" json:"rain_prob_penalty"`
		GradientHighPenalty float64 `mapstructure:"gradient_high_penalty" yaml:"gradient_high_penalty" json:"gradient_high_penalty"`
		GradientMedPenalty  float64 `mapstructure:"gradient_med_penalty" yaml:"gradient_med_penalty" json:"gradient_med_penalty"`
		CAPEBonus           float64 `mapstructure:"cape_bonus" yaml:"cape_bonus" json:"cape_bonus"`
		ThermalStrongBonus  float64 `mapstructure:"thermal_strong_bonus" yaml:"thermal_strong_bonus" json:"thermal_strong_bonus"`
	} `mapstructure:"scoring" yaml:"scoring" json:"scoring"`

	XC struct {
		MinCloudbaseFt  int     `mapstructure:"min_cloudbase_ft" yaml:"min_cloudbase_ft" json:"min_cloudbase_ft"`
		GoodCloudbaseFt int     `mapstructure:"good_cloudbase_ft" yaml:"good_cloudbase_ft" json:"good_cloudbase_ft"`
		MaxWindSpeed    float64 `mapstructure:"max_wind_speed" yaml:"max_wind_speed" json:"max_wind_speed"`
		MinWindSpeed    float64 `mapstructure:"min_wind_speed" yaml:"min_wind_speed" json:"min_wind_speed"`
		EpicThreshold   int     `mapstructure:"epic_threshold" yaml:"epic_threshold" json:"epic_threshold"`
		HighThreshold   int     `mapstructure:"high_threshold" yaml:"high_threshold" json:"high_threshold"`
		MediumThreshold int     `mapstructure:"medium_threshold" yaml:"medium_threshold" json:"medium_threshold"`
	} `mapstructure:"xc" yaml:"xc" json:"xc"`
}

// DefaultTuningConfig returns the default tuning configuration.
func DefaultTuningConfig() *TuningConfig {
	tc := &TuningConfig{}

	tc.Wind.IdealMin = 8
	tc.Wind.IdealMax = 18
	tc.Wind.AcceptableMin = 5
	tc.Wind.AcceptableMax = 22
	tc.Wind.DangerousMax = 25
	tc.Wind.MaxGustFactor = 1.5
	tc.Wind.DangerousGustFactor = 2.0

	tc.Gradient.LowThreshold = 10
	tc.Gradient.HighThreshold = 20
	tc.Gradient.HighPenalty = -2.0
	tc.Gradient.MediumPenalty = -1.0

	tc.Thermal.CAPEWeak = 100
	tc.Thermal.CAPEModerate = 300
	tc.Thermal.CAPEStrong = 1000
	tc.Thermal.CAPEExtreme = 2500
	tc.Thermal.LapseRateBonus = 8.0

	tc.Orographic.MinWindSpeed = 8
	tc.Orographic.StrongAngle = 15
	tc.Orographic.ModerateAngle = 30
	tc.Orographic.WeakAngle = 45

	tc.Cloudbase.MinRealisticFt = 200

	tc.Scoring.BaseScore = 2.5
	tc.Scoring.WindIdealBonus = 1.0
	tc.Scoring.WindAcceptableBonus = 0.5
	tc.Scoring.WindDangerPenalty = -2.0
	tc.Scoring.WindHighPenalty = -1.0
	tc.Scoring.DirOnBonus = 1.5
	tc.Scoring.DirOffPenalty = -2.0
	tc.Scoring.GustHighPenalty = -1.5
	tc.Scoring.GustMedPenalty = -0.5
	tc.Scoring.RainPenalty = -2.5
	tc.Scoring.RainProbPenalty = -0.5
	tc.Scoring.GradientHighPenalty = -1.5
	tc.Scoring.GradientMedPenalty = -0.5
	tc.Scoring.CAPEBonus = 0.5
	tc.Scoring.ThermalStrongBonus = 0.5

	tc.XC.MinCloudbaseFt = 3000
	tc.XC.GoodCloudbaseFt = 4000
	tc.XC.MaxWindSpeed = 20
	tc.XC.MinWindSpeed = 8
	tc.XC.EpicThreshold = 7
	tc.XC.HighThreshold = 5
	tc.XC.MediumThreshold = 3

	return tc
}
