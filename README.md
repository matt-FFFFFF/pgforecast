# 🪂 pgforecast

Paragliding forecast tool and Go library. Fetches weather data from [Open-Meteo](https://open-meteo.com/) and computes paragliding-specific metrics including wind gradients, thermal potential, orographic lift, cloudbase estimates, and flyability scores.

## Features

- **Wind gradient analysis** — surface to 850hPa (~1500m), the levels that matter for paragliding
- **Thermal potential** — CAPE-based rating with lapse rate enhancement
- **Cloudbase estimates** — from temperature/dewpoint spread, with fog detection
- **Orographic lift** — wind direction vs site aspect matching
- **Flyability score** (1-5⭐) — composite rating factoring wind, direction, gusts, gradient, rain
- **XC potential** — cross-country day rating (Low → Epic)
- **Configurable scoring** — all thresholds tunable via Viper config (YAML/env/flags)
- **Multi-site support** — YAML site database with 26 pre-configured Wessex HGPG sites
- **Detailed + extended forecasts** — hourly for days 1-3, daily summary for days 4-16

## Install

```bash
go install github.com/matt-FFFFFF/pgforecast/cmd/pgforecast@latest
```

## Usage

```bash
# All configured sites
pgforecast --sites sites.yaml

# Single site
pgforecast --sites sites.yaml --site Ringstead

# Ad-hoc location
pgforecast --lat 50.64 --lon -2.34 --name "My Spot" --aspect 225 --wind-range 210-260

# JSON output
pgforecast --sites sites.yaml --site Ringstead --json

# Different units
pgforecast --sites sites.yaml --units kph
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--sites` | `-s` | | Path to sites YAML file |
| `--site` | | | Filter to a specific site |
| `--lat` | | | Latitude (ad-hoc site) |
| `--lon` | | | Longitude (ad-hoc site) |
| `--name` | | Custom | Name for ad-hoc site |
| `--aspect` | | | Site aspect in degrees |
| `--wind-range` | | | Wind direction range, e.g. `210-260` |
| `--json` | | false | Output as JSON |
| `--units` | `-u` | mph | Wind units: mph, kph, knots, ms |
| `--days` | | 3 | Number of detailed forecast days |
| `--timezone` | `--tz` | Europe/London | Display timezone |
| `--config` | `-c` | | Path to config YAML for tuning |

## Sites Configuration

Sites are defined in a YAML file:

```yaml
sites:
  - name: Ringstead
    lat: 50.6403
    lon: -2.3425
    elevation: 147
    wind_min: 210
    wind_max: 260
    best_dir: 225
    aspect: 225
```

The included `sites.yaml` has 26 sites from the [Wessex HGPG](http://www.wessexhgpg.org.uk/) club plus Beer Head, Eype, and Cogden.

## Tuning

All scoring parameters are configurable. Copy `pgforecast.example.yaml` and adjust:

```bash
pgforecast --config my-config.yaml --sites sites.yaml
```

Or set via environment variables with `PGF_` prefix:

```bash
PGF_SCORING_BASE_SCORE=3.0 pgforecast --sites sites.yaml
```

## As a Library

```go
import "github.com/matt-FFFFFF/pgforecast"

sites, _ := pgforecast.LoadSites("sites.yaml")
tc := pgforecast.DefaultTuningConfig()
opts := pgforecast.ForecastOptions{
    Units:        "mph",
    DetailedDays: 3,
    Timezone:     "Europe/London",
    Tuning:       tc,
}

forecast, _ := pgforecast.GenerateForecast(sites[0], opts)
pgforecast.FormatText(os.Stdout, forecast)
```

## Data Source

All weather data from [Open-Meteo](https://open-meteo.com/) — free, no API key required. Uses GFS model data with surface parameters and pressure level winds/temperatures at 1000, 950, 925, 900, 850, and 700 hPa.

## License

MIT
