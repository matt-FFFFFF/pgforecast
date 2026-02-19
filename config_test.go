package pgforecast

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSites(t *testing.T) {
	sites, err := LoadSites("sites.yaml")
	if err != nil {
		t.Fatalf("LoadSites: %v", err)
	}
	if len(sites) == 0 {
		t.Fatal("no sites loaded")
	}

	// Check a known site
	found := false
	for _, s := range sites {
		if s.Name == "Ringstead" {
			found = true
			if s.Lat == 0 || s.Lon == 0 {
				t.Error("Ringstead has zero lat/lon")
			}
			if s.Elevation == 0 {
				t.Error("Ringstead has zero elevation")
			}
			if s.WindMin == 0 && s.WindMax == 0 {
				t.Error("Ringstead has no wind range")
			}
		}
	}
	if !found {
		t.Error("Ringstead not found in sites")
	}
}

func TestFilterSite(t *testing.T) {
	sites := []Site{
		{Name: "Ringstead"},
		{Name: "Bell Hill"},
		{Name: "Swallowcliffe"},
	}

	s, ok := FilterSite(sites, "Bell Hill")
	if !ok {
		t.Fatal("Bell Hill not found")
	}
	if s.Name != "Bell Hill" {
		t.Errorf("got %q", s.Name)
	}

	// Case-insensitive / partial match if supported
	_, ok = FilterSite(sites, "nonexistent")
	if ok {
		t.Error("should not find nonexistent site")
	}
}

func TestLoadSitesInvalidPath(t *testing.T) {
	_, err := LoadSites("/nonexistent/sites.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadSitesInvalidYAML(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "bad.yaml")
	os.WriteFile(tmp, []byte("not: [valid: yaml: {{"), 0644)
	_, err := LoadSites(tmp)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
