package pgforecast

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// TuningConfig holds all tunable parameters for the forecast engine.
type TuningConfig struct {
	Wind struct {
		IdealMin            float64 `mapstructure:"ideal_min" yaml:"ideal_min"`
		IdealMax            float64 `mapstructure:"ideal_max" yaml:"ideal_max"`
		AcceptableMin       float64 `mapstructure:"acceptable_min" yaml:"acceptable_min"`
		AcceptableMax       float64 `mapstructure:"acceptable_max" yaml:"acceptable_max"`
		DangerousMax        float64 `mapstructure:"dangerous_max" yaml:"dangerous_max"`
		MaxGustFactor       float64 `mapstructure:"max_gust_factor" yaml:"max_gust_factor"`
		DangerousGustFactor float64 `mapstructure:"dangerous_gust_factor" yaml:"dangerous_gust_factor"`
	} `mapstructure:"wind" yaml:"wind"`

	Gradient struct {
		LowThreshold  float64 `mapstructure:"low_threshold" yaml:"low_threshold"`
		HighThreshold float64 `mapstructure:"high_threshold" yaml:"high_threshold"`
		HighPenalty   float64 `mapstructure:"high_penalty" yaml:"high_penalty"`
		MediumPenalty float64 `mapstructure:"medium_penalty" yaml:"medium_penalty"`
	} `mapstructure:"gradient" yaml:"gradient"`

	Thermal struct {
		CAPEWeak       float64 `mapstructure:"cape_weak" yaml:"cape_weak"`
		CAPEModerate   float64 `mapstructure:"cape_moderate" yaml:"cape_moderate"`
		CAPEStrong     float64 `mapstructure:"cape_strong" yaml:"cape_strong"`
		CAPEExtreme    float64 `mapstructure:"cape_extreme" yaml:"cape_extreme"`
		LapseRateBonus float64 `mapstructure:"lapse_rate_bonus" yaml:"lapse_rate_bonus"`
	} `mapstructure:"thermal" yaml:"thermal"`

	Orographic struct {
		MinWindSpeed  float64 `mapstructure:"min_wind_speed" yaml:"min_wind_speed"`
		StrongAngle   float64 `mapstructure:"strong_angle" yaml:"strong_angle"`
		ModerateAngle float64 `mapstructure:"moderate_angle" yaml:"moderate_angle"`
		WeakAngle     float64 `mapstructure:"weak_angle" yaml:"weak_angle"`
	} `mapstructure:"orographic" yaml:"orographic"`

	Cloudbase struct {
		MinRealisticFt int `mapstructure:"min_realistic_ft" yaml:"min_realistic_ft"`
	} `mapstructure:"cloudbase" yaml:"cloudbase"`

	Scoring struct {
		BaseScore           float64 `mapstructure:"base_score" yaml:"base_score"`
		WindIdealBonus      float64 `mapstructure:"wind_ideal_bonus" yaml:"wind_ideal_bonus"`
		WindAcceptableBonus float64 `mapstructure:"wind_acceptable_bonus" yaml:"wind_acceptable_bonus"`
		WindDangerPenalty   float64 `mapstructure:"wind_danger_penalty" yaml:"wind_danger_penalty"`
		WindHighPenalty     float64 `mapstructure:"wind_high_penalty" yaml:"wind_high_penalty"`
		DirOnBonus          float64 `mapstructure:"dir_on_bonus" yaml:"dir_on_bonus"`
		DirOffPenalty       float64 `mapstructure:"dir_off_penalty" yaml:"dir_off_penalty"`
		GustHighPenalty     float64 `mapstructure:"gust_high_penalty" yaml:"gust_high_penalty"`
		GustMedPenalty      float64 `mapstructure:"gust_med_penalty" yaml:"gust_med_penalty"`
		RainPenalty         float64 `mapstructure:"rain_penalty" yaml:"rain_penalty"`
		RainProbPenalty     float64 `mapstructure:"rain_prob_penalty" yaml:"rain_prob_penalty"`
		GradientHighPenalty float64 `mapstructure:"gradient_high_penalty" yaml:"gradient_high_penalty"`
		GradientMedPenalty  float64 `mapstructure:"gradient_med_penalty" yaml:"gradient_med_penalty"`
		CAPEBonus           float64 `mapstructure:"cape_bonus" yaml:"cape_bonus"`
		ThermalStrongBonus  float64 `mapstructure:"thermal_strong_bonus" yaml:"thermal_strong_bonus"`
	} `mapstructure:"scoring" yaml:"scoring"`

	XC struct {
		MinCloudbaseFt  int     `mapstructure:"min_cloudbase_ft" yaml:"min_cloudbase_ft"`
		GoodCloudbaseFt int     `mapstructure:"good_cloudbase_ft" yaml:"good_cloudbase_ft"`
		MaxWindSpeed    float64 `mapstructure:"max_wind_speed" yaml:"max_wind_speed"`
		MinWindSpeed    float64 `mapstructure:"min_wind_speed" yaml:"min_wind_speed"`
		EpicThreshold   int     `mapstructure:"epic_threshold" yaml:"epic_threshold"`
		HighThreshold   int     `mapstructure:"high_threshold" yaml:"high_threshold"`
		MediumThreshold int     `mapstructure:"medium_threshold" yaml:"medium_threshold"`
	} `mapstructure:"xc" yaml:"xc"`
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

// LoadTuningConfig loads tuning config from file, env vars, merging with defaults.
func LoadTuningConfig(configPath string) (*TuningConfig, error) {
	v := viper.New()
	v.SetEnvPrefix("PGF")
	v.AutomaticEnv()

	// Set defaults from DefaultTuningConfig
	def := DefaultTuningConfig()
	v.SetDefault("wind.ideal_min", def.Wind.IdealMin)
	v.SetDefault("wind.ideal_max", def.Wind.IdealMax)
	v.SetDefault("wind.acceptable_min", def.Wind.AcceptableMin)
	v.SetDefault("wind.acceptable_max", def.Wind.AcceptableMax)
	v.SetDefault("wind.dangerous_max", def.Wind.DangerousMax)
	v.SetDefault("wind.max_gust_factor", def.Wind.MaxGustFactor)
	v.SetDefault("wind.dangerous_gust_factor", def.Wind.DangerousGustFactor)
	v.SetDefault("gradient.low_threshold", def.Gradient.LowThreshold)
	v.SetDefault("gradient.high_threshold", def.Gradient.HighThreshold)
	v.SetDefault("gradient.high_penalty", def.Gradient.HighPenalty)
	v.SetDefault("gradient.medium_penalty", def.Gradient.MediumPenalty)
	v.SetDefault("thermal.cape_weak", def.Thermal.CAPEWeak)
	v.SetDefault("thermal.cape_moderate", def.Thermal.CAPEModerate)
	v.SetDefault("thermal.cape_strong", def.Thermal.CAPEStrong)
	v.SetDefault("thermal.cape_extreme", def.Thermal.CAPEExtreme)
	v.SetDefault("thermal.lapse_rate_bonus", def.Thermal.LapseRateBonus)
	v.SetDefault("orographic.min_wind_speed", def.Orographic.MinWindSpeed)
	v.SetDefault("orographic.strong_angle", def.Orographic.StrongAngle)
	v.SetDefault("orographic.moderate_angle", def.Orographic.ModerateAngle)
	v.SetDefault("orographic.weak_angle", def.Orographic.WeakAngle)
	v.SetDefault("cloudbase.min_realistic_ft", def.Cloudbase.MinRealisticFt)
	v.SetDefault("scoring.base_score", def.Scoring.BaseScore)
	v.SetDefault("scoring.wind_ideal_bonus", def.Scoring.WindIdealBonus)
	v.SetDefault("scoring.wind_acceptable_bonus", def.Scoring.WindAcceptableBonus)
	v.SetDefault("scoring.wind_danger_penalty", def.Scoring.WindDangerPenalty)
	v.SetDefault("scoring.wind_high_penalty", def.Scoring.WindHighPenalty)
	v.SetDefault("scoring.dir_on_bonus", def.Scoring.DirOnBonus)
	v.SetDefault("scoring.dir_off_penalty", def.Scoring.DirOffPenalty)
	v.SetDefault("scoring.gust_high_penalty", def.Scoring.GustHighPenalty)
	v.SetDefault("scoring.gust_med_penalty", def.Scoring.GustMedPenalty)
	v.SetDefault("scoring.rain_penalty", def.Scoring.RainPenalty)
	v.SetDefault("scoring.rain_prob_penalty", def.Scoring.RainProbPenalty)
	v.SetDefault("scoring.gradient_high_penalty", def.Scoring.GradientHighPenalty)
	v.SetDefault("scoring.gradient_med_penalty", def.Scoring.GradientMedPenalty)
	v.SetDefault("scoring.cape_bonus", def.Scoring.CAPEBonus)
	v.SetDefault("scoring.thermal_strong_bonus", def.Scoring.ThermalStrongBonus)
	v.SetDefault("xc.min_cloudbase_ft", def.XC.MinCloudbaseFt)
	v.SetDefault("xc.good_cloudbase_ft", def.XC.GoodCloudbaseFt)
	v.SetDefault("xc.max_wind_speed", def.XC.MaxWindSpeed)
	v.SetDefault("xc.min_wind_speed", def.XC.MinWindSpeed)
	v.SetDefault("xc.epic_threshold", def.XC.EpicThreshold)
	v.SetDefault("xc.high_threshold", def.XC.HighThreshold)
	v.SetDefault("xc.medium_threshold", def.XC.MediumThreshold)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("pgforecast")
		v.AddConfigPath(".")
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "pgforecast"))
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && configPath != "" {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	tc := DefaultTuningConfig()
	if err := v.Unmarshal(tc); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}
	return tc, nil
}
