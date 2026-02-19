package pgforecast

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadSites loads sites from a YAML file.
func LoadSites(path string) ([]Site, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading sites file: %w", err)
	}
	var cfg SitesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing sites YAML: %w", err)
	}
	return cfg.Sites, nil
}

// FilterSite returns a single site by name, case-insensitive prefix match.
func FilterSite(sites []Site, name string) (Site, bool) {
	for _, s := range sites {
		if equalsCI(s.Name, name) {
			return s, true
		}
	}
	// prefix match
	for _, s := range sites {
		if hasPrefixCI(s.Name, name) {
			return s, true
		}
	}
	return Site{}, false
}

func equalsCI(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' { ca += 32 }
		if cb >= 'A' && cb <= 'Z' { cb += 32 }
		if ca != cb { return false }
	}
	return true
}

func hasPrefixCI(s, prefix string) bool {
	if len(s) < len(prefix) { return false }
	return equalsCI(s[:len(prefix)], prefix)
}
