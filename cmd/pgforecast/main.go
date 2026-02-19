package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/matt-FFFFFF/pgforecast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	sitesFile  string
	siteName   string
	latStr     string
	lonStr     string
	nameStr    string
	aspect     int
	windRange  string
	jsonOutput bool
	outputFmt  string
	units      string
	days       int
	timezone   string
	model      string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pgforecast",
		Short: "Paragliding forecast tool",
		RunE:  run,
	}

	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&cfgFile, "config", "c", "", "Path to config YAML file")
	pf.StringVarP(&units, "units", "u", "mph", "Wind units: mph/kph/knots/ms")
	pf.StringVar(&timezone, "timezone", "Europe/London", "Timezone")
	pf.StringVar(&model, "model", "auto", "Weather model (auto/gfs/ecmwf/icon)")

	f := rootCmd.Flags()
	f.StringVarP(&sitesFile, "sites", "s", "", "Path to sites YAML file")
	f.StringVar(&siteName, "site", "", "Filter to specific site name")
	f.StringVar(&latStr, "lat", "", "Latitude for ad-hoc site")
	f.StringVar(&lonStr, "lon", "", "Longitude for ad-hoc site")
	f.StringVar(&nameStr, "name", "", "Name for ad-hoc site")
	f.IntVar(&aspect, "aspect", 0, "Site aspect in degrees")
	f.StringVar(&windRange, "wind-range", "", "Wind direction range e.g. 210-260")
	f.BoolVar(&jsonOutput, "json", false, "Output as JSON")
	f.StringVar(&outputFmt, "output", "text", "Output format: text or json")
	f.IntVar(&days, "days", 3, "Number of detailed forecast days")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	tc, err := loadTuningConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	opts := pgforecast.ForecastOptions{
		Units:        units,
		DetailedDays: days,
		Timezone:     timezone,
		Model:        model,
		OutputFormat: "text",
		Tuning:       tc,
	}
	if jsonOutput || outputFmt == "json" {
		opts.OutputFormat = "json"
	}

	var sites []pgforecast.Site

	if latStr != "" || lonStr != "" {
		lat, _ := strconv.ParseFloat(latStr, 64)
		lon, _ := strconv.ParseFloat(lonStr, 64)
		s := pgforecast.Site{
			Name:    nameStr,
			Lat:     lat,
			Lon:     lon,
			Aspect:  aspect,
			BestDir: aspect,
		}
		if s.Name == "" {
			s.Name = "Custom"
		}
		if windRange != "" {
			parts := strings.Split(windRange, "-")
			if len(parts) == 2 {
				s.WindMin, _ = strconv.Atoi(parts[0])
				s.WindMax, _ = strconv.Atoi(parts[1])
			}
		}
		sites = []pgforecast.Site{s}
	} else if sitesFile != "" {
		sites, err = pgforecast.LoadSites(sitesFile)
		if err != nil {
			return err
		}
		if siteName != "" {
			s, ok := pgforecast.FilterSite(sites, siteName)
			if !ok {
				return fmt.Errorf("site %q not found", siteName)
			}
			sites = []pgforecast.Site{s}
		}
	} else {
		return fmt.Errorf("specify --sites or --lat/--lon")
	}

	for _, site := range sites {
		forecast, err := pgforecast.GenerateForecast(site, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}
		if opts.OutputFormat == "json" {
			pgforecast.FormatJSON(os.Stdout, forecast)
		} else {
			pgforecast.FormatText(os.Stdout, forecast, tc)
		}
	}
	return nil
}

// loadTuningConfig loads tuning config from file, env vars, merging with defaults.
func loadTuningConfig(configPath string) (*pgforecast.TuningConfig, error) {
	v := viper.New()
	v.SetEnvPrefix("PGF")
	v.AutomaticEnv()

	def := pgforecast.DefaultTuningConfig()
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

	tc := pgforecast.DefaultTuningConfig()
	if err := v.Unmarshal(tc); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}
	return tc, nil
}
