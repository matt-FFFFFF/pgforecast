package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/matt-FFFFFF/pgforecast"
	"github.com/spf13/cobra"
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
	tc, err := pgforecast.LoadTuningConfig(cfgFile)
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
