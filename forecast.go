package pgforecast

import (
	"fmt"
	"sort"
	"time"
)

// GenerateForecast fetches weather and computes metrics for a site.
func GenerateForecast(site Site, opts ForecastOptions) (*SiteForecast, error) {
	loc, err := time.LoadLocation(opts.Timezone)
	if err != nil {
		loc = time.UTC
	}

	tc := opts.Tuning
	if tc == nil {
		tc = DefaultTuningConfig()
	}

	hourlyData, err := FetchWeather(site, opts)
	if err != nil {
		return nil, fmt.Errorf("fetching weather for %s: %w", site.Name, err)
	}

	now := time.Now().In(loc)
	forecast := &SiteForecast{
		Site:      site,
		Generated: now,
		Units:     opts.Units,
	}

	// Group hourly data by day (in local timezone)
	dayMap := make(map[string][]int)
	var dayOrder []string
	for i, h := range hourlyData {
		lt := h.Time.In(loc)
		key := lt.Format("2006-01-02")
		if _, exists := dayMap[key]; !exists {
			dayOrder = append(dayOrder, key)
		}
		dayMap[key] = append(dayMap[key], i)
	}

	detailedDays := opts.DetailedDays
	if detailedDays <= 0 {
		detailedDays = 3
	}

	bestScore := 0
	bestWindow := ""

	for dayIdx, dateStr := range dayOrder {
		indices := dayMap[dateStr]
		date, _ := time.ParseInLocation("2006-01-02", dateStr, loc)

		if dayIdx < detailedDays {
			df := DayForecast{Date: date}
			var dayMetrics []HourlyMetrics
			for _, idx := range indices {
				h := &hourlyData[idx]
				lt := h.Time.In(loc)
				hour := lt.Hour()
				if hour >= 8 && hour <= 18 {
					m := ComputeHourlyMetrics(h, site, tc)
					m.Time = lt
					df.Hours = append(df.Hours, m)
					dayMetrics = append(dayMetrics, m)
					if m.FlyabilityScore > bestScore {
						bestScore = m.FlyabilityScore
						bestWindow = lt.Format("Mon 15:04")
					}
				}
			}
			df.Summary = summarizeDay(date, dayMetrics, tc)
			forecast.DetailedDays = append(forecast.DetailedDays, df)
		} else {
			var dayMetrics []HourlyMetrics
			for _, idx := range indices {
				h := &hourlyData[idx]
				lt := h.Time.In(loc)
				hour := lt.Hour()
				if hour >= 8 && hour <= 18 {
					m := ComputeHourlyMetrics(h, site, tc)
					m.Time = lt
					dayMetrics = append(dayMetrics, m)
				}
			}
			if len(dayMetrics) > 0 {
				summary := summarizeDay(date, dayMetrics, tc)
				forecast.ExtendedDays = append(forecast.ExtendedDays, summary)
			}
		}
	}

	if bestScore >= 3 {
		forecast.BestWindow = bestWindow
	}

	return forecast, nil
}

func summarizeDay(date time.Time, metrics []HourlyMetrics, tc *TuningConfig) DaySummary {
	if len(metrics) == 0 {
		return DaySummary{Date: date}
	}

	var totalWind, totalDir, maxGusts, maxPrecip, maxCAPE float64
	bestThermal := "None"
	thermalOrder := map[string]int{"None": 0, "Weak": 1, "Moderate": 2, "Strong": 3, "Extreme": 4}
	totalCloudbase := 0

	// Collect all scores for top-3 averaging
	var scores []int

	for _, m := range metrics {
		totalWind += m.WindSpeed
		totalDir += m.WindDirection
		if m.WindGusts > maxGusts {
			maxGusts = m.WindGusts
		}
		if m.PrecipProb > maxPrecip {
			maxPrecip = m.PrecipProb
		}
		if m.CAPE > maxCAPE {
			maxCAPE = m.CAPE
		}
		scores = append(scores, m.FlyabilityScore)
		if thermalOrder[m.ThermalRating] > thermalOrder[bestThermal] {
			bestThermal = m.ThermalRating
		}
		totalCloudbase += m.CloudbaseFt
	}

	// Day score = average of top 3 hours (not single best)
	sort.Sort(sort.Reverse(sort.IntSlice(scores)))
	topN := 3
	if len(scores) < topN {
		topN = len(scores)
	}
	sum := 0
	for i := 0; i < topN; i++ {
		sum += scores[i]
	}
	dayScore := (sum + topN/2) / topN // rounded integer average

	n := float64(len(metrics))
	avgDir := totalDir / n
	return DaySummary{
		Date:          date,
		AvgWindSpeed:  totalWind / n,
		AvgWindDir:    avgDir,
		WindDirStr:    DegreesToCompass(avgDir),
		MaxGusts:      maxGusts,
		ThermalRating: bestThermal,
		MaxPrecipProb: maxPrecip,
		AvgCloudbase:  totalCloudbase / len(metrics),
		BestScore:     dayScore,
		XCPotential:   CalcXCPotential(maxCAPE, totalCloudbase/len(metrics), totalWind/n, bestThermal, tc),
	}
}
